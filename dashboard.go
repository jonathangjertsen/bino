package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jonathangjertsen/bino/sql"
)

type Species struct {
	ID   int32
	Name string
}

type Tag struct {
	ID          int32
	Name        string
	DefaultShow bool
}

func (l Tag) HTMLID() string {
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

	species := make([]Species, 0, len(speciesRows))
	for _, row := range speciesRows {
		species = append(species, Species{
			ID:   row.SpeciesID,
			Name: row.Name,
		})
	}

	tagRows, err := server.Queries.GetTagWithLanguageCheckin(ctx, commonData.User.LanguageID)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}
	tags := make([]Tag, 0, len(tagRows))
	for _, row := range tagRows {
		tags = append(tags, Tag{ID: row.TagID, Name: row.Name, DefaultShow: row.DefaultShow})
	}

	homeRows, err := server.Queries.GetHomes(ctx)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}
	homes := make([]Home, 0, len(homeRows))
	for _, row := range homeRows {
		homes = append(homes, Home{
			ID:    row.ID,
			Name:  row.Name,
			Users: nil,
		})
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

	if err := server.Transaction(ctx, func(ctx context.Context, q *sql.Queries) error {
		patientID, err := q.AddPatient(ctx, sql.AddPatientParams{
			SpeciesID:    fields["species"],
			CurrStatusID: 0,
			CurrHomeID:   pgtype.Int4{Int32: fields["home"], Valid: true},
			Name:         name,
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

		return nil
	}); err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	fmt.Printf("fields=%+v\n", fields)
	http.Redirect(w, r, "/", http.StatusFound)
}
