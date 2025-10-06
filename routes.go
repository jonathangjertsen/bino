package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func (server *Server) adminRootHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	_ = AdminRootPage(commonData).Render(ctx, w)
}

func (server *Server) postLanguageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	lang, err := ParseLanguageID(r.FormValue("language"))
	if err == nil {
		err = server.Queries.SetUserLanguage(ctx, SetUserLanguageParams{
			AppuserID:  commonData.User.AppuserID,
			LanguageID: int32(lang),
		})
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	server.redirectToReferer(w, r)
}

func (server *Server) redirectToReferer(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusFound)
}

func (server *Server) renderError(w http.ResponseWriter, r *http.Request, commonData *CommonData, err error) {
	ctx := r.Context()
	w.WriteHeader(http.StatusInternalServerError)
	_ = ErrorPage(commonData, err).Render(ctx, w)
	logError(r, err)
}

func (server *Server) render404(w http.ResponseWriter, r *http.Request, commonData *CommonData, err error) {
	ctx := r.Context()
	w.WriteHeader(http.StatusNotFound)
	_ = NotFoundPage(commonData, err.Error()).Render(ctx, w)
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
	server.render404(w, r, commonData, fmt.Errorf("not found: %s %s", r.Method, r.RequestURI))
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
