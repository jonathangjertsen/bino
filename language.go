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
	ID       LanguageID
	Emoji    string
	SelfName string
	Weekdays map[time.Weekday]string
	Months   map[time.Month]string

	AdminDefaultIncludeTag string
	AdminDisplayName       string
	AdminEmailAddress      string
	AdminManageEvents      string
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
	DashboardCheckOut         string
	DashboardSelectHome       string
	DashboardSelectCheckout   string
	DashboardSelectTag        string
	DashboardSelectSpecies    string

	ErrorPageHead         string
	ErrorPageInstructions string

	FooterPrivacy    string
	FooterSourceCode string

	GenericAdd     string
	GenericAge     string
	GenericDelete  string
	GenericDetails string
	GenericHome    string
	GenericJournal string
	GenericLatin   string
	GenericMove    string
	GenericMoveTo  string
	GenericNone    string
	GenericNote    string
	GenericSpecies string
	GenericStatus  string
	GenericTags    string
	GenericUpdate  string

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
	AdminDefaultIncludeTag: "Vis ved innsjekk",
	AdminDisplayName:       "Navn",
	AdminEmailAddress:      "Epostaddresse",
	AdminManageEvents:      "Konfigurer hendelsestyper",
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
	DashboardGoToJournal:      "Gå til pasientjournal",
	DashboardCheckOut:         "Sjekk ut",
	DashboardSelectHome:       "Velg rehabhjem",
	DashboardSelectCheckout:   "Velg status",
	DashboardSelectTag:        "Velg tagg",
	DashboardSelectSpecies:    "Velg art",

	ErrorPageHead:         "Feilmelding",
	ErrorPageInstructions: "Det skjedde noe feil under lasting av siden. Feilen har blitt logget og vil bli undersøkt. Send melding til administrator for hjelp. Den tekniske feilmeldingen følger under.",

	FooterPrivacy:    "Personvern",
	FooterSourceCode: "Kildekode",

	GenericAdd:     "Legg til",
	GenericAge:     "Alder",
	GenericDelete:  "Slett",
	GenericDetails: "Detaljer",
	GenericHome:    "Rehabhjem",
	GenericJournal: "Journal",
	GenericLatin:   "Latin",
	GenericMove:    "Flytt",
	GenericMoveTo:  "Flytt til",
	GenericNone:    "Ingen",
	GenericNote:    "Notis",
	GenericSpecies: "Art",
	GenericStatus:  "Status",
	GenericTags:    "Tagger",
	GenericUpdate:  "Oppdater",

	HomesAddToHome:         "Legg til",
	HomesArchiveHome:       "Arkiver rehabhjem",
	HomesCreateHome:        "Opprett nytt rehabhjem",
	HomesCreateHomeNote:    "Navnet er som regel navnet på en person, men det kan være flere personer i ett rehabhjem, og en person kan være del av flere rehabhjem.",
	HomesEmptyHome:         "Det er ingen brukere i dette rehabhjemmet.",
	HomesHomeName:          "Rehabhjem",
	HomesRemoveFromCurrent: "Fjern fra dette rehabhjemmet",
	HomesUnassignedUsers:   "Brukere som ikke er koblet til noe rehabhjem",
	HomesViewHomes:         "Rehabhjem",

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
	AdminDefaultIncludeTag: "Show at check-in",
	AdminDisplayName:       "Name",
	AdminEmailAddress:      "Email address",
	AdminManageEvents:      "Manage event types",
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
	DashboardGoToJournal:      "Go to patient journal",
	DashboardCheckOut:         "Checkout",
	DashboardSelectHome:       "Select home",
	DashboardSelectCheckout:   "Select status",
	DashboardSelectTag:        "Select tag",
	DashboardSelectSpecies:    "Select species",

	ErrorPageHead:         "Error",
	ErrorPageInstructions: "An error occurred while loading the page. The error has been logged and will be investigated. Send a message to the site admin for help. The technical error message is as follows.",

	FooterPrivacy:    "Privacy",
	FooterSourceCode: "Source code",

	GenericAdd:     "Add",
	GenericAge:     "Age",
	GenericDelete:  "Delete",
	GenericDetails: "Details",
	GenericHome:    "Home",
	GenericJournal: "Journal",
	GenericLatin:   "Latin",
	GenericMove:    "Move",
	GenericMoveTo:  "Move to",
	GenericNone:    "None",
	GenericNote:    "Note",
	GenericSpecies: "Species",
	GenericStatus:  "Status",
	GenericTags:    "Tags",
	GenericUpdate:  "Update",

	HomesAddToHome:         "Add",
	HomesArchiveHome:       "Archive rehab home",
	HomesCreateHome:        "Create new rehab home",
	HomesCreateHomeNote:    "The name is usually that of a person, but there can be multiple people in a rehab home, and one person can be associated with several rehab homes.",
	HomesEmptyHome:         "There are no users in this rehab home.",
	HomesHomeName:          "Name",
	HomesRemoveFromCurrent: "Remove from this rehab home",
	HomesUnassignedUsers:   "Users that are not associated with any rehab homes",
	HomesViewHomes:         "Rehab homes",

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
