package main

import (
	"log"
	"net/http"

	"github.com/jonathangjertsen/bino/sql"
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
		err = server.Queries.SetLoggingConsent(ctx, sql.SetLoggingConsentParams{
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

	http.Redirect(w, r, "/privacy", http.StatusFound)
}

// Log from request if the user has given explicit concent
func LogR(r *http.Request, format string, args ...any) {
	commonData, err := LoadCommonData(r.Context())
	if err != nil {
		return
	}
	if !commonData.User.LoggingConsent {
		return
	}
	log.Printf(format, args...)
}
