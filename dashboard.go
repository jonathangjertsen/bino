package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jonathangjertsen/bino/sql"
)

type SpeciesView struct {
	ID   int32
	Name string
}

type TagView struct {
	ID          int32
	Name        string
	DefaultShow bool
}

func (l TagView) HTMLID() string {
	return fmt.Sprintf("patient-label-%d", l.ID)
}

func (server *Server) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	speciesRows, err := server.Queries.GetSpeciesWithLanguage(ctx, commonData.User.LanguageID)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	species := make([]SpeciesView, 0, len(speciesRows))
	for _, row := range speciesRows {
		species = append(species, SpeciesView{
			ID:   row.SpeciesID,
			Name: row.Name,
		})
	}

	tagRows, err := server.Queries.GetTagWithLanguageCheckin(ctx, commonData.User.LanguageID)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	tags := make([]TagView, 0, len(tagRows))
	for _, row := range tagRows {
		tags = append(tags, TagView{ID: row.TagID, Name: row.Name, DefaultShow: row.DefaultShow})
	}

	homeRows, err := server.Queries.GetHomes(ctx)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}
	homes := make([]HomeView, 0, len(homeRows))
	for _, row := range homeRows {
		if row.ID == commonData.User.PreferredHomeID {
			homes = append(homes, HomeView{
				ID:    row.ID,
				Name:  row.Name,
				Users: nil,
			})
		}
	}
	for _, row := range homeRows {
		if row.ID != commonData.User.PreferredHomeID {
			homes = append(homes, HomeView{
				ID:    row.ID,
				Name:  row.Name,
				Users: nil,
			})
		}
	}
	_ = DashboardPage(commonData, species, tags, homes).Render(r.Context(), w)
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

	if err := server.Transaction(ctx, func(ctx context.Context, q *sql.Queries) error {
		patientID, err := q.AddPatient(ctx, sql.AddPatientParams{
			SpeciesID:  fields["species"],
			CurrHomeID: pgtype.Int4{Int32: fields["home"], Valid: true},
			Name:       name,
			Status:     int32(status),
		})
		if err != nil {
			return err
		}

		if len(tags) > 0 {
			if err := q.AddPatientTags(ctx, sql.AddPatientTagsParams{
				PatientID: patientID,
				Tags:      tags,
			}); err != nil {
				return fmt.Errorf("creating tags: %w", err)
			}
		}

		if _, err := q.AddPatientEvent(ctx, sql.AddPatientEventParams{
			PatientID: patientID,
			EventID:   int32(EventRegistered),
			HomeID:    fields["home"],
			Note:      "",
			Time:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}); err != nil {
			return fmt.Errorf("registering patient: %w", err)
		}

		if admitted {
			if _, err := q.AddPatientEvent(ctx, sql.AddPatientEventParams{
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
