package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Server struct {
	Conn          *pgxpool.Pool
	Queries       *Queries
	Cookies       *sessions.CookieStore
	OAuthConfig   *oauth2.Config
	TokenVerifier *oidc.IDTokenVerifier
	BuildKey      string
	Config        Config
}

type AuthConfig struct {
	SessionKeyLocation       string
	OAuthCredentialsLocation string
	OIDCURL                  string
	OIDCScopes               []string
}

type HTTPConfig struct {
	URL                      string
	ReadTimeoutSeconds       time.Duration
	ReadHeaderTimeoutSeconds time.Duration
	WriteTimeoutSeconds      time.Duration
	IdleTimeoutSeconds       time.Duration
}

func (s *Server) Transaction(ctx context.Context, f func(ctx context.Context, q *Queries) error) error {
	tx, err := s.Conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("starting database transaction: %w", err)
	}
	q := s.Queries.WithTx(tx)
	err = f(ctx, q)
	if err == nil {
		err = tx.Commit(ctx)
	} else {
		tx.Rollback(ctx)
	}
	return err
}

var ErrUnauthorized = errors.New("unauthorized")
var ErrInternalServerError = errors.New("internal server error")

func getStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if errors.Is(err, ErrUnauthorized) {
		return http.StatusUnauthorized
	}
	return http.StatusInternalServerError
}

func startServer(ctx context.Context, conn *pgxpool.Pool, queries *Queries, config Config, buildKey string) error {
	sessionKey, err := os.ReadFile(config.Auth.SessionKeyLocation)
	if err != nil {
		return err
	}

	c, err := loadCreds(config.Auth.OAuthCredentialsLocation)
	if err != nil {
		return err
	}

	provider, err := oidc.NewProvider(ctx, config.Auth.OIDCURL)
	if err != nil {
		return err
	}

	cookies := sessions.NewCookieStore(sessionKey)
	cookies.Options.HttpOnly = true
	cookies.Options.SameSite = http.SameSiteLaxMode
	cookies.Options.Secure = true

	server := &Server{
		Conn:    conn,
		Queries: queries,
		Cookies: cookies,
		OAuthConfig: &oauth2.Config{
			ClientID:     c.Web.ClientID,
			ClientSecret: c.Web.ClientSecret,
			RedirectURL:  c.Web.RedirectURIs[0],
			Endpoint:     google.Endpoint,
			Scopes:       config.Auth.OIDCScopes,
		},
		TokenVerifier: provider.Verifier(&oidc.Config{
			ClientID: c.Web.ClientID,
		}),
		BuildKey: buildKey,
		Config:   config,
	}

	mux := http.NewServeMux()

	loggedInChain := []func(http.Handler) http.Handler{server.requireLogin, withLogging}

	// Home page
	mux.Handle("GET /{$}", chainf(server.dashboardHandler, loggedInChain...))

	// Auth
	mux.Handle("GET /login", chainf(server.loginHandler))
	mux.Handle("POST /login", chainf(server.loginHandler))
	mux.Handle("GET /logout", chainf(server.logoutHandler))
	mux.Handle("POST /logout", chainf(server.logoutHandler))
	mux.Handle("GET /oauth2/callback", chainf(server.callbackHandler))
	mux.Handle("POST /oauth2/callback", chainf(server.callbackHandler))

	// User ajax
	mux.Handle("POST /language", chainf(server.postLanguageHandler, loggedInChain...))

	// Pages
	mux.Handle("GET /species", chainf(server.getSpeciesHandler, loggedInChain...))
	mux.Handle("GET /tag", chainf(server.getTagHandler, loggedInChain...))
	mux.Handle("GET /admin", chainf(server.adminRootHandler, loggedInChain...))
	mux.Handle("GET /homes", chainf(server.getHomesHandler, loggedInChain...))

	// Admin AJAX
	mux.Handle("POST /species", chainf(server.postSpeciesHandler, loggedInChain...))
	mux.Handle("PUT /species", chainf(server.putSpeciesHandler, loggedInChain...))
	mux.Handle("POST /tag", chainf(server.postTagHandler, loggedInChain...))
	mux.Handle("PUT /tag", chainf(server.putTagHandler, loggedInChain...))

	// Forms
	mux.Handle("POST /", chainf(server.postDashboardHandler, loggedInChain...))
	mux.Handle("POST /homes", chainf(server.postHomeHandler, loggedInChain...))

	// Available to all
	staticDir := fmt.Sprintf("/static/%s/", buildKey)
	mux.Handle(
		"GET "+staticDir,
		http.StripPrefix(staticDir, http.FileServer(http.Dir("static"))),
	)

	mux.Handle("GET /privacy", chainf(server.privacyHandler, loggedInChain...))
	mux.Handle("POST /privacy", chainf(server.postPrivacyHandler, loggedInChain...))

	mux.Handle("GET /", chainf(server.fourOhFourHandler, loggedInChain...))

	go func() {
		handler := chain(mux, withRecover)
		srv := &http.Server{
			Addr:              config.HTTP.URL,
			Handler:           handler,
			ReadTimeout:       config.HTTP.ReadTimeoutSeconds * time.Second,
			ReadHeaderTimeout: config.HTTP.ReadHeaderTimeoutSeconds * time.Second,
			WriteTimeout:      config.HTTP.WriteTimeoutSeconds * time.Second,
			IdleTimeout:       config.HTTP.IdleTimeoutSeconds * time.Second,
		}
		srv.ListenAndServe()
	}()

	return nil
}
