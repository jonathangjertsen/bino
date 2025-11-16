package main

import "net/http"

func (server *Server) formerPatientsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	patients, err := server.Queries.GetFormerPatients(ctx, commonData.Lang32())
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	FormerPatients(commonData, SliceToSlice(patients, func(in GetFormerPatientsRow) PatientView {
		return in.ToPatientView()
	})).Render(ctx, w)
}
