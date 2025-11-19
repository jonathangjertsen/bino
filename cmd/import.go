package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

type ImportRequest struct {
	CreateJournals string
	Patients       []ImportPatient
	OK             bool
	Notes          []string
	Txt            string
}

type ImportPatient struct {
	Name      string
	HomeID    int32
	SpeciesID int32
	URL       string
}

func (server *Server) getImportHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	var ir ImportRequest
	server.getCookie(w, r, "import-request", &ir)

	_ = ImportPage(commonData, ir).Render(ctx, w)
}

func (server *Server) postImportHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	result := server.parseImportForm(r)
	if result.OK {
		var patientsRequiringJournal []int32

		if err := server.Transaction(ctx, func(ctx context.Context, q *Queries) error {
			addPatientParams := AddPatientsParams{}
			for _, patient := range result.Patients {
				addPatientParams.CurrHomeID = append(addPatientParams.CurrHomeID, patient.HomeID)
				addPatientParams.Species = append(addPatientParams.Species, patient.SpeciesID)
				addPatientParams.Status = append(addPatientParams.Status, int32(StatusAdmitted))
				addPatientParams.JournalUrl = append(addPatientParams.JournalUrl, patient.URL)
				addPatientParams.Name = append(addPatientParams.Name, patient.Name)

			}

			if ids, err := q.AddPatients(ctx, addPatientParams); err != nil {
				result.Notes = []string{fmt.Sprintf("Error adding patients: %v", err)}
				return err
			} else {
				result.Notes = []string{fmt.Sprintf("Added %d patients", len(ids))}
				addPatientRegisteredEventsParams := AddPatientRegisteredEventsParams{
					EventID:   int32(EventRegistered),
					AppuserID: commonData.User.AppuserID,
				}
				for i, id := range ids {
					addPatientRegisteredEventsParams.PatientID = append(addPatientRegisteredEventsParams.PatientID, id)
					addPatientRegisteredEventsParams.HomeID = append(addPatientRegisteredEventsParams.HomeID, result.Patients[i].HomeID)
					result.Notes = append(result.Notes, fmt.Sprintf("New patient: %s", PatientURL(id)))
					if result.Patients[i].URL == "" {
						patientsRequiringJournal = append(patientsRequiringJournal, id)
					}
				}
				if err := q.AddPatientRegisteredEvents(ctx, addPatientRegisteredEventsParams); err != nil {
					result.Notes = []string{fmt.Sprintf("Error adding registration events for patients: %v", err)}
					return err
				}
			}
			return nil
		}); err == nil {
			if len(patientsRequiringJournal) > 0 {
				result.Notes = append(result.Notes, fmt.Sprintf("Creating %d journals in the background...", len(patientsRequiringJournal)))
				go server.tryCreateJournals(patientsRequiringJournal)
			}
		}
	}

	if err := server.setCookie(w, r, "import-request", &result); err != nil {
		LogR(r, "setting import-request cookie: %w")
	}

	server.redirect(w, r, "/import")
}

func (server *Server) tryCreateJournals(ids []int32) {
}

func (server *Server) ajaxImportValidateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	result := server.parseImportForm(r)
	if err := server.setCookie(w, r, "import-request", &result); err != nil {
		LogR(r, "setting import-request cookie from AJAX: %w")
	}

	_ = ImportValidation(commonData, result).Render(ctx, w)
}

func (server *Server) parseImportForm(r *http.Request) ImportRequest {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	out := ImportRequest{}

	txt, err := server.getFormValue(r, "txt")
	if err != nil {
		out.Notes = append(out.Notes, err.Error())
		return out
	}
	out.Txt = txt

	lines := SliceToSlice(strings.Split(strings.TrimSpace(txt), "\n"), strings.TrimSpace)
	for i, line := range lines {
		fields := SliceToSlice(strings.Split(line, ","), strings.TrimSpace)
		if n := len(fields); n < 2 || n > 4 {
			out.Notes = append(out.Notes, fmt.Sprintf("line %d: expect 2-4 fields, have %d", i, n))
			return out
		}
		patientName := fields[0]
		if patientName == "" {
			out.Notes = append(out.Notes, fmt.Sprintf("line %d: empty patient name", i))
			return out
		}
		speciesName := fields[1]
		if speciesName == "" {
			out.Notes = append(out.Notes, fmt.Sprintf("line %d: empty species name", i))
			return out
		}
		homeName := ""
		if len(fields) >= 3 {
			homeName = fields[2]
		} else {
			homeName = commonData.User.PreferredHome.Name
			if homeName == "" {
				out.Notes = append(out.Notes, fmt.Sprintf("line %d: no home name set, and you don't have a preferred home set", i))
				return out
			}
		}
		url := ""
		if len(fields) >= 4 {
			url = journalRegex.FindString(fields[3])
			if url == "" {
				out.Notes = append(out.Notes, fmt.Sprintf("line %d: '%s' doesn't seem like a journal URL", i, fields[3]))
			}
		}

		homes, err := server.Queries.GetHomeByName(ctx, homeName)
		if err != nil {
			out.Notes = append(out.Notes, fmt.Sprintf("line %d: no home named '%s'", i, homeName))
			return out
		}
		if len(homes) != 1 {
			out.Notes = append(out.Notes, fmt.Sprintf("line %d: expected exactly 1 home named '%s', got %d", i, homeName, len(homes)))
			return out
		}
		homeID := homes[0].ID
		species, err := server.Queries.GetSpeciesByName(ctx, speciesName)
		if err != nil {
			out.Notes = append(out.Notes, fmt.Sprintf("line %d: no species named '%s'", i, speciesName))
			return out
		}
		// NOTE: this will fail if the name is the same in multiple languages
		if len(species) != 1 {
			out.Notes = append(out.Notes, fmt.Sprintf("line %d: expected exactly 1 species named '%s', got %d", i, speciesName, len(species)))
			return out
		}
		speciesID := species[0]

		if rows, err := server.Queries.GetCurrentPatientsForHome(ctx, GetCurrentPatientsForHomeParams{
			CurrHomeID: pgtype.Int4{Int32: homeID, Valid: true},
			LanguageID: commonData.Lang32(),
		}); err == nil {
			for _, row := range rows {
				if row.Name == patientName && row.SpeciesID == speciesID {
					out.Notes = append(out.Notes, fmt.Sprintf("line %d: %s already has a %s named %s", i, homeName, speciesName, patientName))
					return out
				}
			}
		}

		if url == "" {
			out.Notes = append(out.Notes, fmt.Sprintf("line %d: will create a new journal for %s", i, patientName))
		} else {
			if patients, err := server.Queries.GetPatientsByJournalURL(ctx, url); err == nil && len(patients) > 0 {
				out.Notes = append(out.Notes, fmt.Sprintf("line %d: there is already a registered patient with this journal URL: %s", i, PatientURL(patients[0])))
				return out
			}
			out.Notes = append(out.Notes, fmt.Sprintf("line %d: will attach journal URL '%s' to %s", i, url, patientName))
		}

		out.Notes = append(out.Notes, fmt.Sprintf("line %d: will create patient named '%s' with species ID=%d, home ID=%d", i, patientName, speciesID, homeID))

		out.Patients = append(out.Patients, ImportPatient{
			Name:      patientName,
			SpeciesID: speciesID,
			HomeID:    homeID,
			URL:       url,
		})
	}

	out.OK = true
	return out
}
