//go:generate go tool go-enum --no-iota --values
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// ENUM(
//
//	NO = 1,
//	EN = 2,
//
// )
type LanguageID int32

type Language struct {
	ID          LanguageID
	Emoji       string
	SelfName    string
	Weekdays    map[time.Weekday]string
	Months      map[time.Month]string
	GDriveRoles map[string]string

	AdminDefaultIncludeTag string
	AdminDisplayName       string
	AdminEmailAddress      string
	AdminInviteToBino      string
	AdminManageEvents      string
	AdminManageGoogleDrive string
	AdminManageHomes       string
	AdminManageSpecies     string
	AdminManageStatuses    string
	AdminManageTags        string
	AdminRoot              string

	AuthLogOut string

	CheckinCheckInPatient  string
	CheckinIHaveThePatient string
	CheckinPatientName     string
	CheckinYouAreHomeless  string

	DashboardNoPatientsInHome string
	DashboardGoToJournal      string
	DashboardGoToPatientPage  string
	DashboardCheckOut         string
	DashboardSelectHome       string
	DashboardSelectCheckout   string
	DashboardSelectTag        string
	DashboardSelectSpecies    string

	ErrorPageHead         string
	ErrorPageInstructions string
	ErrorSettingLanguage  string

	FooterPrivacy    string
	FooterSourceCode string

	GDriveBaseDir                          string
	GDriveSelectFolder                     string
	GDriveSelectFolderInstruction          string
	GDriveSelectedFolder                   string
	GDriveReloadFolders                    string
	GDriveTemplate                         string
	GDriveSelectTemplate                   string
	GDriveSelectedTemplate                 string
	GDriveTemplateInstruction              string
	GDriveReloadDocs                       string
	GDriveReloadDocsFilter                 string
	GDriveNoBaseDirSelected                string
	GDriveNoTemplateSelected               string
	GDrivePermissionsForBaseDir            string
	GDrivePermissionsForBaseDirInstruction string
	GDriveDisplayName                      string
	GDriveEmail                            string
	GDriveRole                             string
	GDriveFoundBinoUser                    string
	GDriveBinoUsersMissingWritePermission  string
	GDriveEmailInBino                      string
	GDriveGiveAccess                       string
	GDriveLoadFoldersFailed                string
	GDriveBaseDirUpdated                   string
	GDriveTemplateUpdated                  string
	GDriveUserInvited                      string
	GDriveCreateJournalForPatient          string
	GDriveSelectExistingJournalForPatient  string
	GDriveNoJournalForPatient              string

	GenericAdd      string
	GenericAge      string
	GenericDelete   string
	GenericDetails  string
	GenericGoBack   string
	GenericHome     string
	GenericJournal  string
	GenericLatin    string
	GenericMove     string
	GenericMoveTo   string
	GenericName     string
	GenericNone     string
	GenericNote     string
	GenericNotFound string
	GenericSpecies  string
	GenericStatus   string
	GenericTags     string
	GenericUpdate   string

	HomesArchiveHome       string
	HomesAddToHome         string
	HomesAddUserToHome     string
	HomesCreateHome        string
	HomesCreateHomeNote    string
	HomesEmptyHome         string
	HomesHomeName          string
	HomesRemoveFromCurrent string
	HomesViewHomes         string
	HomesUnassignedUsers   string

	LanguageUpdateFailed string
	LanguageUpdateOK     string

	NotFoundPageHead         string
	NotFoundPageInstructions string

	PatientRegisteredTime string
	PatientCheckedOutTime string
	PatientEventTime      string
	PatientEventEvent     string
	PatientEventNote      string
	PatientEventUser      string
	PatientEventHome      string

	NavbarCalendar  string
	NavbarDashboard string

	Status map[Status]string
	Event  map[Event]string
}

func (l *Language) TODO(s string) string {
	return fmt.Sprintf("TODO[%s]", s)
}

var NO = &Language{
	ID:       LanguageIDNO,
	Emoji:    "🇳🇴",
	SelfName: "Norsk",
	Weekdays: map[time.Weekday]string{
		time.Monday:    "mandag",
		time.Tuesday:   "tirsdag",
		time.Wednesday: "onsdag",
		time.Thursday:  "torsdag",
		time.Friday:    "fredag",
		time.Saturday:  "lørdag",
		time.Sunday:    "søndag",
	},
	Months: map[time.Month]string{
		time.January:   "januar",
		time.February:  "februar",
		time.March:     "mars",
		time.April:     "april",
		time.May:       "mai",
		time.June:      "juni",
		time.July:      "juli",
		time.August:    "august",
		time.September: "september",
		time.October:   "oktober",
		time.November:  "november",
		time.December:  "desember",
	},
	GDriveRoles: map[string]string{
		"owner":         "Full tilgang (eier)",
		"organizer":     "Full tilgang, kan endre tilganger",
		"fileOrganizer": "Full tilgang til innhold",
		"writer":        "Kan opprette og redigere journaler",
		"commenter":     "Kan kommentere på journaler",
		"reader":        "Kan lese journaler",
	},
	AdminDefaultIncludeTag: "Vis ved innsjekk",
	AdminDisplayName:       "Navn",
	AdminEmailAddress:      "Epostaddresse",
	AdminInviteToBino:      "Inviter til Bino",
	AdminManageEvents:      "Konfigurer hendelsestyper",
	AdminManageGoogleDrive: "Konfigurer Google Drive",
	AdminManageHomes:       "Konfigurer rehabhjem",
	AdminManageSpecies:     "Konfigurer arter",
	AdminManageStatuses:    "Konfigurer statuser",
	AdminManageTags:        "Konfigurer tagger",
	AdminRoot:              "Konfigurering",

	AuthLogOut: "Logg ut",

	CheckinCheckInPatient:  "Sjekk inn pasient",
	CheckinIHaveThePatient: "Pasienten er her",
	CheckinPatientName:     "Pasientens navn",
	CheckinYouAreHomeless:  "Du kan ikke sjekke inn pasienter ennå fordi du ikke er koblet til et rehabhjem.",

	DashboardNoPatientsInHome: "Ingen pasienter",
	DashboardGoToJournal:      "Gå til pasientjournal i Google Drive",
	DashboardGoToPatientPage:  "Gå til pasientside",
	DashboardCheckOut:         "Sjekk ut",
	DashboardSelectHome:       "Velg rehabhjem",
	DashboardSelectCheckout:   "Velg status",
	DashboardSelectTag:        "Velg tagg",
	DashboardSelectSpecies:    "Velg art",

	ErrorPageHead:         "Feilmelding",
	ErrorPageInstructions: "Det skjedde noe feil under lasting av siden. Feilen har blitt logget og vil bli undersøkt. Send melding til administrator for hjelp. Den tekniske feilmeldingen følger under.",
	ErrorSettingLanguage:  "Kunne ikke oppdatere språk",

	FooterPrivacy:    "Personvern",
	FooterSourceCode: "Kildekode",

	GDriveBaseDir:                          "Velg journalmappe",
	GDriveSelectFolder:                     "Velg mappe",
	GDriveSelectedFolder:                   "Valgt",
	GDriveSelectFolderInstruction:          "Her kan du velge hvilken mappe nye journaler skal opprettes i.",
	GDriveReloadFolders:                    "Hent mapper fra Google Drive på nytt",
	GDriveTemplate:                         "Velg mal",
	GDriveTemplateInstruction:              "Her kan du velge hvilket dokument som skal brukes som mal. Når du henter dokumenter bør du skrive et søkeord, f.eks 'Navn' hvis tittelen på malen inneholder det ordet. Ellers vil det komme opp veldig mange dokumenter.",
	GDriveSelectTemplate:                   "Velg som mal",
	GDriveSelectedTemplate:                 "Valgt",
	GDriveReloadDocs:                       "Hent dokumenter fra Google Drive",
	GDriveReloadDocsFilter:                 "Søkeord",
	GDriveNoBaseDirSelected:                "Det er ikke valgt noen journalmappe",
	GDriveNoTemplateSelected:               "Det er ikke valgt noen mal",
	GDrivePermissionsForBaseDir:            "Tilganger til journalmappen",
	GDrivePermissionsForBaseDirInstruction: "Her kan du se hvem som har tilgang til journalmappen, og sammenligne med tilganger i Bino.",
	GDriveDisplayName:                      "Brukernavn i Google Drive",
	GDriveEmail:                            "Email",
	GDriveRole:                             "Tilganger",
	GDriveFoundBinoUser:                    "Bino-konto",
	GDriveBinoUsersMissingWritePermission:  "Disse brukerne mangler tilgang til å opprette journaler i den valgte mappen:",
	GDriveEmailInBino:                      "Email i Bino",
	GDriveGiveAccess:                       "Gi skrivetilgang",
	GDriveLoadFoldersFailed:                "Kunne ikke laste inn mapper fra Google Drive",
	GDriveBaseDirUpdated:                   "Journalmappen ble oppdatert. Husk å også velge mal.",
	GDriveTemplateUpdated:                  "Malen ble oppdatert.",
	GDriveUserInvited:                      "Brukeren ble invitert til mappen",
	GDriveCreateJournalForPatient:          "Opprett pasientjournal",
	GDriveSelectExistingJournalForPatient:  "Koble en eksisterende journal til pasienten",
	GDriveNoJournalForPatient:              "Det er ikke koblet noen journal til pasienten.",

	GenericAdd:      "Legg til",
	GenericAge:      "Alder",
	GenericDelete:   "Slett",
	GenericDetails:  "Detaljer",
	GenericGoBack:   "Tilbake",
	GenericHome:     "Rehabhjem",
	GenericJournal:  "Journal",
	GenericLatin:    "Latin",
	GenericMove:     "Flytt",
	GenericMoveTo:   "Flytt til",
	GenericName:     "Navn",
	GenericNone:     "Ingen",
	GenericNote:     "Notis",
	GenericNotFound: "Ikke funnet",
	GenericSpecies:  "Art",
	GenericStatus:   "Status",
	GenericTags:     "Tagger",
	GenericUpdate:   "Oppdater",

	HomesAddToHome:         "Legg til",
	HomesArchiveHome:       "Arkiver rehabhjem",
	HomesCreateHome:        "Opprett nytt rehabhjem",
	HomesCreateHomeNote:    "Navnet er som regel navnet på en person, men det kan være flere personer i ett rehabhjem, og en person kan være del av flere rehabhjem.",
	HomesEmptyHome:         "Det er ingen brukere i dette rehabhjemmet.",
	HomesHomeName:          "Rehabhjem",
	HomesRemoveFromCurrent: "Fjern fra dette rehabhjemmet",
	HomesUnassignedUsers:   "Brukere som ikke er koblet til noe rehabhjem",
	HomesViewHomes:         "Rehabhjem",

	LanguageUpdateOK:     "Oppdaterte språk",
	LanguageUpdateFailed: "Kunne ikke oppdatere språk",

	NotFoundPageHead:         "Ikke funnet",
	NotFoundPageInstructions: "Siden ble ikke funnet. Se feilmelding:",

	NavbarCalendar:  "Kalender",
	NavbarDashboard: "Hovedside",

	PatientRegisteredTime: "Registrert",
	PatientCheckedOutTime: "Sjekket ut",
	PatientEventTime:      "Tidspunkt",
	PatientEventEvent:     "Hendelse",
	PatientEventNote:      "Notis",
	PatientEventUser:      "Endret av",
	PatientEventHome:      "Rehabhjem",

	Status: map[Status]string{
		StatusUnknown:                        "Ukjent",
		StatusPendingAdmission:               "Venter på inntak",
		StatusAdmitted:                       "I rehab",
		StatusAdopted:                        "Adoptert",
		StatusReleased:                       "Sluppet fri",
		StatusTransferredOutsideOrganization: "Overført til annet tiltak",
		StatusDead:                           "Død",
		StatusEuthanized:                     "Avlivet",
		StatusDeleted:                        "Slettet",
	},

	Event: map[Event]string{
		EventUnknown:                        "Ukjent",
		EventRegistered:                     "Registrert",
		EventAdmitted:                       "Tatt inn",
		EventAdopted:                        "Adoptert",
		EventReleased:                       "Sluppet fri",
		EventTransferredToOtherHome:         "Overført",
		EventTransferredOutsideOrganization: "Overført til annen organisasjon",
		EventDied:                           "Døde",
		EventEuthanized:                     "Avlivet",
		EventTagAdded:                       "La til tagg",
		EventTagRemoved:                     "Fjernet tagg",
		EventStatusChanged:                  "Endret status",
		EventDeleted:                        "Slettet",
		EventNameChanged:                    "Endret navn",
	},
}

var EN = &Language{
	ID:       LanguageIDEN,
	Emoji:    "🇬🇧",
	SelfName: "English",
	Weekdays: map[time.Weekday]string{
		time.Monday:    time.Monday.String(),
		time.Tuesday:   time.Tuesday.String(),
		time.Wednesday: time.Wednesday.String(),
		time.Thursday:  time.Thursday.String(),
		time.Friday:    time.Friday.String(),
		time.Saturday:  time.Saturday.String(),
		time.Sunday:    time.Sunday.String(),
	},
	Months: map[time.Month]string{
		time.January:   time.January.String(),
		time.February:  time.February.String(),
		time.March:     time.March.String(),
		time.April:     time.April.String(),
		time.May:       time.May.String(),
		time.June:      time.June.String(),
		time.July:      time.July.String(),
		time.August:    time.August.String(),
		time.September: time.September.String(),
		time.October:   time.October.String(),
		time.November:  time.November.String(),
		time.December:  time.December.String(),
	},
	GDriveRoles: map[string]string{
		"owner":         "Owner",
		"organizer":     "Admin, can set permissions",
		"fileOrganizer": "Content administrator",
		"writer":        "Can create and edit journals",
		"commenter":     "Can comment on journals",
		"reader":        "Can read journals",
	},
	AdminDefaultIncludeTag: "Show at check-in",
	AdminDisplayName:       "Name",
	AdminEmailAddress:      "Email address",
	AdminInviteToBino:      "Invite to Bino",
	AdminManageEvents:      "Manage event types",
	AdminManageGoogleDrive: "Configure Google Drive",
	AdminManageHomes:       "Manage rehab homes",
	AdminManageSpecies:     "Manage species",
	AdminManageStatuses:    "Manage statuses",
	AdminManageTags:        "Manage tags",
	AdminRoot:              "Admin",

	AuthLogOut: "Log out",

	CheckinCheckInPatient:  "Check in",
	CheckinIHaveThePatient: "The patient is here",
	CheckinPatientName:     "Name of the patient",
	CheckinYouAreHomeless:  "You can't check in patients yet because you're not connected to a rehab home.",

	DashboardNoPatientsInHome: "No patients",
	DashboardGoToJournal:      "Go to patient journal in Google Drive",
	DashboardGoToPatientPage:  "Go to patient page",
	DashboardCheckOut:         "Checkout",
	DashboardSelectHome:       "Select home",
	DashboardSelectCheckout:   "Select status",
	DashboardSelectTag:        "Select tag",
	DashboardSelectSpecies:    "Select species",

	ErrorPageHead:         "Error",
	ErrorPageInstructions: "An error occurred while loading the page. The error has been logged and will be investigated. Send a message to the site admin for help. The technical error message is as follows.",
	ErrorSettingLanguage:  "Failed to update language",

	FooterPrivacy:    "Privacy",
	FooterSourceCode: "Source code",

	GDriveBaseDir:                          "Select journal folder",
	GDriveSelectFolder:                     "Select folder",
	GDriveSelectedFolder:                   "Selected",
	GDriveSelectFolderInstruction:          "Select the folder in which new patient journals will be created.",
	GDriveReloadFolders:                    "Reload folders from Google Drive",
	GDriveTemplate:                         "Choose template",
	GDriveTemplateInstruction:              "Select the document that will be used as a template. When reloading documents you should select a search filter, e.g., 'Name' if the template contains that word in the title. That way you don't need to look through all the documents.",
	GDriveSelectTemplate:                   "Select as template",
	GDriveSelectedTemplate:                 "Selected",
	GDriveReloadDocs:                       "Load documents",
	GDriveReloadDocsFilter:                 "Search filter",
	GDriveNoBaseDirSelected:                "No base folder selected",
	GDriveNoTemplateSelected:               "No template document selected",
	GDrivePermissionsForBaseDir:            "Journal folder permissions",
	GDrivePermissionsForBaseDirInstruction: "Check who has permissions to the folder, and compare with the permissions in Bino.",
	GDriveDisplayName:                      "Username in Google Drive",
	GDriveEmail:                            "Email",
	GDriveRole:                             "Role",
	GDriveFoundBinoUser:                    "Bino account",
	GDriveBinoUsersMissingWritePermission:  "These users do not have access to create new journals in the selected folder:",
	GDriveEmailInBino:                      "Email address in Bino",
	GDriveGiveAccess:                       "Give write-access",
	GDriveLoadFoldersFailed:                "Failed to load folders from Google Drive",
	GDriveBaseDirUpdated:                   "Google Drive journal folder was updated. Remember to also update the template.",
	GDriveTemplateUpdated:                  "Template journal was updated",
	GDriveUserInvited:                      "The user was invited to the journal folder",
	GDriveCreateJournalForPatient:          "Create a new journal for the patient",
	GDriveSelectExistingJournalForPatient:  "Connect an existing journal to the patient",
	GDriveNoJournalForPatient:              "No journal found",

	GenericAdd:      "Add",
	GenericAge:      "Age",
	GenericDelete:   "Delete",
	GenericDetails:  "Details",
	GenericGoBack:   "Go back",
	GenericHome:     "Home",
	GenericJournal:  "Journal",
	GenericLatin:    "Latin",
	GenericMove:     "Move",
	GenericMoveTo:   "Move to",
	GenericName:     "Name",
	GenericNone:     "None",
	GenericNote:     "Note",
	GenericNotFound: "Not found",
	GenericSpecies:  "Species",
	GenericStatus:   "Status",
	GenericTags:     "Tags",
	GenericUpdate:   "Update",

	HomesAddToHome:         "Add",
	HomesArchiveHome:       "Archive rehab home",
	HomesCreateHome:        "Create new rehab home",
	HomesCreateHomeNote:    "The name is usually that of a person, but there can be multiple people in a rehab home, and one person can be associated with several rehab homes.",
	HomesEmptyHome:         "There are no users in this rehab home.",
	HomesHomeName:          "Name",
	HomesRemoveFromCurrent: "Remove from this rehab home",
	HomesUnassignedUsers:   "Users that are not associated with any rehab homes",
	HomesViewHomes:         "Rehab homes",

	LanguageUpdateFailed: "Failed to update language",
	LanguageUpdateOK:     "Updated language",

	NavbarCalendar:  "Calendar",
	NavbarDashboard: "Dashboard",

	PatientRegisteredTime: "Registrert",
	PatientCheckedOutTime: "Checked out",
	PatientEventTime:      "Time",
	PatientEventEvent:     "Event",
	PatientEventNote:      "Note",
	PatientEventUser:      "User",
	PatientEventHome:      "Home",

	Status: map[Status]string{
		StatusUnknown:                        "Unknown",
		StatusPendingAdmission:               "Pending admission",
		StatusAdmitted:                       "In rehab",
		StatusAdopted:                        "Adopted",
		StatusReleased:                       "Released",
		StatusTransferredOutsideOrganization: "Transferred outside organization",
		StatusDead:                           "Dead",
		StatusEuthanized:                     "Euthanized",
		StatusDeleted:                        "Deleted",
	},

	Event: map[Event]string{
		EventUnknown:                        "Unknown",
		EventRegistered:                     "Registered",
		EventAdmitted:                       "Admitted",
		EventAdopted:                        "Adopted",
		EventReleased:                       "Released",
		EventTransferredToOtherHome:         "Transferred",
		EventTransferredOutsideOrganization: "Transferred outside of organisation",
		EventDied:                           "Died",
		EventEuthanized:                     "Euthanized",
		EventTagAdded:                       "Tag added",
		EventTagRemoved:                     "Tag removed",
		EventStatusChanged:                  "Status changed",
		EventDeleted:                        "Deleted",
		EventNameChanged:                    "Name changed",
	},
}

func (l *Language) FormatEvent(ctx context.Context, e int32, assocID pgtype.Int4, server *Server) string {
	event := Event(e)

	switch event {
	case EventTagAdded:
		if tagName, err := server.Queries.GetTagName(ctx, GetTagNameParams{
			LanguageID: int32(l.ID),
			TagID:      assocID.Int32,
		}); err == nil {
			return l.formatTagAdded(tagName)
		}
	case EventTagRemoved:
		if tagName, err := server.Queries.GetTagName(ctx, GetTagNameParams{
			LanguageID: int32(l.ID),
			TagID:      assocID.Int32,
		}); err == nil {
			return l.formatTagRemoved(tagName)
		}
	case EventStatusChanged:
		return l.formatStatusChanged(Status(assocID.Int32))
	default:
		if str, ok := l.Event[event]; ok {
			return str
		}
	}
	return event.String()
}

func (l *Language) formatTagAdded(tagName string) string {
	switch l.ID {
	case LanguageIDNO:
		return fmt.Sprintf("Tagget som '%s'", tagName)
	case LanguageIDEN:
		return fmt.Sprintf("Tagged as '%s'", tagName)
	default:
		return tagName
	}
}

func (l *Language) formatTagRemoved(tagName string) string {
	switch l.ID {
	case LanguageIDNO:
		return fmt.Sprintf("Fjernet taggen '%s'", tagName)
	case LanguageIDEN:
		return fmt.Sprintf("Removed tag '%s'", tagName)
	default:
		return tagName
	}
}

func (l *Language) formatStatusChanged(status Status) string {
	switch l.ID {
	case LanguageIDNO:
		return fmt.Sprintf("Endret status til '%s'", status)
	case LanguageIDEN:
		return fmt.Sprintf("Changed status to '%s'", status)
	default:
		return status.String()
	}
}

func (l *Language) FormatTimeAbs(t time.Time) string {
	switch l.ID {
	case LanguageIDNO:
		return fmt.Sprintf("%d. %s %d kl. %02d:%02d",
			t.Day(),
			l.Months[t.Month()],
			t.Year(),
			t.Hour(),
			t.Minute(),
		)
	case LanguageIDEN:
		return t.Format("January 2, 2006 at 3:04 PM")
	default:
		return t.String()
	}
}

func (l *Language) FormatTimeRel(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch l.ID {
	case LanguageIDNO:
		if diff < time.Minute {
			return "akkurat nå"
		} else if diff < time.Hour {
			return fmt.Sprintf("for %d minutter siden", int(diff.Minutes()))
		} else if diff < 24*time.Hour {
			return fmt.Sprintf("for %d timer siden", int(diff.Hours()))
		}
		if now.Year() == t.Year() && now.YearDay()-t.YearDay() < 7 {
			return fmt.Sprintf("%s kl. %02d:%02d", l.Weekdays[t.Weekday()], t.Hour(), t.Minute())
		}
	case LanguageIDEN:
		if diff < time.Minute {
			return "just now"
		} else if diff < time.Hour {
			return fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
		} else if diff < 24*time.Hour {
			return fmt.Sprintf("%d hours ago", int(diff.Hours()))
		}
		if now.Year() == t.Year() && now.YearDay()-t.YearDay() < 7 {
			return t.Format("Monday at 3:04 PM")
		}
	}
	return l.FormatTimeAbs(t)
}

var Languages = map[int32]*Language{
	int32(LanguageIDNO): NO,
	int32(LanguageIDEN): EN,
}

func GetLanguage(id int32) *Language {
	lang, ok := Languages[id]
	if !ok {
		return EN
	}
	return lang
}
