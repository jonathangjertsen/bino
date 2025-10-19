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

	patients, err := server.Queries.GetCurrentPatientsForHome(ctx, GetCurrentPatientsForHomeParams{
		CurrHomeID: pgtype.Int4{Int32: id, Valid: true},
		LanguageID: commonData.Lang32(),
	})
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	tags, err := server.Queries.GetTagsForCurrentPatientsForHome(ctx, GetTagsForCurrentPatientsForHomeParams{
		CurrHomeID: pgtype.Int4{Int32: id, Valid: true},
		LanguageID: commonData.Lang32(),
	})
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	HomePage(ctx, commonData, HomeView{
		Home: homeData,
		Patients: SliceToSlice(patients, func(p GetCurrentPatientsForHomeRow) PatientView {
			return PatientView{
				ID:      p.ID,
				Status:  p.Status,
				Name:    p.Name,
				Species: p.SpeciesName,
				Tags: SliceToSlice(FilterSlice(tags, func(t GetTagsForCurrentPatientsForHomeRow) bool {
					return t.PatientID == p.ID
				}), func(t GetTagsForCurrentPatientsForHomeRow) TagView {
					return t.ToTagView()
				}),
			}
		}),
	}).Render(r.Context(), w)
}
