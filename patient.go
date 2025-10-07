package main

import (
	"fmt"
	"net/http"
)

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
			ID:      patientData.ID,
			Status:  patientData.Status,
			Name:    patientData.Name,
			Species: patientData.SpeciesName,
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
