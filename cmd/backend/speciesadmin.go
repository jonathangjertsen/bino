package main

import (
	"net/http"

	"github.com/jonathangjertsen/bino/sql"
	"github.com/jonathangjertsen/bino/views"
)

func (server *Server) postSpeciesHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	type reqT struct {
		Latin     string
		Languages map[int32]string
	}
	jsonHandler(server, w, r, func(q *sql.Queries, req reqT) error {
		ctx := r.Context()
		id, err := q.AddSpecies(ctx, req.Latin)
		if err != nil {
			return err
		}
		for langID, name := range req.Languages {
			if err := q.UpsertSpeciesLanguage(ctx, sql.UpsertSpeciesLanguageParams{
				SpeciesID:  id,
				LanguageID: langID,
				Name:       name,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (server *Server) putSpeciesHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	type reqT struct {
		ID        int32
		Latin     string
		Languages map[int32]string
	}
	jsonHandler(server, w, r, func(q *sql.Queries, req reqT) error {
		ctx := r.Context()
		for langID, name := range req.Languages {
			if err := q.UpsertSpeciesLanguage(ctx, sql.UpsertSpeciesLanguageParams{
				SpeciesID:  req.ID,
				LanguageID: langID,
				Name:       name,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (server *Server) getSpeciesHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	ctx := r.Context()

	rows, err := server.Queries.GetSpecies(ctx)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	langRows, err := server.Queries.GetSpeciesLanguage(ctx)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	speciesView := make([]views.SpeciesLangs, 0, len(rows))
	iLangRows := 0
	for _, row := range rows {
		species := views.SpeciesLangs{
			ID:        row.ID,
			LatinName: row.ScientificName,
			Names:     map[int32]string{},
		}
		for {
			if iLangRows >= len(langRows) {
				break
			}
			langRow := langRows[iLangRows]
			if langRow.SpeciesID == row.ID {
				species.Names[langRow.LanguageID] = langRow.Name
				iLangRows++
			} else {
				break
			}
		}

		speciesView = append(speciesView, species)
	}

	_ = views.SpeciesPage(commonData, speciesView).Render(ctx, w)
}
