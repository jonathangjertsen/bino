package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type DashboardData struct {
	Species []GetSpeciesWithLanguageRow
	Tags    []GetTagWithLanguageCheckinRow
	Homes   []HomeView
}

type HomeView struct {
	Home     Home
	Patients []PatientView
}

func (hv HomeView) URL() string {
	return fmt.Sprintf("/home/%d", hv.Home.ID)
}

type PatientView struct {
	ID      int32
	Name    string
	Species string
	Tags    []TagView
}

func (pv PatientView) URL() string {
	return fmt.Sprintf("/patient/%d", pv.ID)
}

type TagView struct {
	ID   int32
	Name string
}

func (r GetTagWithLanguageCheckinRow) HTMLID() string {
	return fmt.Sprintf("patient-label-%d", r.TagID)
}

func (server *Server) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	species, err := server.Queries.GetSpeciesWithLanguage(ctx, commonData.User.Language.ID)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	tags, err := server.Queries.GetTagWithLanguageCheckin(ctx, commonData.User.Language.ID)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	homes, err := server.Queries.GetHomes(ctx)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	patients, err := server.Queries.GetActivePatients(ctx, commonData.User.Language.ID)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	patientTags, err := server.Queries.GetTagsForActivePatients(ctx, commonData.User.Language.ID)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	homeViews := MapSlice(homes, func(h Home) HomeView {
		return HomeView{
			Home: h,
			Patients: MapSlice(FilterSlice(patients, func(p GetActivePatientsRow) bool {
				return p.CurrHomeID.Valid && p.CurrHomeID.Int32 == h.ID
			}), func(p GetActivePatientsRow) PatientView {
				return PatientView{
					ID:      p.CurrHomeID.Int32,
					Species: p.Species,
					Name:    p.Name,
					Tags: MapSlice(FilterSlice(patientTags, func(t GetTagsForActivePatientsRow) bool {
						return t.PatientID == p.ID
					}), func(t GetTagsForActivePatientsRow) TagView {
						return TagView{
							ID:   t.TagID,
							Name: t.Name,
						}
					}),
				}
			}),
		}
	})

	if preferredHomeIdx := Find(homes, func(h Home) bool {
		return h.ID == commonData.User.PreferredHomeID
	}); preferredHomeIdx != -1 {
		MoveToFront(homeViews, preferredHomeIdx)
	}

	_ = DashboardPage(commonData, &DashboardData{
		Species: species,
		Tags:    tags,
		Homes:   homeViews,
	}).Render(r.Context(), w)
}

func (server *Server) postDashboardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	name, err := server.getFormValue(r, "name")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	fields, err := server.getFormIDs(r, "home", "species")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	tags, err := IDSlice(r.Form["tag"])
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	admitted := server.getCheckboxValue(r, "admitted")
	status := StatusPendingAdmission
	if admitted {
		status = StatusAdmitted
	}

	if err := server.Transaction(ctx, func(ctx context.Context, q *Queries) error {
		patientID, err := q.AddPatient(ctx, AddPatientParams{
			SpeciesID:  fields["species"],
			CurrHomeID: pgtype.Int4{Int32: fields["home"], Valid: true},
			Name:       name,
			Status:     int32(status),
		})
		if err != nil {
			return err
		}

		if len(tags) > 0 {
			if err := q.AddPatientTags(ctx, AddPatientTagsParams{
				PatientID: patientID,
				Tags:      tags,
			}); err != nil {
				return fmt.Errorf("creating tags: %w", err)
			}
		}

		if _, err := q.AddPatientEvent(ctx, AddPatientEventParams{
			PatientID: patientID,
			EventID:   int32(EventRegistered),
			HomeID:    fields["home"],
			Note:      "",
			Time:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}); err != nil {
			return fmt.Errorf("registering patient: %w", err)
		}

		if admitted {
			if _, err := q.AddPatientEvent(ctx, AddPatientEventParams{
				PatientID: patientID,
				EventID:   int32(EventAdmitted),
				HomeID:    fields["home"],
				Note:      "",
				Time:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}); err != nil {
				return fmt.Errorf("marking patient as admitted: %w", err)
			}
		}

		return nil
	}); err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	fmt.Printf("fields=%+v\n", fields)
	http.Redirect(w, r, "/", http.StatusFound)
}
