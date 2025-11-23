package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const OIDCURL = "https://accounts.google.com"

var ProfileScopes = []string{
	"openid",
	"email",
	"profile",
}

type Server struct {
	Conn          *pgxpool.Pool
	Queries       *Queries
	Cookies       *sessions.CookieStore
	OAuthConfig   *oauth2.Config
	TokenVerifier *oidc.IDTokenVerifier
	GDriveWorker  *GDriveWorker
	FileBackend   FileBackend
	Runtime       RuntimeInfo
	BuildKey      string
	Config        Config
}

type AuthConfig struct {
	SessionKeyLocation       string
	OAuthCredentialsLocation string
	ClientID                 string
	OAuthRedirectURI         string
}

type HTTPConfig struct {
	URL                      string
	ReadTimeoutSeconds       time.Duration
	ReadHeaderTimeoutSeconds time.Duration
	WriteTimeoutSeconds      time.Duration
	IdleTimeoutSeconds       time.Duration
	StaticDir                string
}

type RuntimeInfo struct {
	PublicIP    string
	TimeStarted time.Time
}

type Middleware = func(http.Handler) http.Handler

func (s *Server) Transaction(ctx context.Context, f func(ctx context.Context, q *Queries) error) error {
	tx, err := s.Conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("starting database transaction: %w", err)
	}
	q := s.Queries.WithTx(tx)
	err = f(ctx, q)
	if err == nil {
		err = tx.Commit(ctx)
	} else {
		tx.Rollback(ctx)
	}
	return err
}

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

func startServer(ctx context.Context, conn *pgxpool.Pool, queries *Queries, gdriveWorker *GDriveWorker, config Config, buildKey string) error {
	sessionKey, err := os.ReadFile(config.Auth.SessionKeyLocation)
	if err != nil {
		return err
	}

	c, err := loadCreds(config.Auth.OAuthCredentialsLocation)
	if err != nil {
		return err
	}

	provider, err := oidc.NewProvider(ctx, OIDCURL)
	if err != nil {
		return err
	}

	cookies := sessions.NewCookieStore(sessionKey)
	cookies.Options.HttpOnly = true
	cookies.Options.SameSite = http.SameSiteLaxMode
	cookies.Options.Secure = true

	redirectURI := config.Auth.OAuthRedirectURI
	if redirectURI == "" {
		redirectURI = c.Web.RedirectURIs[0]
	}
	fmt.Printf("using redirect URL: %s\n", redirectURI)

	server := &Server{
		Conn:         conn,
		Queries:      queries,
		Cookies:      cookies,
		GDriveWorker: gdriveWorker,
		OAuthConfig: &oauth2.Config{
			ClientID:     c.Web.ClientID,
			ClientSecret: c.Web.ClientSecret,
			RedirectURL:  redirectURI,
			Endpoint:     google.Endpoint,
			Scopes:       ProfileScopes,
		},
		TokenVerifier: provider.Verifier(&oidc.Config{
			ClientID: c.Web.ClientID,
		}),
		Runtime: RuntimeInfo{
			PublicIP:    fetchPublicIP(),
			TimeStarted: time.Now(),
		},
		FileBackend: NewLocalFileStorage(ctx, "file", "tmp"),
		BuildKey:    buildKey,
		Config:      config,
	}

	mux := http.NewServeMux()

	// Set up auth middlewares
	requiresLogin := []Middleware{server.requireLogin, withLogging, server.withFeedbackFromRedirects}

	requiresRehabber := slices.Clone(requiresLogin)
	requiresRehabber = append(requiresRehabber, server.requireAccessLevel(AccessLevelRehabber))

	requiresCoordinator := slices.Clone(requiresLogin)
	requiresCoordinator = append(requiresCoordinator, server.requireAccessLevel(AccessLevelCoordinator))

	requiresAdmin := slices.Clone(requiresLogin)
	requiresAdmin = append(requiresAdmin, server.requireAccessLevel(AccessLevelAdmin))

	loggedInHandler := func(handler http.HandlerFunc, cap Capability) http.Handler {
		requirements := slices.Clone(requiresLogin)
		requirements = append(requirements, server.requireCapability(cap))
		return chainf(handler, requirements...)
	}

	//// PUBLIC
	// Pages
	mux.Handle("GET /{$}", chainf(server.dashboardHandler, requiresLogin...))   // TODO should show something public for logged-out
	mux.Handle("GET /privacy", chainf(server.privacyHandler, requiresLogin...)) // TODO should be public
	mux.Handle("GET /access", chainf(server.accessHandler, requiresLogin...))
	// Static content
	staticDir := fmt.Sprintf("/static/%s/", buildKey)
	mux.Handle("GET "+staticDir, http.StripPrefix(staticDir, http.FileServer(http.Dir(config.HTTP.StaticDir))))
	// User content
	mux.Handle("GET /file/{id}/{filename}", chainf(server.fileHandler, requiresLogin...))

	//// LOGIN
	mux.Handle("GET /login", chainf(server.loginHandler))
	mux.Handle("POST /login", chainf(server.loginHandler))
	mux.Handle("GET /AuthLogOut", chainf(server.AuthLogOutHandler))
	mux.Handle("POST /AuthLogOut", chainf(server.AuthLogOutHandler))
	mux.Handle("GET /oauth2/callback", chainf(server.callbackHandler))
	mux.Handle("POST /oauth2/callback", chainf(server.callbackHandler))

	//// LOGGED-IN USER / REHABBER
	// Pages
	mux.Handle("GET /patient/{patient}", loggedInHandler(server.getPatientHandler, CapViewAllActivePatients))
	mux.Handle("GET /home/{home}", loggedInHandler(server.getHomeHandler, CapViewAllHomes))
	mux.Handle("GET /user/{user}", loggedInHandler(server.getUserHandler, CapViewAllHomes))
	mux.Handle("GET /former-patients", loggedInHandler(server.formerPatientsHandler, CapViewAllFormerPatients))
	mux.Handle("GET /calendar", loggedInHandler(server.calendarHandler, CapViewCalendar))
	mux.Handle("GET /import", loggedInHandler(server.getImportHandler, CapUseImportTool))
	mux.Handle("GET /search", loggedInHandler(server.searchHandler, CapSearch))
	mux.Handle("GET /search/live", loggedInHandler(server.searchLiveHandler, CapSearch))
	mux.Handle("GET /file", loggedInHandler(server.filePage, CapUploadFile))
	mux.Handle("GET /editor", loggedInHandler(server.editor, CapEditWiki))
	// Forms
	mux.Handle("POST /checkin", loggedInHandler(server.postCheckinHandler, CapCheckInPatient))
	mux.Handle("POST /privacy", loggedInHandler(server.postPrivacyHandler, CapSetOwnPreferences))
	mux.Handle("POST /patient/{patient}/move", loggedInHandler(server.movePatientHandler, CapManageOwnPatients))
	mux.Handle("POST /patient/{patient}/checkout", loggedInHandler(server.postCheckoutHandler, CapManageOwnPatients))
	mux.Handle("POST /patient/{patient}/set-name", loggedInHandler(server.postSetNameHandler, CapManageOwnPatients))
	mux.Handle("POST /patient/{patient}/create-journal", loggedInHandler(server.createJournalHandler, CapCreatePatientJournal))
	mux.Handle("POST /patient/{patient}/attach-journal", loggedInHandler(server.attachJournalHandler, CapManageOwnPatients))
	mux.Handle("POST /event/{event}/set-note", loggedInHandler(server.postEventSetNoteHandler, CapManageOwnPatients))
	mux.Handle("POST /home/{home}/set-capacity", loggedInHandler(server.setCapacityHandler, CapManageOwnHomes))
	mux.Handle("POST /home/{home}/add-unavailable", loggedInHandler(server.addHomeUnavailablePeriodHandler, CapManageOwnHomes))
	mux.Handle("POST /home/{home}/set-note", loggedInHandler(server.homeSetNoteHandler, CapManageOwnHomes))
	mux.Handle("POST /home/{home}/species/add", loggedInHandler(server.addPreferredSpeciesHandler, CapManageOwnHomes))
	mux.Handle("POST /home/{home}/species/delete/{species}", loggedInHandler(server.deletePreferredSpeciesHandler, CapManageOwnHomes))
	mux.Handle("POST /home/{home}/species/reorder", loggedInHandler(server.reorderSpeciesHandler, CapManageOwnHomes))
	mux.Handle("POST /period/{period}/delete", loggedInHandler(server.deleteHomeUnavailableHandler, CapManageOwnHomes))
	mux.Handle("POST /import", loggedInHandler(server.postImportHandler, CapUseImportTool))
	// Ajax
	mux.Handle("POST /language", loggedInHandler(server.postLanguageHandler, CapSetOwnPreferences))
	mux.Handle("POST /ajaxreorder", loggedInHandler(server.ajaxReorderHandler, CapManageOwnPatients))
	mux.Handle("POST /ajaxtransfer", loggedInHandler(server.ajaxTransferHandler, CapManageOwnPatients))
	mux.Handle("GET /calendar/away", loggedInHandler(server.ajaxCalendarAwayHandler, CapViewCalendar))
	mux.Handle("GET /calendar/patientevents", loggedInHandler(server.ajaxCalendarPatientEventsHandler, CapViewCalendar))
	mux.Handle("GET /import/validation", loggedInHandler(server.ajaxImportValidateHandler, CapViewCalendar))
	// Filepond
	mux.Handle("POST /file/filepond", loggedInHandler(server.filepondProcess, CapUploadFile))
	mux.Handle("DELETE /file/filepond", loggedInHandler(server.imageFilepondRevert, CapUploadFile))
	mux.Handle("GET /file/filepond/{id}", loggedInHandler(server.imageFilepondRestore, CapUploadFile))
	mux.Handle("POST /file/submit", loggedInHandler(server.filepondSubmit, CapUploadFile))
	mux.Handle("POST /file/{id}/delete", loggedInHandler(server.fileDelete, CapUploadFile))

	//// CONTENT MANAGEMENT
	// Pages
	mux.Handle("GET /species", loggedInHandler(server.getSpeciesHandler, CapManageSpecies))
	mux.Handle("GET /admin", loggedInHandler(server.adminRootHandler, CapViewAdminTools))
	mux.Handle("GET /homes", loggedInHandler(server.getHomesHandler, CapManageAllHomes))
	mux.Handle("GET /users", loggedInHandler(server.userAdminHandler, CapManageUsers))
	// Forms
	mux.Handle("POST /homes", loggedInHandler(server.postHomeHandler, CapManageAllHomes))
	mux.Handle("POST /homes/{home}/set-name", loggedInHandler(server.postHomeSetName, CapManageOwnHomes))
	// Ajax
	mux.Handle("POST /species", loggedInHandler(server.postSpeciesHandler, CapManageSpecies))
	mux.Handle("PUT /species", loggedInHandler(server.putSpeciesHandler, CapManageSpecies))

	//// ADMIN
	// Pages
	mux.Handle("GET /gdrive", loggedInHandler(server.getGDriveHandler, CapViewGDriveSettings))
	mux.Handle("GET /user/{user}/confirm-scrub", loggedInHandler(server.userConfirmScrubHandler, CapDeleteUsers))
	mux.Handle("GET /user/{user}/confirm-nuke", loggedInHandler(server.userConfirmNukeHandler, CapDeleteUsers))
	mux.Handle("GET /debug", loggedInHandler(server.debugHandler, CapDebug))
	// Forms
	mux.Handle("POST /user/{user}/scrub", loggedInHandler(server.userDoScrubHandler, CapDeleteUsers))
	mux.Handle("POST /user/{user}/nuke", loggedInHandler(server.userDoNukeHandler, CapDeleteUsers))
	mux.Handle("POST /gdrive/invite/{email}", loggedInHandler(server.gdriveInviteUserHandler, CapInviteToGDrive))
	mux.Handle("POST /invite", loggedInHandler(server.inviteHandler, CapInviteToBino))
	mux.Handle("POST /invite/{email}", loggedInHandler(server.inviteHandler, CapInviteToBino))
	mux.Handle("POST /invite/{id}/delete", loggedInHandler(server.inviteDeleteHandler, CapInviteToBino))

	//// FALLBACK
	// Pages
	mux.Handle("GET /", chainf(server.fourOhFourHandler, requiresLogin...))  // TODO: should be public
	mux.Handle("POST /", chainf(server.fourOhFourHandler, requiresLogin...)) // TODO: should be public

	go func() {
		handler := chain(mux, withRecover)
		srv := &http.Server{
			Addr:              config.HTTP.URL,
			Handler:           handler,
			ReadTimeout:       config.HTTP.ReadTimeoutSeconds * time.Second,
			ReadHeaderTimeout: config.HTTP.ReadHeaderTimeoutSeconds * time.Second,
			WriteTimeout:      config.HTTP.WriteTimeoutSeconds * time.Second,
			IdleTimeout:       config.HTTP.IdleTimeoutSeconds * time.Second,
		}
		if err := srv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	return nil
}

func (server *Server) getQueryValue(r *http.Request, field string) (string, error) {
	q := r.URL.Query()
	values, ok := q[field]
	if !ok {
		return "", fmt.Errorf("no such value: '%s'", field)
	}
	if len(values) != 1 {
		return "", fmt.Errorf("%d values named '%s", len(values), field)
	}
	return values[0], nil
}

func (server *Server) getFormIDs(r *http.Request, fields ...string) (map[string]int32, error) {
	strings, err := server.getFormValues(r, fields...)
	if err != nil {
		return nil, err
	}
	return stringsToIDs(strings)
}

func (server *Server) getFormValues(r *http.Request, fields ...string) (map[string]string, error) {
	return SliceToMapErr(fields, func(_ int, field string) (string, string, error) {
		v, err := server.getFormValue(r, field)
		return field, v, err
	})
}

func (server *Server) getOptionalFormValues(r *http.Request, fields ...string) map[string]string {
	return SliceToMap(fields, func(field string) (string, string) {
		v, _ := server.getFormValue(r, field)
		return field, v
	})
}

func (server *Server) getFormID(r *http.Request, field string) (int32, error) {
	vStr, err := server.getFormValue(r, field)
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseInt(vStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}

func (server *Server) getFormMultiValue(r *http.Request, field string) ([]string, error) {
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("parsing form: %w", err)
	}
	values, ok := r.Form[field]
	if !ok {
		return nil, fmt.Errorf("missing form value '%s'", field)
	}
	return values, nil
}

func (server *Server) getFormValue(r *http.Request, field string) (string, error) {
	values, err := server.getFormMultiValue(r, field)
	if err != nil {
		return "", err
	}
	if len(values) != 1 {
		return "", fmt.Errorf("expect 1 form value for '%s', got %d", field, len(values))
	}
	return values[0], nil
}

func (server *Server) getPathIDs(r *http.Request, fields ...string) (map[string]int32, error) {
	strings, err := server.getPathValues(r, fields...)
	if err != nil {
		return nil, err
	}
	return stringsToIDs(strings)
}

func (server *Server) getPathValues(r *http.Request, fields ...string) (map[string]string, error) {
	return SliceToMapErr(fields, func(_ int, field string) (string, string, error) {
		v, err := server.getPathValue(r, field)
		return field, v, err
	})
}

func (server *Server) getPathID(r *http.Request, field string) (int32, error) {
	vStr := r.PathValue(field)
	v, err := strconv.ParseInt(vStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}

func (server *Server) getPathValue(r *http.Request, field string) (string, error) {
	v := r.PathValue(field)
	var err error
	if v == "" {
		err = fmt.Errorf("no such path value: '%s'", field)
	}
	return v, err
}

func (server *Server) getCheckboxValue(r *http.Request, field string) bool {
	v, err := server.getFormValue(r, field)
	return err == nil && v == "on"
}
