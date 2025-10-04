package main

import (
	"net/http"

	"github.com/jonathangjertsen/bino/sql"
)

type StatusLangs struct {
	ID int32
	Names
}

func (server *Server) postStatusHandler(w http.ResponseWriter, r *http.Request, commonData *CommonData) {
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

func (server *Server) putStatusHandler(w http.ResponseWriter, r *http.Request, commonData *CommonData) {
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

func (server *Server) getStatusHandler(w http.ResponseWriter, r *http.Request, commonData *CommonData) {
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

	statusView := make([]StatusLangs, 0, len(rows))
	iLangRows := 0
	for _, row := range rows {
		status := StatusLangs{
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

	_ = StatusPage(commonData, statusView).Render(ctx, w)
}
