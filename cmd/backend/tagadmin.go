package main

import (
	"net/http"

	"github.com/jonathangjertsen/bino/sql"
	"github.com/jonathangjertsen/bino/views"
)

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
