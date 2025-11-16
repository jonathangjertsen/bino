package main

import (
	"context"
	"log"
	"net/http"
)

type PrivacyConfig struct {
	LogDeletionPolicy   int32
	RevokeConsentPolicy int32
}

func (server *Server) privacyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)
	_ = Privacy(commonData, server.Config.Privacy).Render(ctx, w)
}

func (server *Server) postPrivacyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	consent := server.getCheckboxValue(r, "logging-consent")

	var err error
	if consent {
		err = server.Queries.SetLoggingConsent(ctx, SetLoggingConsentParams{
			ID:     commonData.User.AppuserID,
			Period: server.Config.Privacy.RevokeConsentPolicy,
		})
	} else {
		err = server.Queries.RevokeLoggingConsent(ctx, commonData.User.AppuserID)
	}
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	server.redirectToReferer(w, r)
}

// Log from request if the user has given explicit concent
func LogR(r *http.Request, format string, args ...any) {
	LogCtx(r.Context(), format, args...)
}

// Log from context if the user has given explicit concent
func LogCtx(ctx context.Context, format string, args ...any) {
	commonData, err := LoadCommonData(ctx)
	if err != nil {
		return
	}
	commonData.Log(format, args...)
}

// Log if the user has given explicit concent
func (cd *CommonData) Log(format string, args ...any) {
	if !cd.User.LoggingConsent {
		return
	}
	log.Printf(format, args...)
}
