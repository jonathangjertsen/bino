//go:generate go tool go-enum --no-iota --values
package main

import (
	"errors"
	"net/http"
)

// ENUM(None, Newer, Older)
type TimePreference int

const (
	pageSize = int32(20)
)

func NewBasicSearchParams(q string, page int32) SearchBasicParams {
	return SearchBasicParams{
		WFtsHeader: 1.0,
		WFtsBody:   1.0,
		Lang:       "norwegian",
		Offset:     page * pageSize,
		Limit:      pageSize,
		Query:      q,
	}
}

func NewSearchAdvancedParams(q string, page int32, tp TimePreference) SearchAdvancedParams {
	wRecency := float32(0.0)
	switch tp {
	case TimePreferenceNewer:
		wRecency = 0.3
	case TimePreferenceOlder:
		wRecency = -0.3
	case TimePreferenceNone:
		wRecency = 0.0
	}

	return SearchAdvancedParams{
		Lang:                "norwegian",
		Query:               q,
		WFtsHeader:          1.0,
		WFtsBody:            1.0,
		WSimHeader:          0.4,
		WSimBody:            0.2,
		WIlikeHeader:        0.3,
		WIlikeBody:          0.1,
		WRecency:            wRecency,
		Simthreshold:        0.25,
		RecencyHalfLifeDays: 60,
		Offset:              page * pageSize,
		Limit:               pageSize,
	}
}

func (server *Server) doSearch(r *http.Request) (string, []MatchView, error) {
	q, err := server.getFormValue(r, "q")
	if err != nil {
		return "", nil, err
	}
	if len(q) < 3 {
		return q, nil, errors.New("too short")
	}
	mode, err := server.getFormValue(r, "mode")
	if err != nil {
		mode = "basic"
	}
	page := int32(0)
	var matches []MatchView
	if mode == "advanced" {
		searchParams := NewSearchAdvancedParams(q, page, TimePreferenceNone)
		rows, err := server.Queries.SearchAdvanced(r.Context(), searchParams)
		if err != nil {
			return q, nil, err
		}
		matches = SliceToSlice(rows, func(in SearchAdvancedRow) MatchView {
			return in.ToMatchView(q)
		})
	} else {
		searchParams := NewBasicSearchParams(q, page)
		rows, err := server.Queries.SearchBasic(r.Context(), searchParams)
		if err != nil {
			return q, nil, err
		}
		matches = SliceToSlice(rows, func(in SearchBasicRow) MatchView {
			return in.ToMatchView()
		})
	}

	return q, matches, nil
}

func (server *Server) emptySearch(w http.ResponseWriter, r *http.Request, q, msg string, fullPage bool) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)
	if fullPage {
		_ = SearchPage(commonData, q, nil, msg).Render(ctx, w)
	} else {
		_ = SearchMatches(commonData, q, nil, msg).Render(ctx, w)
	}
}

func (server *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	q, matches, err := server.doSearch(r)
	if err != nil {
		server.emptySearch(w, r, q, err.Error(), true)
		return
	}
	_ = SearchPage(commonData, q, matches, commonData.User.Language.GenericNotFound).Render(ctx, w)
}

func (server *Server) searchLiveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	q, matches, err := server.doSearch(r)
	if err != nil {
		server.emptySearch(w, r, q, err.Error(), false)
		return
	}
	_ = SearchMatches(commonData, q, matches, commonData.User.Language.GenericNotFound).Render(ctx, w)
}
