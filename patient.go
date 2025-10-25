package main

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

var journalRegex = regexp.MustCompile(`(https:\/\/docs\.google\.com\/document\/d\/[^\/?#\n]+)`)

type PatientPageView struct {
	Patient PatientView
	Home    *HomeView
	Tags    []GetTagsWithLanguageCheckinRow
	Events  []EventView
	Homes   []HomeView
}

type EventView struct {
	Row     GetEventsForPatientRow
	TimeAbs string
	TimeRel string
	Home    HomeView
	User    UserView
}

func (ev *EventView) SetNoteURL() string {
	return fmt.Sprintf("/event/%d/set-note", ev.Row.ID)
}

func (server *Server) getPatientHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	id, err := server.getPathID(r, "patient")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	patientData, err := server.Queries.GetPatientWithSpecies(ctx, GetPatientWithSpeciesParams{
		ID:         id,
		LanguageID: commonData.Lang32(),
	})
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	var home *HomeView
	if patientData.CurrHomeID.Valid {
		homeResult, err := server.Queries.GetHome(ctx, patientData.CurrHomeID.Int32)
		if err != nil {
			server.renderError(w, r, commonData, err)
			return
		}
		home = &HomeView{
			Home: homeResult,
		}
	}

	homes, err := server.Queries.GetHomes(ctx)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	availableTags, err := server.Queries.GetTagsWithLanguageCheckin(ctx, commonData.Lang32())
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	patientTags, err := server.Queries.GetTagsForPatient(ctx, GetTagsForPatientParams{
		PatientID:  patientData.ID,
		LanguageID: commonData.Lang32(),
	})
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	eventData, err := server.Queries.GetEventsForPatient(ctx, patientData.ID)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	events := SliceToSlice(eventData, func(r GetEventsForPatientRow) EventView {
		return EventView{
			Row:     r,
			TimeRel: commonData.User.Language.FormatTimeRel(r.Time.Time),
			TimeAbs: commonData.User.Language.FormatTimeAbs(r.Time.Time),
			User: UserView{
				ID:           r.AppuserID,
				Name:         r.UserName,
				AvatarURL:    r.AvatarUrl.String,
				HasAvatarURL: r.AvatarUrl.Valid,
				Email:        "",
			},
			Home: HomeView{
				Home: Home{
					ID:   r.HomeID,
					Name: r.HomeName,
				},
			},
		}
	})

	PatientPage(ctx, commonData, PatientPageView{
		Patient: PatientView{
			ID:         patientData.ID,
			Status:     patientData.Status,
			Name:       patientData.Name,
			Species:    patientData.SpeciesName,
			JournalURL: patientData.JournalUrl.String,
			Tags: SliceToSlice(patientTags, func(in GetTagsForPatientRow) TagView {
				return TagView{
					ID:        in.TagID,
					Name:      in.Name,
					PatientID: patientData.ID,
				}
			}),
		},
		Home: home,
		Homes: SliceToSlice(homes, func(home Home) HomeView {
			return HomeView{Home: home}
		}),
		Tags:   availableTags,
		Events: events,
	}, server).Render(ctx, w)
}

func (server *Server) createJournalHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	patient, err := server.getPathID(r, "patient")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	patientData, err := server.Queries.GetPatientWithSpecies(ctx, GetPatientWithSpeciesParams{
		ID:         patient,
		LanguageID: int32(server.Config.SystemLanguage),
	})
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	if patientData.JournalUrl.Valid {
		commonData.Warning(commonData.User.Language.TODO("journal URL already exists"), nil)
		server.redirectToReferer(w, r)
		return
	}

	created, err := server.Queries.GetFirstEventOfTypeForPatient(ctx, GetFirstEventOfTypeForPatientParams{
		PatientID: patient,
		EventID:   int32(EventRegistered),
	})
	if err != nil || !created.Valid {
		LogR(r, "vad creation date, using current time. Err=%v", err)
		created = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	}

	item, err := server.GDriveWorker.CreateJournal(GDriveTemplateVars{
		Time:    created.Time,
		Name:    patientData.Name,
		Species: patientData.SpeciesName,
		BinoURL: server.Config.BinoURLForPatient(patient),
	})
	if err != nil {
		commonData.Error(commonData.User.Language.TODO("failed to create"), err)
		server.redirectToReferer(w, r)
		return
	}

	if err := server.Queries.SetPatientJournal(ctx, SetPatientJournalParams{
		ID: patient,
		JournalUrl: pgtype.Text{
			String: item.DocumentURL(),
			Valid:  true,
		},
	}); err != nil {
		commonData.Error(commonData.User.Language.TODO("failed to set in DB"), err)
		server.redirectToReferer(w, r)
		return
	}

	if _, err := server.Queries.AddPatientEvent(ctx, AddPatientEventParams{
		PatientID: patient,
		HomeID:    patientData.CurrHomeID.Int32,
		EventID:   int32(EventJournalCreated),
		AppuserID: commonData.User.AppuserID,
		Time:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}); err != nil {
		commonData.Warning(commonData.User.Language.TODO("failed to create event"), err)
	}

	commonData.Success(commonData.User.Language.TODO("document created"))
	server.redirectToReferer(w, r)
}

func (server *Server) attachJournalHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	patient, err := server.getPathID(r, "patient")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	url, err := server.getFormValue(r, "url")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	baseURL := journalRegex.FindString(url)
	if baseURL == "" {
		commonData.Error(commonData.User.Language.TODO("bad URL"), err)
		server.redirectToReferer(w, r)
		return
	}

	patientData, err := server.Queries.GetPatient(ctx, patient)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	if err := server.Queries.SetPatientJournal(ctx, SetPatientJournalParams{
		ID: patient,
		JournalUrl: pgtype.Text{
			String: baseURL,
			Valid:  true,
		},
	}); err != nil {
		commonData.Error(commonData.User.Language.TODO("failed to set in DB"), err)
		server.redirectToReferer(w, r)
		return
	}

	if _, err := server.Queries.AddPatientEvent(ctx, AddPatientEventParams{
		PatientID: patient,
		HomeID:    patientData.CurrHomeID.Int32,
		EventID:   int32(EventJournalAttached),
		AppuserID: commonData.User.AppuserID,
		Time:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}); err != nil {
		commonData.Warning(commonData.User.Language.TODO("failed to create event"), err)
	}

	commonData.Success(commonData.User.Language.TODO("journal attached"))
	server.redirectToReferer(w, r)
}
