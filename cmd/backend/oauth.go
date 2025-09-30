package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/coreos/go-oidc"
	"github.com/gorilla/sessions"
	"github.com/jonathangjertsen/bino/sql"
	"github.com/jonathangjertsen/bino/views"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

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

type googleCreds struct {
	Web struct {
		ClientID     string   `json:"client_id"`
		ClientSecret string   `json:"client_secret"`
		RedirectURIs []string `json:"redirect_uris"`
	} `json:"web"`
}

func loadCreds(path string) (googleCreds, error) {
	var c googleCreds
	f, err := os.Open(path)
	if err != nil {
		return c, err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&c)
	return c, err
}

func startServer(ctx context.Context, queries *sql.Queries) error {
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

	http.HandleFunc("/login", server.loginHandler)
	http.HandleFunc("/logout", server.logoutHandler)
	http.HandleFunc("/oauth2/callback", server.callbackHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", server.rootHandler)

	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	return nil
}

func randState() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func (server *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	state := randState()
	session, _ := server.Cookies.Get(r, "auth")
	session.Values["state"] = state
	if err := session.Save(r, w); err != nil {
		fmt.Fprintf(os.Stderr, "saving cookie: %v", err)
	}
	http.Redirect(w, r, server.OAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline), http.StatusFound)
}

func (server *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	sess, _ := server.Cookies.Get(r, "auth")
	sess.Options.MaxAge = -1
	_ = sess.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (server *Server) callbackHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()
	sess, _ := server.Cookies.Get(r, "auth")
	if r.URL.Query().Get("state") != sess.Values["state"] {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	token, err := server.OAuthConfig.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "exchange failed", http.StatusUnauthorized)
		return
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "no id_token", http.StatusUnauthorized)
		return
	}
	idToken, err := server.TokenVerifier.Verify(ctx, rawIDToken)
	if err != nil {
		http.Error(w, "verify failed", http.StatusUnauthorized)
	}
	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "claims failed", http.StatusUnauthorized)
		return
	}
	userID, err := server.Queries.UpsertUser(ctx, sql.UpsertUserParams{
		GoogleSub:   claims.Sub,
		Email:       claims.Email,
		DisplayName: claims.Name,
	})
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	sess.Values["user_id"] = userID
	sess.Values["email"] = claims.Email
	_ = sess.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (server *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	user, err := server.authenticate(w, r)
	if err != nil {
		return
	}

	userData := views.UserData{
		DisplayName: user.DisplayName,
		Email:       user.Email,
	}

	// TODO: placeholder
	species := []views.Species{
		{ID: 1, Name: "Feral pigeon"},
		{ID: 2, Name: "Wood pigeon"},
	}

	// TODO: placeholder
	labels := []views.Label{
		{ID: 1, Name: "Critical", Checked: false},
		{ID: 1, Name: "Disease", Checked: false},
		{ID: 1, Name: "Injury", Checked: false},
		{ID: 1, Name: "Force feeding", Checked: false},
		{ID: 1, Name: "Juvenile", Checked: false},
	}

	_ = views.DashboardPage(userData, species, labels).Render(r.Context(), w)
}

func (server *Server) authenticate(w http.ResponseWriter, r *http.Request) (sql.Appuser, error) {
	user, err := server.getUser(r)

	if err != nil {
		server.loginHandler(w, r)
	}

	return user, err
}

func (server *Server) getUser(r *http.Request) (sql.Appuser, error) {
	ctx := r.Context()

	sess, _ := server.Cookies.Get(r, "auth")
	uidIF, ok := sess.Values["user_id"]
	if !ok {
		return sql.Appuser{}, ErrUnauthorized
	}
	uid, ok := uidIF.(int32)
	if !ok {
		return sql.Appuser{}, fmt.Errorf("%w: uid is %T", ErrInternalServerError, uid)
	}

	user, err := server.Queries.GetUser(ctx, uid)
	if err != nil {
		return sql.Appuser{}, fmt.Errorf("%w: database error", ErrInternalServerError)
	}

	return user, nil
}

type Server struct {
	Queries       *sql.Queries
	Cookies       *sessions.CookieStore
	OAuthConfig   *oauth2.Config
	TokenVerifier *oidc.IDTokenVerifier
}
