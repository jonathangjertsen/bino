package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
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
			fmt.Println(err)
			return
		}

		ctx := WithCommonData(r.Context(), &commonData)
		r = r.WithContext(ctx)

		f.ServeHTTP(w, r)
	})
}

func (server *Server) authenticate(w http.ResponseWriter, r *http.Request) (CommonData, error) {
	ctx := r.Context()

	user, err := server.getUser(r, w)

	if err != nil {
		server.loginHandler(w, r)
		return CommonData{}, err
	}

	homes, err := server.Queries.GetHomesForUser(ctx, user.ID)
	var preferredHome Home
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
		HasGDriveAccess: user.HasGdriveAccess,
		Language:        GetLanguage(user.LanguageID),
		PreferredHome:   preferredHome,
		Homes:           homes,
		LoggingConsent:  loggingConsent,
	}

	commonData := CommonData{
		User:     userData,
		BuildKey: server.BuildKey,
	}

	return commonData, err
}

func (server *Server) getUser(r *http.Request, w http.ResponseWriter) (GetUserRow, error) {
	// Get the encrypted auth cookie
	ctx := r.Context()
	sess, _ := server.Cookies.Get(r, "auth")

	// OAuth token data must be valid
	tokenData, ok := sess.Values["oauth_token"].(string)
	if !ok {
		return GetUserRow{}, fmt.Errorf("no OAuth token in session")
	}
	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenData), &token); err != nil {
		return GetUserRow{}, fmt.Errorf("correpted OAuth token in session")
	}

	ts := oauth2.ReuseTokenSource(&token, server.OAuthConfig.TokenSource(ctx, &token))
	newTok, err := ts.Token()
	if err != nil {
		return GetUserRow{}, fmt.Errorf("unable to refresh token: %w", err)
	}
	if newTok.AccessToken != token.AccessToken || newTok.Expiry != token.Expiry {
		b, err := json.Marshal(newTok)
		if err != nil {
			return GetUserRow{}, fmt.Errorf("unable to marshal new token: %w", err)
		}
		sess.Values["oauth_token"] = string(b)
		_ = sess.Save(r, w)
	}

	// Look up session
	sessionIDIF, ok := sess.Values["session_id"]
	if !ok {
		return GetUserRow{}, ErrUnauthorized
	}
	sessionID, ok := sessionIDIF.(string)
	if !ok {
		return GetUserRow{}, fmt.Errorf("%w: session_id is %T", ErrInternalServerError, sessionID)
	}
	session, err := server.Queries.GetSession(ctx, sessionID)
	if err != nil {
		return GetUserRow{}, fmt.Errorf("%w: no such session %s", err, sessionID)
	}
	if time.Now().After(session.Expires.Time) {
		return GetUserRow{}, fmt.Errorf("session expired")
	}
	if err := server.Queries.UpdateSessionLastSeen(ctx, sessionID); err != nil {
		return GetUserRow{}, fmt.Errorf("updating session-last-seen failed")
	}

	// Look up user
	user, err := server.Queries.GetUser(ctx, session.AppuserID)
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
	http.Redirect(
		w,
		r,
		server.OAuthConfig.AuthCodeURL(
			state,
			oauth2.AccessTypeOffline,
			oauth2.SetAuthURLParam("prompt", "consent"),
		),
		http.StatusFound,
	)
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

	// Get OAuth token
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

	// Invalid token
	if !token.Valid() {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}
	if token.Expiry.IsZero() {
		http.Error(w, "no expiry set for token", http.StatusBadRequest)
		return
	}

	// Store the OAuth token for Drive API access
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		http.Error(w, "token serialization failed", http.StatusInternalServerError)
		return
	}

	// Get the ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "no id_token", http.StatusUnauthorized)
		return
	}
	idToken, err := server.TokenVerifier.Verify(ctx, rawIDToken)
	if err != nil {
		http.Error(w, "verify failed", http.StatusUnauthorized)
	}

	// Get stuff associated with the ID token
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

	// User must exist or be invited
	userID, err := server.Queries.GetUserIDByEmail(ctx, claims.Email)
	if err == nil {
		// User exists; update personal data
		if err := server.Queries.UpdateUser(ctx, UpdateUserParams{
			ID:          userID,
			GoogleSub:   claims.Sub,
			Email:       claims.Email,
			DisplayName: claims.Name,
			AvatarUrl:   pgtype.Text{String: claims.Picture, Valid: claims.Picture != ""},
		}); err != nil {
			http.Error(w, "creating user failed", http.StatusInternalServerError)
			return
		}
	} else if invitation, err := server.Queries.GetInvitation(ctx, pgtype.Text{String: claims.Email, Valid: true}); err == nil {
		// User has been invited; create user
		if server.Transaction(ctx, func(ctx context.Context, q *Queries) error {
			if createdUserID, err := server.Queries.CreateUser(ctx, CreateUserParams{
				DisplayName: claims.Name,
				Email:       claims.Email,
				GoogleSub:   claims.Sub,
				AvatarUrl:   pgtype.Text{String: claims.Picture, Valid: claims.Picture != ""},
			}); err != nil {
				return err
			} else {
				userID = createdUserID
			}
			if err := server.Queries.DeleteInvitation(ctx, invitation); err != nil {
				return err
			}
			return nil
		}); err != nil {
			http.Error(w, "creating user failed", http.StatusInternalServerError)
			return
		}
	} else {
		fmt.Printf("no inv for %s\n", claims.Email)
		http.Error(w, "user doesn't exist and is not invited", http.StatusUnauthorized)
		return
	}

	// Create session
	sessionID := rand.Text()
	if err := server.Queries.InsertSession(ctx, InsertSessionParams{
		ID:        sessionID,
		AppuserID: userID,
		Expires:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
	}); err != nil {
		http.Error(w, "creating session failed", http.StatusInternalServerError)
		return
	}

	// Store to cookie
	sess.Values["oauth_token"] = string(tokenJSON)
	sess.Values["session_id"] = sessionID
	sess.Values["email"] = claims.Email
	_ = sess.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (server *Server) getTokenFromSession(sess *sessions.Session) (*oauth2.Token, error) {
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

func (server *Server) getTokenFromRequest(r *http.Request) (*oauth2.Token, error) {
	sess, _ := server.Cookies.Get(r, "auth")
	return server.getTokenFromSession(sess)
}
