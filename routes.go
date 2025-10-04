package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

func (server *Server) adminRootHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	_ = AdminRootPage(commonData).Render(ctx, w)
}

func (server *Server) postLanguageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	lang, err := getSelectedLanguage(r.FormValue("language"), commonData)
	if err == nil {
		err = server.Queries.SetUserLanguage(ctx, SetUserLanguageParams{
			AppuserID:  commonData.User.AppuserID,
			LanguageID: lang,
		})
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusFound)
}

func getSelectedLanguage(langStr string, commonData *CommonData) (int32, error) {
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

func (server *Server) renderError(w http.ResponseWriter, r *http.Request, commonData *CommonData, err error) {
	ctx := r.Context()
	_ = ErrorPage(commonData, err).Render(ctx, w)
	logError(r, err)
}

func ajaxError(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	logError(r, err)
	w.WriteHeader(statusCode)
}

func logError(r *http.Request, err error) {
	LogR(r, "%s %s ERROR: %v", r.Method, r.URL.Path, err)
}

func (server *Server) fourOhFourHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)
	server.renderError(w, r, commonData, fmt.Errorf("not found: %s", r.RequestURI))
}

func jsonHandler[T any](
	server *Server,
	w http.ResponseWriter,
	r *http.Request,
	f func(q *Queries, req T) error,
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
	if err := server.Transaction(ctx, func(ctx context.Context, q *Queries) error {
		return f(q, recv)
	}); err != nil {
		ajaxError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
