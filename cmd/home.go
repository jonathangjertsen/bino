package main

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
)

func (server *Server) getHomeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	id, err := server.getPathID(r, "home")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	homeData, err := server.Queries.GetHome(ctx, id)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	users, err := server.Queries.GetAppusersForHome(ctx, id)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	patients, err := server.Queries.GetCurrentPatientsForHome(ctx, GetCurrentPatientsForHomeParams{
		CurrHomeID: pgtype.Int4{Int32: id, Valid: true},
		LanguageID: commonData.Lang32(),
	})
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	patientTags, err := server.Queries.GetTagsForCurrentPatientsForHome(ctx, GetTagsForCurrentPatientsForHomeParams{
		CurrHomeID: pgtype.Int4{Int32: id, Valid: true},
		LanguageID: commonData.Lang32(),
	})

	availableTags, err := server.Queries.GetTagsWithLanguageCheckin(ctx, commonData.Lang32())
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	homes, err := server.Queries.GetHomes(ctx)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	preferredSpecies, otherSpecies, err := server.getSpeciesForUser(ctx, homeData.ID)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	unavailablePeriods, err := server.Queries.GetHomeUnavailablePeriods(ctx, id)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	HomePage(ctx, commonData, &DashboardData{
		NonPreferredSpecies: otherSpecies,
		Tags:                availableTags,
		Homes: SliceToSlice(homes, func(h Home) HomeView {
			return h.ToHomeView()
		}),
	}, &HomeView{
		Home: homeData,
		Users: SliceToSlice(users, func(u Appuser) UserView {
			return u.ToUserView()
		}),
		Patients: SliceToSlice(patients, func(p GetCurrentPatientsForHomeRow) PatientView {
			return PatientView{
				ID:      p.ID,
				Status:  p.Status,
				Name:    p.Name,
				Species: p.SpeciesName,
				Tags: SliceToSlice(FilterSlice(patientTags, func(t GetTagsForCurrentPatientsForHomeRow) bool {
					return t.PatientID == p.ID
				}), func(t GetTagsForCurrentPatientsForHomeRow) TagView {
					return t.ToTagView()
				}),
			}
		}),
		PreferredSpecies: preferredSpecies,
		UnavailablePeriods: SliceToSlice(unavailablePeriods, func(in HomeUnavailable) PeriodView {
			return in.ToPeriodView()
		}),
	}).Render(r.Context(), w)
}

func (server *Server) setCapacityHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	id, err := server.getPathID(r, "home")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	capacity, err := server.getFormID(r, "capacity")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	if err := server.Queries.SetHomeCapacity(ctx, SetHomeCapacityParams{
		ID:       id,
		Capacity: capacity,
	}); err != nil {
		commonData.Error(commonData.User.Language.GenericFailed, err)
	}

	server.redirectToReferer(w, r)
}

func (server *Server) addPreferredSpeciesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	id, err := server.getPathID(r, "home")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	species, err := server.getFormID(r, "species")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	if err := server.Queries.AddPreferredSpecies(ctx, AddPreferredSpeciesParams{
		HomeID:    id,
		SpeciesID: species,
	}); err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	server.redirectToReferer(w, r)
}

func (server *Server) addHomeUnavailablePeriodHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	homeID, err := server.getPathID(r, "home")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	values, err := server.getFormValues(r, "unavailable-from", "unavailable-to", "unavailable-note")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	var fromV DateView
	var toV DateView
	note, hasNote := values["unavailable-note"]

	if n, err := fmt.Sscanf(values["unavailable-from"], "%d-%d-%d", &fromV.Year, &fromV.Month, &fromV.Day); err != nil || n != 3 {
		commonData.Warning(commonData.User.Language.HomePeriodInvalid, err)
		server.redirectToReferer(w, r)
		return
	}
	if n, err := fmt.Sscanf(values["unavailable-to"], "%d-%d-%d", &toV.Year, &toV.Month, &toV.Day); err != nil || n != 3 {
		commonData.Warning(commonData.User.Language.HomePeriodInvalid, err)
		server.redirectToReferer(w, r)
		return
	}

	if toV.Before(fromV) {
		commonData.Warning(commonData.User.Language.HomePeriodInvalid, fmt.Errorf("to is before from: %+v < %+v", toV, fromV))
		server.redirectToReferer(w, r)
		return
	}
	if _, err := server.Queries.AddHomeUnavailablePeriod(ctx, AddHomeUnavailablePeriodParams{
		HomeID:   homeID,
		FromDate: fromV.ToPGDate(),
		ToDate:   toV.ToPGDate(),
		Note:     pgtype.Text{String: note, Valid: hasNote && note != ""},
	}); err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	server.redirectToReferer(w, r)
}

func (server *Server) deleteHomeUnavailableHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	periodID, err := server.getPathID(r, "period")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	if err := server.Queries.DeleteHomeUnavailablePeriod(ctx, periodID); err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	server.redirectToReferer(w, r)
}

func (server *Server) homeSetNoteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	homeID, err := server.getPathID(r, "home")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	note, err := server.getFormValue(r, "value")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	if err := server.Queries.SetHomeNote(ctx, SetHomeNoteParams{
		ID:   homeID,
		Note: note,
	}); err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	server.redirectToReferer(w, r)
}
