package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/oauth2"
)

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

func (server *Server) requireLogin(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		commonData, err := server.authenticate(w, r)
		if err != nil {
			return
		}

		ctx := WithCommonData(r.Context(), &commonData)
		r = r.WithContext(ctx)

		f.ServeHTTP(w, r)
	})
}

func (server *Server) authenticate(w http.ResponseWriter, r *http.Request) (CommonData, error) {
	ctx := r.Context()

	user, err := server.getUser(r)

	if err != nil {
		server.loginHandler(w, r)
		return CommonData{}, err
	}

	homes, err := server.Queries.GetHomesForUser(ctx, user.ID)
	preferredHome := int32(0)
	if len(homes) > 0 {
		preferredHome = homes[0]
	}

	loggingConsent := user.LoggingConsent.Valid && user.LoggingConsent.Time.After(time.Now())

	userData := UserData{
		AppuserID:       user.ID,
		DisplayName:     user.DisplayName,
		Email:           user.Email,
		AvatarURL:       user.AvatarUrl.String,
		HasAvatarURL:    user.AvatarUrl.Valid,
		Language:        GetLanguage(user.LanguageID),
		PreferredHomeID: preferredHome,
		Homes:           homes,
		LoggingConsent:  loggingConsent,
	}

	commonData := CommonData{
		User:     userData,
		BuildKey: server.BuildKey,
	}

	return commonData, err
}

func (server *Server) getUser(r *http.Request) (GetUserRow, error) {
	ctx := r.Context()

	sess, _ := server.Cookies.Get(r, "auth")
	uidIF, ok := sess.Values["user_id"]
	if !ok {
		return GetUserRow{}, ErrUnauthorized
	}
	uid, ok := uidIF.(int32)
	if !ok {
		return GetUserRow{}, fmt.Errorf("%w: uid is %T", ErrInternalServerError, uid)
	}

	user, err := server.Queries.GetUser(ctx, uid)
	if err != nil {
		return GetUserRow{}, fmt.Errorf("%w: database error", ErrInternalServerError)
	}

	return user, nil
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

func (server *Server) AuthLogOutHandler(w http.ResponseWriter, r *http.Request) {
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

	// Store the OAuth token for Drive API access
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		http.Error(w, "token serialization failed", http.StatusInternalServerError)
		return
	}
	sess.Values["oauth_token"] = string(tokenJSON)

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
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "claims failed", http.StatusUnauthorized)
		return
	}
	userID, err := server.Queries.UpsertUser(ctx, UpsertUserParams{
		GoogleSub:   claims.Sub,
		Email:       claims.Email,
		DisplayName: claims.Name,
		AvatarUrl:   pgtype.Text{String: claims.Picture, Valid: claims.Picture != ""},
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

func (server *Server) getTokenFromSession(r *http.Request) (*oauth2.Token, error) {
	sess, _ := server.Cookies.Get(r, "auth")
	tokenData, ok := sess.Values["oauth_token"].(string)
	if !ok {
		return nil, fmt.Errorf("no oauth token in session")
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenData), &token); err != nil {
		return nil, err
	}

	return &token, nil
}
