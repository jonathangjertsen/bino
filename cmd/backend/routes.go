package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/jonathangjertsen/bino/sql"
	"github.com/jonathangjertsen/bino/views"
)

func (server *Server) adminRootHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	ctx := r.Context()
	_ = views.AdminRootPage(commonData).Render(ctx, w)
}

func (server *Server) postLanguageHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	ctx := r.Context()

	lang, err := getSelectedLanguage(r.FormValue("language"), commonData)
	if err == nil {
		err = server.Queries.SetUserLanguage(ctx, sql.SetUserLanguageParams{
			AppuserID:  commonData.User.AppuserID,
			LanguageID: lang,
		})
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusFound)
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

func (server *Server) renderError(w http.ResponseWriter, r *http.Request, commonData *views.CommonData, err error) {
	ctx := r.Context()
	_ = views.ErrorPage(commonData, err).Render(ctx, w)
	logError(r, err)
}

func ajaxError(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	logError(r, err)
	w.WriteHeader(statusCode)
}

func logError(r *http.Request, err error) {
	log.Printf("%s %s ERROR: %v", r.Method, r.URL.Path, err)
}

func (server *Server) fourOhFourHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	server.renderError(w, r, commonData, fmt.Errorf("not found: %s", r.RequestURI))
}

func (server *Server) privacyHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	_ = views.Privacy(commonData, views.PrivacyConfig{LogDeletionPolicy: 3, RevokeConsentPolicy: 3}).Render(r.Context(), w) // TODO
}

func jsonHandler[T any](
	server *Server,
	w http.ResponseWriter,
	r *http.Request,
	f func(q *sql.Queries, req T) error,
) {
	ctx := r.Context()

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		ajaxError(w, r, err, http.StatusBadRequest)
		return
	}
	var recv T
	if err := json.Unmarshal(bytes, &recv); err != nil {
		ajaxError(w, r, err, http.StatusBadRequest)
		return
	}
	if err := server.Transaction(ctx, func(ctx context.Context, q *sql.Queries) error {
		return f(q, recv)
	}); err != nil {
		ajaxError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
