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
	logError(r, err)
}

func ajaxError(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	logError(r, err)
	w.WriteHeader(statusCode)
}

func logError(r *http.Request, err error) {
	log.Printf("%s %s ERROR: %v", r.Method, r.URL.Path, err)
}

func (server *Server) fourOhFourHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	server.renderError(w, r, commonData, fmt.Errorf("not found: %s", r.RequestURI))
}

func jsonHandler[T any](
	server *Server,
	w http.ResponseWriter,
	r *http.Request,
	f func(q *sql.Queries, req T) error,
) {
	ctx := r.Context()

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		ajaxError(w, r, err, http.StatusBadRequest)
		return
	}
	var recv T
	if err := json.Unmarshal(bytes, &recv); err != nil {
		ajaxError(w, r, err, http.StatusBadRequest)
		return
	}
	tx, err := server.Conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		ajaxError(w, r, err, http.StatusInternalServerError)
		return
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	q := server.Queries.WithTx(tx)
	if err := f(q, recv); err != nil {
		ajaxError(w, r, err, http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		ajaxError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

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

func (server *Server) postStatusHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	type reqT struct {
		Latin     string
		Languages map[int32]string
	}
	jsonHandler(server, w, r, func(q *sql.Queries, req reqT) error {
		ctx := r.Context()
		id, err := q.AddStatus(ctx)
		if err != nil {
			return err
		}
		for langID, name := range req.Languages {
			if err := q.UpsertStatusLanguage(ctx, sql.UpsertStatusLanguageParams{
				StatusID:   id,
				LanguageID: langID,
				Name:       name,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (server *Server) putStatusHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	type reqT struct {
		ID        int32
		Latin     string
		Languages map[int32]string
	}
	jsonHandler(server, w, r, func(q *sql.Queries, req reqT) error {
		ctx := r.Context()
		for langID, name := range req.Languages {
			if err := q.UpsertStatusLanguage(ctx, sql.UpsertStatusLanguageParams{
				StatusID:   req.ID,
				LanguageID: langID,
				Name:       name,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (server *Server) getStatusHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	ctx := r.Context()

	rows, err := server.Queries.GetStatuses(ctx)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	langRows, err := server.Queries.GetStatusesLanguage(ctx)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	statusView := make([]views.StatusLangs, 0, len(rows))
	iLangRows := 0
	for _, row := range rows {
		status := views.StatusLangs{
			ID:    row,
			Names: map[int32]string{},
		}
		for {
			if iLangRows >= len(langRows) {
				break
			}
			langRow := langRows[iLangRows]
			if langRow.StatusID == row {
				status.Names[langRow.LanguageID] = langRow.Name
				iLangRows++
			} else {
				break
			}
		}

		statusView = append(statusView, status)
	}

	_ = views.StatusPage(commonData, statusView).Render(ctx, w)
}

func (server *Server) postEventHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	type reqT struct {
		Latin     string
		Languages map[int32]string
	}
	jsonHandler(server, w, r, func(q *sql.Queries, req reqT) error {
		ctx := r.Context()
		id, err := q.AddEvent(ctx)
		if err != nil {
			return err
		}
		for langID, name := range req.Languages {
			if err := q.UpsertEventLanguage(ctx, sql.UpsertEventLanguageParams{
				EventID:    id,
				LanguageID: langID,
				Name:       name,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (server *Server) putEventHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	type reqT struct {
		ID        int32
		Latin     string
		Languages map[int32]string
	}
	jsonHandler(server, w, r, func(q *sql.Queries, req reqT) error {
		ctx := r.Context()
		for langID, name := range req.Languages {
			if err := q.UpsertEventLanguage(ctx, sql.UpsertEventLanguageParams{
				EventID:    req.ID,
				LanguageID: langID,
				Name:       name,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (server *Server) getEventHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	ctx := r.Context()

	rows, err := server.Queries.GetEvents(ctx)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	langRows, err := server.Queries.GetEventsLanguage(ctx)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	eventView := make([]views.EventLangs, 0, len(rows))
	iLangRows := 0
	for _, row := range rows {
		status := views.EventLangs{
			ID:    row,
			Names: map[int32]string{},
		}
		for {
			if iLangRows >= len(langRows) {
				break
			}
			langRow := langRows[iLangRows]
			if langRow.EventID == row {
				status.Names[langRow.LanguageID] = langRow.Name
				iLangRows++
			} else {
				break
			}
		}

		eventView = append(eventView, status)
	}

	_ = views.EventPage(commonData, eventView).Render(ctx, w)
}

func (server *Server) postTagHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	type reqT struct {
		DefaultShow bool
		Languages   map[int32]string
	}
	jsonHandler(server, w, r, func(q *sql.Queries, req reqT) error {
		ctx := r.Context()
		id, err := q.AddTag(ctx, req.DefaultShow)
		if err != nil {
			return err
		}
		for langID, name := range req.Languages {
			if err := q.UpsertTagLanguage(ctx, sql.UpsertTagLanguageParams{
				TagID:      id,
				LanguageID: langID,
				Name:       name,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (server *Server) putTagHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	type reqT struct {
		ID          int32
		DefaultShow bool
		Languages   map[int32]string
	}
	jsonHandler(server, w, r, func(q *sql.Queries, req reqT) error {
		ctx := r.Context()
		err := q.UpdateTagDefaultShown(ctx, sql.UpdateTagDefaultShownParams{ID: req.ID, DefaultShow: req.DefaultShow})
		if err != nil {
			return err
		}
		for langID, name := range req.Languages {
			if err := q.UpsertTagLanguage(ctx, sql.UpsertTagLanguageParams{
				TagID:      req.ID,
				LanguageID: langID,
				Name:       name,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (server *Server) getTagHandler(w http.ResponseWriter, r *http.Request, commonData *views.CommonData) {
	ctx := r.Context()

	rows, err := server.Queries.GetTags(ctx)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	langRows, err := server.Queries.GetTagsLanguage(ctx)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	tagView := make([]views.TagLangs, 0, len(rows))
	iLangRows := 0
	for _, row := range rows {
		status := views.TagLangs{
			ID:          row.ID,
			DefaultShow: row.DefaultShow,
			Names:       map[int32]string{},
		}
		for {
			if iLangRows >= len(langRows) {
				break
			}
			langRow := langRows[iLangRows]
			if langRow.TagID == row.ID {
				status.Names[langRow.LanguageID] = langRow.Name
				iLangRows++
			} else {
				break
			}
		}

		tagView = append(tagView, status)
	}

	_ = views.TagPage(commonData, tagView).Render(ctx, w)
}
