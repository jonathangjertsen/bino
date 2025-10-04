package main

import (
	"net/http"

	"github.com/jonathangjertsen/bino/sql"
)

type EventLangs struct {
	ID int32
	Names
}

func (server *Server) postEventHandler(w http.ResponseWriter, r *http.Request, commonData *CommonData) {
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

func (server *Server) putEventHandler(w http.ResponseWriter, r *http.Request, commonData *CommonData) {
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

func (server *Server) getEventHandler(w http.ResponseWriter, r *http.Request, commonData *CommonData) {
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

	eventView := make([]EventLangs, 0, len(rows))
	iLangRows := 0
	for _, row := range rows {
		status := EventLangs{
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

	_ = EventPage(commonData, eventView).Render(ctx, w)
}
