package main

import (
	"net/http"

	"github.com/jonathangjertsen/bino/views"
)

func (server *Server) dashboardHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	ctx := r.Context()

	speciesRows, err := server.Queries.GetSpeciesWithLanguage(ctx, commonData.User.LanguageID)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	species := make([]views.Species, 0, len(speciesRows))
	for _, row := range speciesRows {
		species = append(species, views.Species{
			ID:   row.SpeciesID,
			Name: row.Name,
		})
	}

	tagRows, err := server.Queries.GetTagWithLanguage(ctx, commonData.User.LanguageID)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}
	tags := make([]views.Tag, 0, len(tagRows))
	for _, row := range tagRows {
		tags = append(tags, views.Tag{ID: row.TagID, Name: row.Name})
	}

	_ = views.DashboardPage(commonData, species, tags).Render(r.Context(), w)
}
