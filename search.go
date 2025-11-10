package main

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
)

func (server *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	q, err := server.getFormValue(r, "q")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	rows, err := server.Queries.Search(ctx, SearchParams{
		Offset: pgtype.Int4{Int32: 0, Valid: true},
		Limit:  pgtype.Int4{Int32: 20, Valid: true},
		Lang:   "norwegian",
		Query:  q,
	})
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}
	matches := SliceToSlice(rows, func(in SearchRow) MatchView {
		return in.ToMatchView()
	})

	_ = SearchPage(commonData, q, matches).Render(ctx, w)
}

func (server *Server) searchLiveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	q, err := server.getFormValue(r, "q")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	rows, err := server.Queries.Search(ctx, SearchParams{
		Offset: pgtype.Int4{Int32: 0, Valid: true},
		Limit:  pgtype.Int4{Int32: 20, Valid: true},
		Lang:   "norwegian",
		Query:  q,
	})
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}
	matches := SliceToSlice(rows, func(in SearchRow) MatchView {
		return in.ToMatchView()
	})

	_ = SearchMatches(commonData, q, matches).Render(ctx, w)
}
