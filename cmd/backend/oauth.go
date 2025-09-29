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
	verifier := provider.Verifier(&oidc.Config{
		ClientID: c.Web.ClientID,
	})
	oauthConfig := &oauth2.Config{
		ClientID:     c.Web.ClientID,
		ClientSecret: c.Web.ClientSecret,
		RedirectURL:  c.Web.RedirectURIs[0],
		Endpoint:     google.Endpoint,
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
	}

	store := sessions.NewCookieStore(sessionKey)
	store.Options.HttpOnly = true
	store.Options.SameSite = http.SameSiteLaxMode
	store.Options.Secure = true

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		loginHandler(w, r, store, oauthConfig)
	})
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		logoutHandler(w, r, store)
	})
	http.HandleFunc("/oauth2/callback", func(w http.ResponseWriter, r *http.Request) {
		callbackHandler(w, r, store, oauthConfig, verifier, queries)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rootHandler(w, r, store, queries)
	})

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

func loginHandler(w http.ResponseWriter, r *http.Request, store *sessions.CookieStore, oauthConfig *oauth2.Config) {
	state := randState()
	session, _ := store.Get(r, "auth")
	session.Values["state"] = state
	if err := session.Save(r, w); err != nil {
		fmt.Fprintf(os.Stderr, "saving cookie: %v", err)
	}
	http.Redirect(w, r, oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline), http.StatusFound)
}

func logoutHandler(w http.ResponseWriter, r *http.Request, store *sessions.CookieStore) {
	sess, _ := store.Get(r, "auth")
	sess.Options.MaxAge = -1
	_ = sess.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func callbackHandler(
	w http.ResponseWriter,
	r *http.Request,
	store *sessions.CookieStore,
	oauthConfig *oauth2.Config,
	verifier *oidc.IDTokenVerifier,
	queries *sql.Queries,
) {
	ctx := r.Context()
	sess, _ := store.Get(r, "auth")
	if r.URL.Query().Get("state") != sess.Values["state"] {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "exchange failed", http.StatusUnauthorized)
		return
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "no id_token", http.StatusUnauthorized)
		return
	}
	idToken, err := verifier.Verify(ctx, rawIDToken)
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
	userID, err := queries.UpsertUser(ctx, sql.UpsertUserParams{
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

func rootHandler(w http.ResponseWriter, r *http.Request, store *sessions.CookieStore, queries *sql.Queries) {
	user, err := authenticate(r, store, queries)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		http.Error(w, "authentication failed", getStatusCode(err))
		return
	}
	fmt.Fprintf(w, "user=%+v", user)
}

func authenticate(r *http.Request, store *sessions.CookieStore, queries *sql.Queries) (sql.Appuser, error) {
	ctx := r.Context()

	sess, _ := store.Get(r, "auth")
	uidIF, ok := sess.Values["user_id"]
	if !ok {
		return sql.Appuser{}, ErrUnauthorized
	}
	uid, ok := uidIF.(int32)
	if !ok {
		return sql.Appuser{}, fmt.Errorf("%w: uid is %T", ErrInternalServerError, uid)
	}

	user, err := queries.GetUser(ctx, uid)
	if err != nil {
		return sql.Appuser{}, fmt.Errorf("%w: database error", ErrInternalServerError)
	}

	return user, nil
}
