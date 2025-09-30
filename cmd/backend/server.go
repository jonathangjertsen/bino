package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/coreos/go-oidc"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

	http.HandleFunc("/login", server.loginHandler)
	http.HandleFunc("/logout", server.logoutHandler)
	http.HandleFunc("/oauth2/callback", server.callbackHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", server.rootHandler)
	http.HandleFunc("POST /species", server.postSpeciesHandler)
	http.HandleFunc("GET /species", server.getSpeciesHandler)
	http.HandleFunc("PUT /species", server.putSpeciesHandler)
	http.HandleFunc("POST /language", server.postLanguageHandler)

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

func (server *Server) postLanguageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	commonData, err := server.authenticate(w, r)
	if err != nil {
		// TODO
		return
	}

	lang, err := getSelectedLanguage(r.FormValue("language"), &commonData)
	if err == nil {
		err = server.Queries.SetUserLanguage(ctx, sql.SetUserLanguageParams{
			AppuserID:  commonData.User.AppuserID,
			LanguageID: lang,
		})
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	// TODO: should redirect back to where we came from
	http.Redirect(w, r, "/", http.StatusFound)
}

func getSelectedLanguage(langStr string, commonData *views.CommonData) (int32, error) {
	langID, err := strconv.Atoi(langStr)
	if err != nil {
		return 0, fmt.Errorf("invalid language ID: %w", langStr)
	}
	for _, lang := range commonData.Languages {
		if lang.ID == int32(langID) {
			return lang.ID, nil
		}
	}
	return 0, fmt.Errorf("unsupported language ID: %d", langID)
}

func (server *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	commonData, err := server.authenticate(w, r)
	if err != nil {
		return
	}

	rows, err := server.Queries.GetSpeciesWithLanguage(ctx, commonData.User.LanguageID)
	if err != nil {
		// TODO
		return
	}

	species := make([]views.Species, 0, len(rows))
	for _, row := range rows {
		species = append(species, views.Species{
			ID:   row.SpeciesID,
			Name: row.Name,
		})
	}

	// TODO: placeholder
	labels := []views.Label{
		{ID: 1, Name: "Critical", Checked: false},
		{ID: 1, Name: "Disease", Checked: false},
		{ID: 1, Name: "Injury", Checked: false},
		{ID: 1, Name: "Force feeding", Checked: false},
		{ID: 1, Name: "Juvenile", Checked: false},
	}

	_ = views.DashboardPage(commonData, species, labels).Render(r.Context(), w)
}

func (server *Server) postSpeciesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, err := server.authenticate(w, r)
	if err != nil {
		// TODO
		fmt.Println(err)
		return
	}

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		// TODO
		fmt.Println(err)
		return
	}
	var req struct {
		Latin     string
		Languages map[int32]string
	}
	if err := json.Unmarshal(bytes, &req); err != nil {
		// TODO
		fmt.Println(err)
		return
	}

	tx, err := server.Conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		// TODO
		fmt.Println(err)
		return
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	q := server.Queries.WithTx(tx)
	id, err := q.AddSpecies(ctx, req.Latin)
	if err != nil {
		// TODO
		fmt.Println(err)
		return
	}
	for langID, name := range req.Languages {
		if err := q.UpsertSpeciesLanguage(ctx, sql.UpsertSpeciesLanguageParams{
			SpeciesID:  id,
			LanguageID: langID,
			Name:       name,
		}); err != nil {
			// TODO
			fmt.Println(err)
			return
		}
	}
	_ = tx.Commit(ctx)

	w.WriteHeader(http.StatusOK)
}

func (server *Server) putSpeciesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, err := server.authenticate(w, r)
	if err != nil {
		// TODO
		fmt.Println(err)
		return
	}

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		// TODO
		fmt.Println(err)
		return
	}
	var req struct {
		ID        int32
		Latin     string
		Languages map[int32]string
	}
	if err := json.Unmarshal(bytes, &req); err != nil {
		// TODO
		fmt.Println(err)
		return
	}

	tx, err := server.Conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		// TODO
		fmt.Println(err)
		return
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	q := server.Queries.WithTx(tx)
	for langID, name := range req.Languages {
		if err := q.UpsertSpeciesLanguage(ctx, sql.UpsertSpeciesLanguageParams{
			SpeciesID:  req.ID,
			LanguageID: langID,
			Name:       name,
		}); err != nil {
			// TODO
			fmt.Println(err)
			return
		}
	}
	_ = tx.Commit(ctx)

	w.WriteHeader(http.StatusOK)
}

func (server *Server) getSpeciesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	commonData, err := server.authenticate(w, r)
	if err != nil {
		return
	}

	rows, err := server.Queries.GetSpecies(ctx)
	if err != nil {
		// TODO
		return
	}

	langRows, err := server.Queries.GetSpeciesLanguage(ctx)
	if err != nil {
		// TODO
		return
	}

	speciesView := make([]views.SpeciesLangs, 0, len(rows))
	iLangRows := 0
	for _, row := range rows {
		species := views.SpeciesLangs{
			ID:        row.ID,
			LatinName: row.ScientificName,
			Names:     map[int32]string{},
		}
		for {
			if iLangRows >= len(langRows) {
				break
			}
			langRow := langRows[iLangRows]
			if langRow.SpeciesID == row.ID {
				species.Names[langRow.LanguageID] = langRow.Name
				iLangRows++
			} else {
				break
			}
		}

		speciesView = append(speciesView, species)
	}

	_ = views.SpeciesPage(commonData, speciesView).Render(ctx, w)
}

func (server *Server) authenticate(w http.ResponseWriter, r *http.Request) (views.CommonData, error) {
	ctx := r.Context()

	user, err := server.getUser(r)

	if err != nil {
		server.loginHandler(w, r)
	}

	userData := views.UserData{
		AppuserID:   user.ID,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		LanguageID:  user.LanguageID,
	}

	languages, err := server.Queries.GetLanguages(ctx)
	if err != nil {
		return views.CommonData{}, fmt.Errorf("couldn't read languages: %w", err)
	}

	viewLanguages := make([]views.Language, 0, len(languages))
	for _, lang := range languages {
		viewLanguages = append(viewLanguages, views.Language{ID: lang.ID, Emoji: lang.ShortName, SelfName: lang.SelfName})
	}

	commonData := views.CommonData{
		User:      userData,
		Languages: viewLanguages,
	}

	return commonData, err
}

func (server *Server) getUser(r *http.Request) (sql.GetUserRow, error) {
	ctx := r.Context()

	sess, _ := server.Cookies.Get(r, "auth")
	uidIF, ok := sess.Values["user_id"]
	if !ok {
		return sql.GetUserRow{}, ErrUnauthorized
	}
	uid, ok := uidIF.(int32)
	if !ok {
		return sql.GetUserRow{}, fmt.Errorf("%w: uid is %T", ErrInternalServerError, uid)
	}

	user, err := server.Queries.GetUser(ctx, uid)
	if err != nil {
		return sql.GetUserRow{}, fmt.Errorf("%w: database error", ErrInternalServerError)
	}

	return user, nil
}

type Server struct {
	Conn          *pgxpool.Pool
	Queries       *sql.Queries
	Cookies       *sessions.CookieStore
	OAuthConfig   *oauth2.Config
	TokenVerifier *oidc.IDTokenVerifier
}
