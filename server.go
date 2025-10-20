package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const OIDCURL = "https://accounts.google.com"

var ProfileScopes = []string{
	"openid",
	"email",
	"profile",
}

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
	ClientID                 string
	DriveBase                string
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

	provider, err := oidc.NewProvider(ctx, OIDCURL)
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
			Scopes:       append(ProfileScopes, GoogleDriveScopes...),
		},
		TokenVerifier: provider.Verifier(&oidc.Config{
			ClientID: c.Web.ClientID,
		}),
		BuildKey: buildKey,
		Config:   config,
	}

	mux := http.NewServeMux()

	loggedInChain := []func(http.Handler) http.Handler{server.requireLogin, withLogging, server.withFeedbackFromRedirects}

	// Home page
	mux.Handle("GET /{$}", chainf(server.dashboardHandler, loggedInChain...))

	// Auth
	mux.Handle("GET /login", chainf(server.loginHandler))
	mux.Handle("POST /login", chainf(server.loginHandler))
	mux.Handle("GET /AuthLogOut", chainf(server.AuthLogOutHandler))
	mux.Handle("POST /AuthLogOut", chainf(server.AuthLogOutHandler))
	mux.Handle("GET /oauth2/callback", chainf(server.callbackHandler))
	mux.Handle("POST /oauth2/callback", chainf(server.callbackHandler))

	// User ajax
	mux.Handle("POST /language", chainf(server.postLanguageHandler, loggedInChain...))

	// Pages
	mux.Handle("GET /species", chainf(server.getSpeciesHandler, loggedInChain...))
	mux.Handle("GET /tag", chainf(server.getTagHandler, loggedInChain...))
	mux.Handle("GET /admin", chainf(server.adminRootHandler, loggedInChain...))
	mux.Handle("GET /homes", chainf(server.getHomesHandler, loggedInChain...))
	mux.Handle("GET /privacy", chainf(server.privacyHandler, loggedInChain...))
	mux.Handle("GET /gdrive", chainf(server.getGDriveHandler, loggedInChain...))
	mux.Handle("POST /gdrive/set-base-folder/{id}", chainf(server.setGDriveBaseFolderHandler, loggedInChain...))
	mux.Handle("POST /gdrive/find-template", chainf(server.gdriveFindTemplate, loggedInChain...))
	mux.Handle("POST /gdrive/set-template/{id}", chainf(server.gdriveSetTemplate, loggedInChain...))
	mux.Handle("POST /gdrive/invite/{email}", chainf(server.gdriveInviteUserHandler, loggedInChain...))
	mux.Handle("GET /patient/{patient}", chainf(server.getPatientHandler, loggedInChain...))
	mux.Handle("GET /home/{home}", chainf(server.getHomeHandler, loggedInChain...))
	mux.Handle("GET /user/{user}", chainf(server.getUserHandler, loggedInChain...))
	mux.Handle("GET /user/{user}/confirm-scrub", chainf(server.userConfirmScrubHandler, loggedInChain...))
	mux.Handle("GET /user/{user}/confirm-nuke", chainf(server.userConfirmNukeHandler, loggedInChain...))
	mux.Handle("POST /user/{user}/scrub", chainf(server.userDoScrubHandler, loggedInChain...))
	mux.Handle("POST /user/{user}/nuke", chainf(server.userDoNukeHandler, loggedInChain...))
	mux.Handle("GET /former-patients", chainf(server.formerPatientsHandler, loggedInChain...))
	mux.Handle("GET /users", chainf(server.userAdminHandler, loggedInChain...))

	// Admin AJAX
	mux.Handle("POST /species", chainf(server.postSpeciesHandler, loggedInChain...))
	mux.Handle("PUT /species", chainf(server.putSpeciesHandler, loggedInChain...))
	mux.Handle("POST /tag", chainf(server.postTagHandler, loggedInChain...))
	mux.Handle("PUT /tag", chainf(server.putTagHandler, loggedInChain...))

	// Dashboard ajax
	mux.Handle("DELETE /patient/{patient}/tag/{tag}", chainf(server.deletePatientTagHandler, loggedInChain...))
	mux.Handle("POST /patient/{patient}/tag/{tag}", chainf(server.createPatientTagHandler, loggedInChain...))

	// Forms
	mux.Handle("POST /checkin", chainf(server.postCheckinHandler, loggedInChain...))
	mux.Handle("POST /homes", chainf(server.postHomeHandler, loggedInChain...))
	mux.Handle("POST /homes/{home}/set-name", chainf(server.postHomeSetName, loggedInChain...))
	mux.Handle("POST /privacy", chainf(server.postPrivacyHandler, loggedInChain...))
	mux.Handle("POST /patient/{patient}/move", chainf(server.movePatientHandler, loggedInChain...))
	mux.Handle("POST /patient/{patient}/checkout", chainf(server.postCheckoutHandler, loggedInChain...))
	mux.Handle("POST /patient/{patient}/set-name", chainf(server.postSetNameHandler, loggedInChain...))
	mux.Handle("POST /event/{event}/set-note", chainf(server.postEventSetNoteHandler, loggedInChain...))

	// Available to all
	staticDir := fmt.Sprintf("/static/%s/", buildKey)
	mux.Handle(
		"GET "+staticDir,
		http.StripPrefix(staticDir, http.FileServer(http.Dir("static"))),
	)

	mux.Handle("GET /", chainf(server.fourOhFourHandler, loggedInChain...))
	mux.Handle("POST /", chainf(server.fourOhFourHandler, loggedInChain...))

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

func (server *Server) getFormIDs(r *http.Request, fields ...string) (map[string]int32, error) {
	strings, err := server.getFormValues(r, fields...)
	if err != nil {
		return nil, err
	}
	return stringsToIDs(strings)
}

func (server *Server) getFormValues(r *http.Request, fields ...string) (map[string]string, error) {
	return SliceToMapErr(fields, func(_ int, field string) (string, string, error) {
		v, err := server.getFormValue(r, field)
		return field, v, err
	})
}

func (server *Server) getFormID(r *http.Request, field string) (int32, error) {
	vStr, err := server.getFormValue(r, field)
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseInt(vStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}

func (server *Server) getFormValue(r *http.Request, field string) (string, error) {
	if err := r.ParseForm(); err != nil {
		return "", fmt.Errorf("parsing form: %w", err)
	}
	values, ok := r.Form[field]
	if !ok {
		return "", fmt.Errorf("missing form value '%s'", field)
	}
	if len(values) != 1 {
		return "", fmt.Errorf("expect 1 form value for '%s', got %d", field, len(values))
	}
	return values[0], nil
}

func (server *Server) getPathIDs(r *http.Request, fields ...string) (map[string]int32, error) {
	strings, err := server.getPathValues(r, fields...)
	if err != nil {
		return nil, err
	}
	return stringsToIDs(strings)
}

func (server *Server) getPathValues(r *http.Request, fields ...string) (map[string]string, error) {
	return SliceToMapErr(fields, func(_ int, field string) (string, string, error) {
		v, err := server.getPathValue(r, field)
		return field, v, err
	})
}

func (server *Server) getPathID(r *http.Request, field string) (int32, error) {
	vStr := r.PathValue(field)
	v, err := strconv.ParseInt(vStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}
