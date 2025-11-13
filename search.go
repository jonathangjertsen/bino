//go:generate go tool go-enum --no-iota --values
package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// ENUM(None, Newer, Older)
type TimePreference int

const (
	pageSize = int32(20)
)

type SearchQuery struct {
	Mode           string
	Query          string
	TimePreference TimePreference
	Page           int32
	MinCreated     int64
	MaxCreated     int64
	MinUpdated     int64
	MaxUpdated     int64
}

type SearchSkipInfo struct {
	Reason string
}

type SearchJournalInfo struct {
	FolderURL   string
	FolderName  string
	CreatedTime int64
}

type SearchPatientInfo struct {
	JournalInfo SearchJournalInfo
	JournalURL  string
}

func NewBasicSearchParams(q SearchQuery) SearchBasicParams {
	return SearchBasicParams{
		WFtsHeader: 1.0,
		WFtsBody:   1.0,
		Lang:       "norwegian",
		Offset:     q.Page * pageSize,
		Limit:      pageSize,
		Query:      q.Query,
	}
}

func NewSearchAdvancedParams(q SearchQuery) SearchAdvancedParams {
	wRecency := float32(0.0)
	switch q.TimePreference {
	case TimePreferenceNewer:
		wRecency = 0.3
	case TimePreferenceOlder:
		wRecency = -0.3
	case TimePreferenceNone:
		wRecency = 0.0
	}

	return SearchAdvancedParams{
		Lang:                "norwegian",
		Query:               q.Query,
		WFtsHeader:          1.0,
		WFtsBody:            1.0,
		WSimHeader:          0.4,
		WSimBody:            0.2,
		WIlikeHeader:        0.3,
		WIlikeBody:          0.1,
		WRecency:            wRecency,
		Simthreshold:        0.25,
		RecencyHalfLifeDays: 60,
		Offset:              q.Page * pageSize,
		Limit:               pageSize,
		MinCreated:          pgtype.Timestamptz{Time: time.Unix(q.MinCreated, 0), Valid: q.MinCreated > 0},
		MaxCreated:          pgtype.Timestamptz{Time: time.Unix(q.MaxCreated, 0), Valid: q.MaxCreated > 0},
		MinUpdated:          pgtype.Timestamptz{Time: time.Unix(q.MinUpdated, 0), Valid: q.MinUpdated > 0},
		MaxUpdated:          pgtype.Timestamptz{Time: time.Unix(q.MaxUpdated, 0), Valid: q.MaxUpdated > 0},
	}
}

type SearchResult struct {
	Query        SearchQuery
	PageMatches  []MatchView
	Offset       int32
	TotalMatches int32
	Milliseconds int
}

func (server *Server) doSearch(r *http.Request) (SearchResult, error) {
	q, err := server.getFormValue(r, "q")
	if err != nil {
		return SearchResult{}, err
	}
	if len(q) < 3 {
		return SearchResult{}, errors.New("too short")
	}
	mode, err := server.getFormValue(r, "mode")
	if err != nil {
		mode = "basic"
	}
	page, err := server.getFormID(r, "page")
	if err != nil {
		page = 0
	}

	query := SearchQuery{
		Query:          q,
		Mode:           mode,
		Page:           page,
		TimePreference: TimePreferenceNone,
		MinCreated:     0,
		MaxCreated:     0,
		MinUpdated:     0,
		MaxUpdated:     0,
	}

	var matches []MatchView
	var totalMatches int32
	var offset int32
	t0 := time.Now()
	if mode == "advanced" {
		searchParams := NewSearchAdvancedParams(query)
		rows, err := server.Queries.SearchAdvanced(r.Context(), searchParams)
		if err != nil {
			return SearchResult{Query: query}, err
		}
		matches = SliceToSlice(rows, func(in SearchAdvancedRow) MatchView {
			return in.ToMatchView(q)
		})
		if searchParams.Offset > 0 || len(matches) >= int(searchParams.Limit) {
			totalMatches, err = server.Queries.SearchAdvancedCount(r.Context(), SearchAdvancedCountParams{
				Query:        query.Query,
				Simthreshold: searchParams.Simthreshold,
				Lang:         searchParams.Lang,
			})
			if err != nil {
				LogR(r, "counting: %s", err.Error())
				totalMatches = int32(len(matches))
			}
		} else {
			totalMatches = int32(len(matches))
		}
		offset = searchParams.Offset
	} else {
		searchParams := NewBasicSearchParams(query)
		rows, err := server.Queries.SearchBasic(r.Context(), searchParams)
		if err != nil {
			return SearchResult{Query: query}, err
		}
		matches = SliceToSlice(rows, func(in SearchBasicRow) MatchView {
			return in.ToMatchView()
		})
		if searchParams.Offset > 0 || len(matches) >= int(searchParams.Limit) {
			totalMatches, err = server.Queries.SearchBasicCount(r.Context(), SearchBasicCountParams{
				Query: query.Query,
				Lang:  searchParams.Lang,
			})
			if err != nil {
				LogR(r, "counting: %s", err.Error())
				totalMatches = int32(len(matches))
			}
		} else {
			totalMatches = int32(len(matches))
		}
		offset = searchParams.Offset
	}
	elapsed := time.Since(t0)

	return SearchResult{
		Query:        query,
		PageMatches:  matches,
		TotalMatches: totalMatches,
		Offset:       offset,
		Milliseconds: int(elapsed / time.Millisecond),
	}, nil
}

func (server *Server) emptySearch(w http.ResponseWriter, r *http.Request, result SearchResult, msg string, fullPage bool) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)
	if fullPage {
		_ = SearchPage(commonData, result, msg).Render(ctx, w)
	} else {
		_ = SearchMatches(commonData, result, msg).Render(ctx, w)
	}
}

func (server *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	result, err := server.doSearch(r)
	if err != nil {
		server.emptySearch(w, r, result, err.Error(), true)
		return
	}
	_ = SearchPage(commonData, result, commonData.User.Language.GenericNotFound).Render(ctx, w)
}

func (server *Server) searchLiveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	result, err := server.doSearch(r)
	if err != nil {
		server.emptySearch(w, r, result, err.Error(), false)
		return
	}
	_ = SearchMatches(commonData, result, commonData.User.Language.GenericNotFound).Render(ctx, w)
}
