package main

import (
	"fmt"
	"net/http"
)

type Species struct {
	ID   int32
	Name string
}

type Tag struct {
	ID   int32
	Name string
}

func (l Tag) HTMLID() string {
	return fmt.Sprintf("patient-label-%d", l.ID)
}

func (server *Server) dashboardHandler(w http.ResponseWriter, r *http.Request, commonData *CommonData) {
	ctx := r.Context()

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
		tags = append(tags, Tag{ID: row.TagID, Name: row.Name})
	}

	_ = DashboardPage(commonData, species, tags).Render(r.Context(), w)
}
