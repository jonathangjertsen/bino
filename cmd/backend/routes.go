package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jonathangjertsen/bino/sql"
	"github.com/jonathangjertsen/bino/views"
)

func (server *Server) adminRootHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	ctx := r.Context()
	_ = views.AdminRootPage(commonData).Render(ctx, w)
}

func (server *Server) postLanguageHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	ctx := r.Context()

	lang, err := getSelectedLanguage(r.FormValue("language"), commonData)
	if err == nil {
		err = server.Queries.SetUserLanguage(ctx, sql.SetUserLanguageParams{
			AppuserID:  commonData.User.AppuserID,
			LanguageID: lang,
		})
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusFound)
}

func getSelectedLanguage(langStr string, commonData *views.CommonData) (int32, error) {
	langID, err := strconv.Atoi(langStr)
	if err != nil {
		return 0, fmt.Errorf("invalid language ID: %w", langStr)
	}
	for _, lang := range commonData.Languages {
		if lang.ID == int32(langID) {
			return lang.ID, nil
		}
	}
	return 0, fmt.Errorf("unsupported language ID: %d", langID)
}

func (server *Server) renderError(w http.ResponseWriter, r *http.Request, commonData *views.CommonData, err error) {
	ctx := r.Context()
	_ = views.ErrorPage(commonData, err).Render(ctx, w)
	log.Printf("%s %s ERROR: %v", r.Method, r.URL.Path, err)
}

func (server *Server) fourOhFourHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	server.renderError(w, r, commonData, fmt.Errorf("not found: %s", r.RequestURI))
}

func (server *Server) dashboardHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	ctx := r.Context()

	rows, err := server.Queries.GetSpeciesWithLanguage(ctx, commonData.User.LanguageID)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	species := make([]views.Species, 0, len(rows))
	for _, row := range rows {
		species = append(species, views.Species{
			ID:   row.SpeciesID,
			Name: row.Name,
		})
	}

	// TODO: placeholder
	labels := []views.Label{
		{ID: 1, Name: "Critical", Checked: false},
		{ID: 1, Name: "Disease", Checked: false},
		{ID: 1, Name: "Injury", Checked: false},
		{ID: 1, Name: "Force feeding", Checked: false},
		{ID: 1, Name: "Juvenile", Checked: false},
	}

	_ = views.DashboardPage(commonData, species, labels).Render(r.Context(), w)
}

func (server *Server) postSpeciesHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	ctx := r.Context()

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}
	var req struct {
		Latin     string
		Languages map[int32]string
	}
	if err := json.Unmarshal(bytes, &req); err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	tx, err := server.Conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	q := server.Queries.WithTx(tx)
	id, err := q.AddSpecies(ctx, req.Latin)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}
	for langID, name := range req.Languages {
		if err := q.UpsertSpeciesLanguage(ctx, sql.UpsertSpeciesLanguageParams{
			SpeciesID:  id,
			LanguageID: langID,
			Name:       name,
		}); err != nil {
			server.renderError(w, r, commonData, err)
			return
		}
	}
	_ = tx.Commit(ctx)

	w.WriteHeader(http.StatusOK)
}

func (server *Server) putSpeciesHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	ctx := r.Context()

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}
	var req struct {
		ID        int32
		Latin     string
		Languages map[int32]string
	}
	if err := json.Unmarshal(bytes, &req); err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	tx, err := server.Conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	q := server.Queries.WithTx(tx)
	for langID, name := range req.Languages {
		if err := q.UpsertSpeciesLanguage(ctx, sql.UpsertSpeciesLanguageParams{
			SpeciesID:  req.ID,
			LanguageID: langID,
			Name:       name,
		}); err != nil {
			server.renderError(w, r, commonData, err)
			return
		}
	}
	if err := tx.Commit(ctx); err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	w.WriteHeader(http.StatusOK)
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
