package main

import (
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

	species, err := server.Queries.GetSpeciesWithLanguage(ctx, commonData.Lang32())
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	homes, err := server.Queries.GetHomes(ctx)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	HomePage(ctx, commonData, &DashboardData{
		Species: species,
		Tags:    availableTags,
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
	} else {
		commonData.Success(commonData.User.Language.GenericSuccess)
	}

	server.redirectToReferer(w, r)
}
