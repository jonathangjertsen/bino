package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"
)

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		LogR(
			r,
			"%s %s %s",
			r.Method,
			r.URL.Path,
			time.Since(start),
		)
		r.ParseForm()
		for k, v := range r.Form {
			LogR(
				r,
				"Form value: %s=%+v",
				k,
				v,
			)
		}
	})
}

func withRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				msg := fmt.Sprintf("runtime panic: %s\ntraceback: %s", rec, string(debug.Stack()))
				http.Error(w, msg, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (server *Server) withFeedbackFromRedirects(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.eatFeedbackCookie(w, r)
		next.ServeHTTP(w, r)
	})
}

func chain(h http.Handler, m ...func(http.Handler) http.Handler) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

func chainf(h http.HandlerFunc, m ...func(http.Handler) http.Handler) http.Handler {
	return chain(h, m...)
}
