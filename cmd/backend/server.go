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
	"github.com/jonathangjertsen/bino/sql"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Server struct {
	Conn          *pgxpool.Pool
	Queries       *sql.Queries
	Cookies       *sessions.CookieStore
	OAuthConfig   *oauth2.Config
	TokenVerifier *oidc.IDTokenVerifier
}

func (s *Server) Transaction(ctx context.Context, f func(ctx context.Context, q *sql.Queries) error) error {
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

func startServer(ctx context.Context, conn *pgxpool.Pool, queries *sql.Queries) error {
	sessionKey, err := os.ReadFile("secret/session_key")
	if err != nil {
		return err
	}

	c, err := loadCreds("secret/oauth.json")
	if err != nil {
		return err
	}

	issuer := "https://accounts.google.com"
	provider, err := oidc.NewProvider(ctx, issuer)
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
			Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
		},
		TokenVerifier: provider.Verifier(&oidc.Config{
			ClientID: c.Web.ClientID,
		}),
	}

	mux := http.NewServeMux()

	// Home page
	mux.HandleFunc("/{$}", server.requireLogin(server.dashboardHandler))

	// Auth
	mux.HandleFunc("/login", server.loginHandler)
	mux.HandleFunc("/logout", server.logoutHandler)
	mux.HandleFunc("/oauth2/callback", server.callbackHandler)

	// User ajax
	mux.HandleFunc("POST /language", server.requireLogin(server.postLanguageHandler))

	// Pages
	mux.HandleFunc("GET /species", server.requireLogin(server.getSpeciesHandler))
	mux.HandleFunc("GET /status", server.requireLogin(server.getStatusHandler))
	mux.HandleFunc("GET /event", server.requireLogin(server.getEventHandler))
	mux.HandleFunc("GET /tag", server.requireLogin(server.getTagHandler))
	mux.HandleFunc("GET /admin", server.requireLogin(server.adminRootHandler))
	mux.HandleFunc("GET /homes", server.requireLogin(server.getHomesHandler))

	// Admin AJAX
	mux.HandleFunc("POST /species", server.requireLogin(server.postSpeciesHandler))
	mux.HandleFunc("PUT /species", server.requireLogin(server.putSpeciesHandler))
	mux.HandleFunc("POST /status", server.requireLogin(server.postStatusHandler))
	mux.HandleFunc("PUT /status", server.requireLogin(server.putStatusHandler))
	mux.HandleFunc("POST /event", server.requireLogin(server.postEventHandler))
	mux.HandleFunc("PUT /event", server.requireLogin(server.putEventHandler))
	mux.HandleFunc("POST /tag", server.requireLogin(server.postTagHandler))
	mux.HandleFunc("PUT /tag", server.requireLogin(server.putTagHandler))

	// Admin forms
	mux.HandleFunc("POST /homes", server.requireLogin(server.postHomeHandler))

	// Available to all
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("/privacy", server.requireLogin(server.privacyHandler))

	mux.HandleFunc("/", server.requireLogin(server.fourOhFourHandler))

	go func() {
		handler := chain(mux, withRecover, withLogging)
		srv := &http.Server{
			Addr:              ":8080",
			Handler:           handler,
			ReadTimeout:       10 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
		}
		srv.ListenAndServe()
	}()

	return nil
}
