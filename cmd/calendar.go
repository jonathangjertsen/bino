package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

const (
	timeFormatFullCalendar     = "2006-01-02T15:04:05-07:00"
	timeFormatFullCalendarNoTZ = "2006-01-02T15:04:05"
)

// https://fullcalendar.io/docs/event-parsing
type FullCalendarEvent struct {
	ID      string `json:"id"`
	AllDay  bool   `json:"allDay"`
	Start   string `json:"start"`
	End     string `json:"end,omitempty"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Display string `json:"display"`
}

func (server *Server) calendarHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	initialtime := time.Now().Format(timeFormatFullCalendarNoTZ)
	if altTime, err := server.getQueryValue(r, "t"); err == nil {
		initialtime = altTime
	}
	initialview := "dayGridMonth"
	if altView, err := server.getQueryValue(r, "v"); err == nil {
		initialview = altView
	}

	_ = CalendarPage(commonData, initialtime, initialview).Render(ctx, w)
}

func (server *Server) ajaxCalendarRange(r *http.Request) (time.Time, time.Time, error) {
	start, startErr := server.getQueryValue(r, "start")
	end, endErr := server.getQueryValue(r, "end")

	startT, startParseErr := time.Parse(timeFormatFullCalendar, start)
	endT, endParseErr := time.Parse(timeFormatFullCalendar, end)

	err := errors.Join(startErr, endErr)
	if err == nil {
		err = errors.Join(startParseErr, endParseErr)
	}

	return startT, endT, err
}

func (server *Server) ajaxCalendarAwayHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)
	start, end, err := server.ajaxCalendarRange(r)
	if err != nil {
		logError(r, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	periods, err := server.Queries.GetUnavailablePeriodsInRange(ctx, GetUnavailablePeriodsInRangeParams{
		RangeBegin: pgtype.Date{Time: start, Valid: true},
		RangeEnd:   pgtype.Date{Time: end, Valid: true},
	})
	if err != nil {
		logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	out := SliceToSlice(periods, func(in GetUnavailablePeriodsInRangeRow) FullCalendarEvent {
		return in.ToFullCalendarEvent(commonData.User.Language)
	})
	bin, err := json.Marshal(out)
	if err != nil {
		logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(bin)
}

func (server *Server) ajaxCalendarPatientEventsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)
	start, end, err := server.ajaxCalendarRange(r)
	if err != nil {
		logError(r, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	periods, err := server.Queries.GetEventsForCalendar(ctx, GetEventsForCalendarParams{
		RangeBegin: pgtype.Timestamptz{Time: start, Valid: true},
		RangeEnd:   pgtype.Timestamptz{Time: end, Valid: true},
	})
	if err != nil {
		logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	out := SliceToSlice(periods, func(in GetEventsForCalendarRow) FullCalendarEvent {
		return in.ToFullCalendarEvent(ctx, server, commonData.User.Language)
	})
	bin, err := json.Marshal(out)
	if err != nil {
		logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(bin)
}

func (gupirr GetUnavailablePeriodsInRangeRow) ToFullCalendarEvent(language *Language) FullCalendarEvent {
	return FullCalendarEvent{
		ID:      fmt.Sprintf("unavailable/%d", gupirr.ID),
		AllDay:  true,
		Start:   gupirr.FromDate.Time.Format(timeFormatFullCalendar),
		End:     gupirr.ToDate.Time.Format(timeFormatFullCalendar),
		Title:   language.HomeIsUnavailable(gupirr.Name, gupirr.Note.String),
		URL:     HomeURL(gupirr.HomeID),
		Display: "block",
	}
}

func (gefcr GetEventsForCalendarRow) ToFullCalendarEvent(ctx context.Context, server *Server, language *Language) FullCalendarEvent {
	t := gefcr.Time.Time.Format(timeFormatFullCalendar)
	return FullCalendarEvent{
		ID:      fmt.Sprintf("patientevent/%d", gefcr.ID),
		AllDay:  false,
		Start:   t,
		End:     t,
		Title:   fmt.Sprintf("%s: %s", gefcr.PatientName, language.FormatEvent(ctx, gefcr.EventID, gefcr.AssociatedID, server)),
		URL:     PatientURL(gefcr.PatientID),
		Display: "list-item",
	}
}
