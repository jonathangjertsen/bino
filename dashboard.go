package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type DashboardData struct {
	Species           []GetSpeciesWithLanguageRow
	Tags              []GetTagsWithLanguageCheckinRow
	PreferredHomeView HomeView
	Homes             []HomeView
}

func (r GetTagsWithLanguageCheckinRow) HTMLID() string {
	return fmt.Sprintf("patient-label-%d", r.TagID)
}

func (server *Server) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	species, err := server.Queries.GetSpeciesWithLanguage(ctx, commonData.Lang32())
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	tags, err := server.Queries.GetTagsWithLanguageCheckin(ctx, commonData.Lang32())
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

	patientTags, err := server.Queries.GetTagsForActivePatients(ctx, commonData.Lang32())
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
					Tags: SliceToSlice(FilterSlice(patientTags, func(t GetTagsForActivePatientsRow) bool {
						return t.PatientID == p.ID
					}), func(t GetTagsForActivePatientsRow) TagView {
						return t.ToTagView()
					}),
				}
			}),
			Users: SliceToSlice(FilterSlice(users, func(u GetAppusersRow) bool {
				return u.HomeID.Valid && u.HomeID.Int32 == h.ID
			}), func(u GetAppusersRow) UserView {
				return u.ToUserView()
			}),
		}
	})

	var preferredHomeView HomeView
	if preferredHomeIdx := Find(homes, func(h Home) bool {
		return h.ID == commonData.User.PreferredHomeID
	}); preferredHomeIdx != -1 {
		preferredHomeView = homeViews[preferredHomeIdx]
		homeViews = append(homeViews[:preferredHomeIdx], homeViews[preferredHomeIdx+1:]...)
	}

	_ = DashboardPage(commonData, &DashboardData{
		Species:           species,
		Tags:              tags,
		PreferredHomeView: preferredHomeView,
		Homes:             homeViews,
	}).Render(r.Context(), w)
}

func (server *Server) postCheckinHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	name, err := server.getFormValue(r, "name")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	fields, err := server.getFormIDs(r, "home", "species")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	tags, err := IDSlice(r.Form["tag"])
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	admitted := server.getCheckboxValue(r, "admitted")
	status := StatusPendingAdmission
	if admitted {
		status = StatusAdmitted
	}

	if err := server.Transaction(ctx, func(ctx context.Context, q *Queries) error {
		patientID, err := q.AddPatient(ctx, AddPatientParams{
			SpeciesID:  fields["species"],
			CurrHomeID: pgtype.Int4{Int32: fields["home"], Valid: true},
			Name:       name,
			Status:     int32(status),
		})
		if err != nil {
			return err
		}

		if len(tags) > 0 {
			if err := q.AddPatientTags(ctx, AddPatientTagsParams{
				PatientID: patientID,
				Tags:      tags,
			}); err != nil {
				return fmt.Errorf("creating tags: %w", err)
			}
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

		if admitted {
			if _, err := q.AddPatientEvent(ctx, AddPatientEventParams{
				PatientID: patientID,
				AppuserID: commonData.User.AppuserID,
				EventID:   int32(EventAdmitted),
				HomeID:    fields["home"],
				Note:      "",
				Time:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}); err != nil {
				return fmt.Errorf("marking patient as admitted: %w", err)
			}
		}

		return nil
	}); err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	server.redirectToReferer(w, r)
}

func (server *Server) deletePatientTagHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	fields, err := server.getPathIDs(r, "patient", "tag")
	if err != nil {
		ajaxError(w, r, err, http.StatusInternalServerError)
		return
	}

	patient, tag := fields["patient"], fields["tag"]

	patientData, err := server.Queries.GetPatient(ctx, patient)
	if err != nil {
		ajaxError(w, r, err, http.StatusInternalServerError)
		return
	}

	if err := server.Transaction(ctx, func(ctx context.Context, q *Queries) error {
		if err := q.DeletePatientTag(ctx, DeletePatientTagParams{
			PatientID: patient,
			TagID:     tag,
		}); err != nil {
			return err
		}

		if _, err := q.AddPatientEvent(ctx, AddPatientEventParams{
			PatientID:    patient,
			HomeID:       patientData.CurrHomeID.Int32,
			EventID:      int32(EventTagRemoved),
			AssociatedID: pgtype.Int4{Int32: tag, Valid: true},
			Note:         "",
			AppuserID:    commonData.User.AppuserID,
			Time:         pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		ajaxError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (server *Server) createPatientTagHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	fields, err := server.getPathIDs(r, "patient", "tag")
	if err != nil {
		ajaxError(w, r, err, http.StatusInternalServerError)
		return
	}

	patient, tag := fields["patient"], fields["tag"]

	patientData, err := server.Queries.GetPatient(ctx, patient)
	if err != nil {
		ajaxError(w, r, err, http.StatusInternalServerError)
		return
	}

	if err := server.Transaction(ctx, func(ctx context.Context, q *Queries) error {
		if err := q.AddPatientTags(ctx, AddPatientTagsParams{
			PatientID: patient,
			Tags:      []int32{tag},
		}); err != nil {
			return err
		}

		if _, err := q.AddPatientEvent(ctx, AddPatientEventParams{
			PatientID:    patient,
			HomeID:       patientData.CurrHomeID.Int32,
			EventID:      int32(EventTagAdded),
			AssociatedID: pgtype.Int4{Int32: tag, Valid: true},
			Note:         "",
			AppuserID:    commonData.User.AppuserID,
			Time:         pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		ajaxError(w, r, err, http.StatusInternalServerError)
		return
	}

	tagName, err := server.Queries.GetTagName(ctx, GetTagNameParams{
		LanguageID: commonData.Lang32(),
		TagID:      tag,
	})
	if err != nil {
		ajaxError(w, r, err, http.StatusInternalServerError)
		return
	}

	DashboardTag(commonData, TagView{
		ID:        tag,
		PatientID: patient,
		Name:      tagName,
	}).Render(ctx, w)
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
			Time:         pgtype.Timestamptz{Time: time.Now(), Valid: true},
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
