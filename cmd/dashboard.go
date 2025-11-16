package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type DashboardData struct {
	PreferredHomeView      HomeView
	DefaultSelectedSpecies int32
	NonPreferredSpecies    []SpeciesView
	Homes                  []HomeView
}

func (server *Server) getSpeciesForUser(ctx context.Context, user int32) ([]SpeciesView, []SpeciesView, error) {
	commonData := MustLoadCommonData(ctx)

	preferredSpeciesRows, err := server.Queries.GetPreferredSpeciesForHome(ctx, GetPreferredSpeciesForHomeParams{
		HomeID:     user,
		LanguageID: commonData.Lang32(),
	})
	if err != nil {
		return nil, nil, err
	}
	preferredSpecies := SliceToSlice(preferredSpeciesRows, func(in GetPreferredSpeciesForHomeRow) SpeciesView {
		return in.ToSpeciesView()
	})

	otherSpeciesRows, err := server.Queries.GetSpeciesWithLanguage(ctx, commonData.Lang32())
	if err != nil {
		return nil, nil, err
	}
	otherSpecies := SliceToSlice(FilterSlice(otherSpeciesRows, func(in GetSpeciesWithLanguageRow) bool {
		return Find(preferredSpecies, func(preferred SpeciesView) bool {
			return preferred.ID == in.SpeciesID
		}) == -1
	}), func(in GetSpeciesWithLanguageRow) SpeciesView {
		return in.ToSpeciesView(false)
	})
	return preferredSpecies, otherSpecies, nil
}

func (server *Server) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	preferredSpecies, otherSpecies, err := server.getSpeciesForUser(ctx, commonData.User.PreferredHome.ID)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	users, err := server.Queries.GetAppusers(ctx) // TODO(perf) use a more specific query
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	homes, err := server.Queries.GetHomes(ctx)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	patients, err := server.Queries.GetActivePatients(ctx, commonData.Lang32())
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	unavailablePeriods, err := server.Queries.GetAllUnavailablePeriods(ctx)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	homeViews := SliceToSlice(homes, func(h Home) HomeView {
		return HomeView{
			Home: h,
			Patients: SliceToSlice(FilterSlice(patients, func(p GetActivePatientsRow) bool {
				return p.CurrHomeID.Valid && p.CurrHomeID.Int32 == h.ID
			}), func(p GetActivePatientsRow) PatientView {
				return PatientView{
					ID:         p.ID,
					Species:    p.Species,
					Name:       p.Name,
					Status:     p.Status,
					JournalURL: p.JournalUrl.String,
				}
			}),
			Users: SliceToSlice(FilterSlice(users, func(u GetAppusersRow) bool {
				return u.HomeID.Valid && u.HomeID.Int32 == h.ID
			}), func(u GetAppusersRow) UserView {
				return u.ToUserView()
			}),
			UnavailablePeriods: SliceToSlice(FilterSlice(unavailablePeriods, func(p HomeUnavailable) bool {
				return p.HomeID == h.ID
			}), func(in HomeUnavailable) PeriodView {
				return in.ToPeriodView()
			}),
		}
	})

	var preferredHomeView HomeView
	if preferredHomeIdx := Find(homes, func(h Home) bool {
		return h.ID == commonData.User.PreferredHome.ID
	}); preferredHomeIdx != -1 {
		preferredHomeView = homeViews[preferredHomeIdx]
		homeViews = append(homeViews[:preferredHomeIdx], homeViews[preferredHomeIdx+1:]...)
	}
	preferredHomeView.PreferredSpecies = preferredSpecies

	defaultSpecies := int32(1)
	if len(preferredSpecies) > 0 {
		defaultSpecies = preferredSpecies[0].ID
	}

	_ = DashboardPage(commonData, &DashboardData{
		NonPreferredSpecies:    otherSpecies,
		DefaultSelectedSpecies: defaultSpecies,
		PreferredHomeView:      preferredHomeView,
		Homes:                  homeViews,
	}).Render(r.Context(), w)
}

func (server *Server) postCheckinHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	if !server.ensureAccess(w, r, AccessLevelRehabber) {
		return
	}

	name, err := server.getFormValue(r, "name")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	var createJournal bool
	if _, err := server.getFormValue(r, "create-journal"); err == nil {
		createJournal = true
	}

	fields, err := server.getFormIDs(r, "home", "species")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	systemSpeciesName, err := server.Queries.GetNameOfSpecies(ctx, GetNameOfSpeciesParams{
		SpeciesID:  fields["species"],
		LanguageID: int32(server.Config.SystemLanguage),
	})
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	var patientID int32
	if err := server.Transaction(ctx, func(ctx context.Context, q *Queries) error {
		var err error
		patientID, err = q.AddPatient(ctx, AddPatientParams{
			SpeciesID:  fields["species"],
			CurrHomeID: pgtype.Int4{Int32: fields["home"], Valid: true},
			Name:       name,
			Status:     int32(StatusAdmitted),
		})
		if err != nil {
			return err
		}

		if _, err := q.AddPatientEvent(ctx, AddPatientEventParams{
			PatientID: patientID,
			AppuserID: commonData.User.AppuserID,
			EventID:   int32(EventRegistered),
			HomeID:    fields["home"],
			Note:      "",
			Time:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}); err != nil {
			return fmt.Errorf("registering patient: %w", err)
		}
		return nil
	}); err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	if createJournal {
		if item, err := server.GDriveWorker.CreateJournal(GDriveTemplateVars{
			Time:    time.Now(),
			Name:    name,
			Species: systemSpeciesName,
			BinoURL: server.Config.BinoURLForPatient(patientID),
		}); err != nil {
			commonData.Warning(commonData.User.Language.GDriveCreateJournalFailed, err)
		} else {
			if tag, err := server.Queries.SetPatientJournal(ctx, SetPatientJournalParams{
				ID:         patientID,
				JournalUrl: pgtype.Text{String: item.DocumentURL(), Valid: true},
			}); err != nil || tag.RowsAffected() == 0 {
				commonData.Warning(commonData.User.Language.GDriveCreateJournalFailed, err)
			}
		}
	}

	server.redirectToReferer(w, r)
}

func (server *Server) movePatientHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	patient, err := server.getPathID(r, "patient")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	newHomeID, err := server.getFormID(r, "home")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	patientData, err := server.Queries.GetPatient(ctx, patient)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	if err := server.Transaction(ctx, func(ctx context.Context, q *Queries) error {
		if err := q.MovePatient(ctx, MovePatientParams{
			ID:         patient,
			CurrHomeID: pgtype.Int4{Int32: newHomeID, Valid: true},
		}); err != nil {
			return err
		}

		q.AddPatientEvent(ctx, AddPatientEventParams{
			PatientID:    patient,
			AppuserID:    commonData.User.AppuserID,
			HomeID:       newHomeID,
			EventID:      int32(EventTransferredToOtherHome),
			AssociatedID: patientData.CurrHomeID,
			Note:         "",
			Time:         pgtype.Timestamptz{Time: time.Now(), Valid: true},
		})

		return nil
	}); err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	server.redirectToReferer(w, r)
}

func (server *Server) postCheckoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	patient, err := server.getPathID(r, "patient")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	status, err := server.getFormID(r, "status")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	note, _ := server.getFormValue(r, "note")

	patientData, err := server.Queries.GetPatient(ctx, patient)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	if err := server.Transaction(ctx, func(ctx context.Context, q *Queries) error {
		now := pgtype.Timestamptz{Time: time.Now(), Valid: true}

		if err := q.CheckoutPatient(ctx, CheckoutPatientParams{
			ID:           patientData.ID,
			TimeCheckout: now,
		}); err != nil {
			return err
		}

		if err := q.SetPatientStatus(ctx, SetPatientStatusParams{
			ID:     patient,
			Status: status,
		}); err != nil {
			return err
		}

		if err := q.MovePatient(ctx, MovePatientParams{
			ID:         patient,
			CurrHomeID: pgtype.Int4{},
		}); err != nil {
			return err
		}

		var event Event
		var associatedID pgtype.Int4
		switch status {
		case int32(StatusDead):
			event = EventDied
		case int32(StatusDeleted):
			event = EventDeleted
		case int32(StatusEuthanized):
			event = EventEuthanized
		case int32(StatusReleased):
			event = EventReleased
		case int32(StatusTransferredOutsideOrganization):
			event = EventTransferredOutsideOrganization
		default:
			event = EventStatusChanged
			associatedID = pgtype.Int4{Int32: int32(status), Valid: true}
		}

		if _, err := q.AddPatientEvent(ctx, AddPatientEventParams{
			PatientID:    patient,
			AppuserID:    commonData.User.AppuserID,
			HomeID:       patientData.CurrHomeID.Int32,
			EventID:      int32(event),
			AssociatedID: associatedID,
			Note:         note,
			Time:         now,
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	server.redirectToReferer(w, r)
}

func (server *Server) postSetNameHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	patient, err := server.getPathID(r, "patient")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	newName, err := server.getFormValue(r, "value")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	patientData, err := server.Queries.GetPatient(ctx, patient)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	if newName == patientData.Name {
		server.redirectToReferer(w, r)
		return
	}

	if err := server.Transaction(ctx, func(ctx context.Context, q *Queries) error {
		if err := q.SetPatientName(ctx, SetPatientNameParams{
			ID:   patient,
			Name: newName,
		}); err != nil {
			return err
		}

		if _, err := q.AddPatientEvent(ctx, AddPatientEventParams{
			PatientID: patient,
			HomeID:    patientData.CurrHomeID.Int32,
			EventID:   int32(EventNameChanged),
			Note:      fmt.Sprintf("'%s' -> '%s'", patientData.Name, newName),
			AppuserID: commonData.User.AppuserID,
			Time:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	server.redirectToReferer(w, r)
}

type AJAXReorderRequest struct {
	ID    int32
	Order []int32
}

type AJAXTransferRequest struct {
	Patient  int32
	Sender   AJAXReorderRequest
	Receiver AJAXReorderRequest
}

func (server *Server) ajaxReorderHandler(w http.ResponseWriter, r *http.Request) {
	jsonHandler(server, w, r, func(q *Queries, req AJAXReorderRequest) error {
		ctx := r.Context()
		return sortPatients(ctx, server.Queries, req)
	})
}

func (server *Server) ajaxTransferHandler(w http.ResponseWriter, r *http.Request) {
	jsonHandler(server, w, r, func(q *Queries, req AJAXTransferRequest) error {
		ctx := r.Context()
		cd := MustLoadCommonData(ctx)
		return server.Transaction(ctx, func(ctx context.Context, q *Queries) error {
			patientData, err := q.GetPatient(ctx, req.Patient)
			if err != nil {
				return err
			}

			if err := q.MovePatient(ctx, MovePatientParams{
				ID:         req.Patient,
				CurrHomeID: pgtype.Int4{Int32: req.Receiver.ID, Valid: true},
			}); err != nil {
				return err
			}

			if _, err := q.AddPatientEvent(ctx, AddPatientEventParams{
				PatientID:    req.Patient,
				AppuserID:    cd.User.AppuserID,
				HomeID:       req.Receiver.ID,
				EventID:      int32(EventTransferredToOtherHome),
				AssociatedID: patientData.CurrHomeID,
				Note:         "",
				Time:         pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}); err != nil {
				return err
			}

			if err := sortPatients(ctx, q, req.Sender); err != nil {
				return err
			}
			if err := sortPatients(ctx, q, req.Receiver); err != nil {
				return err
			}
			return nil
		})
	})
}

func sortPatients(ctx context.Context, q *Queries, req AJAXReorderRequest) error {
	ids := []int32{}
	orders := []int32{}
	for idx, id := range req.Order {
		ids = append(ids, id)
		orders = append(orders, int32(idx))
	}
	return q.UpdatePatientSortOrder(ctx, UpdatePatientSortOrderParams{
		Ids:    ids,
		Orders: orders,
	})
}
