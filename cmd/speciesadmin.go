package main

import (
	"net/http"
)

type SpeciesLangs struct {
	ID        int32
	LatinName string
	Names
}

func (server *Server) postSpeciesHandler(w http.ResponseWriter, r *http.Request) {
	type reqT struct {
		Latin     string
		Languages map[int32]string
	}
	jsonHandler(server, w, r, func(q *Queries, req reqT) error {
		ctx := r.Context()
		id, err := q.AddSpecies(ctx, req.Latin)
		if err != nil {
			return err
		}
		for langID, name := range req.Languages {
			if err := q.UpsertSpeciesLanguage(ctx, UpsertSpeciesLanguageParams{
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

func (server *Server) putSpeciesHandler(w http.ResponseWriter, r *http.Request) {
	type reqT struct {
		ID        int32
		Latin     string
		Languages map[int32]string
	}
	jsonHandler(server, w, r, func(q *Queries, req reqT) error {
		ctx := r.Context()
		for langID, name := range req.Languages {
			if err := q.UpsertSpeciesLanguage(ctx, UpsertSpeciesLanguageParams{
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

func (server *Server) getSpeciesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

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

	speciesView := make([]SpeciesLangs, 0, len(rows))
	iLangRows := 0
	for _, row := range rows {
		species := SpeciesLangs{
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

	_ = SpeciesPage(commonData, speciesView).Render(ctx, w)
}
