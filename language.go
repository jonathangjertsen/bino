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

	AdminDefaultIncludeTag      string
	AdminDisplayName            string
	AdminEmailAddress           string
	AdminInviteToBino           string
	AdminManageEvents           string
	AdminManageGoogleDrive      string
	AdminManageHomes            string
	AdminManageSpecies          string
	AdminManageStatuses         string
	AdminManageTags             string
	AdminManageInvites          string
	AdminManageUsers            string
	AdminScrubUserData          string
	AdminScrubUserDataConfirm   string
	AdminNukeUser               string
	AdminNukeUserConfirm        string
	AdminAbortedDueToWrongEmail string
	AdminUserDeletionFailed     string
	AdminUserWasDeleted         string
	AdminExistingUsers          string
	AdminInviteUsers            string
	AdminInviteExpires          string
	AdminPendingInvitations     string
	AdminInvitationFailed       string
	AdminInvitationOKNoEmail    string
	AdminInviteCode             string
	AdminRoot                   string

	AuthLogOut string

	CheckinCheckInPatient  string
	CheckinIHaveThePatient string
	CheckinPatientName     string
	CheckinYouAreHomeless  string

	DashboardNoPatientsInHome      string
	DashboardGoToJournal           string
	DashboardGoToPatientPage       string
	DashboardCheckOut              string
	DashboardSearch                string
	DashboardSearchFilter          string
	DashboardSearchShowMine        string
	DashboardSearchShowFull        string
	DashboardSearchShowUnavailable string
	DashboardSelectHome            string
	DashboardSelectCheckout        string
	DashboardSelectTag             string
	DashboardSelectSpecies         string
	DashboardNonPreferredSpecies   string
	DashboardOtherHome             string

	ErrorPageHead         string
	ErrorPageInstructions string
	ErrorSettingLanguage  string

	FooterPrivacy    string
	FooterSourceCode string

	FormerPatients string

	GDriveBaseDir                          string
	GDriveSelectFolder                     string
	GDriveSelectFolderInstruction          string
	GDriveSelectedFolder                   string
	GDriveReloadFolders                    string
	GDriveTemplate                         string
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
	GDriveSelectExistingJournalInstruction string
	GDriveNoJournalForPatient              string
	GDriveCreateJournalFailed              string

	GenericAdd      string
	GenericAge      string
	GenericAvatar   string
	GenericCancel   string
	GenericConfirm  string
	GenericDelete   string
	GenericDetails  string
	GenericEmail    string
	GenericFrom     string
	GenericTo       string
	GenericGoBack   string
	GenericHome     string
	GenericJournal  string
	GenericLatin    string
	GenericMove     string
	GenericMoveTo   string
	GenericName     string
	GenericNone     string
	GenericNever    string
	GenericNote     string
	GenericNotFound string
	GenericSpecies  string
	GenericStatus   string
	GenericTags     string
	GenericUpdate   string
	GenericFailed   string
	GenericSuccess  string
	GenericMessage  string
	GenericURL      string

	HomesArchiveHome               string
	HomesAddToHome                 string
	HomesAddUserToHome             string
	HomesCreateHome                string
	HomesCreateHomeNote            string
	HomesEmptyHome                 string
	HomesHomeName                  string
	HomesRemoveFromCurrent         string
	HomesViewHomes                 string
	HomesUnassignedUsers           string
	HomesPatients                  string
	HomesUsers                     string
	HomeCapacity                   string
	HomePreferredSpecies           string
	HomeAvailability               string
	HomePeriodInvalid              string
	HomeAvailableIndefinitely      string
	HomeUnavailableIndefinitely    string
	HomeUnavailableFromInstruction string
	HomeUnavailableToInstruction   string

	LanguageUpdateFailed string

	NotFoundPageHead         string
	NotFoundPageInstructions string

	PatientRegisteredTime string
	PatientCheckedOutTime string
	PatientEventTime      string
	PatientEventEvent     string
	PatientEventNote      string
	PatientEventUser      string
	PatientEventHome      string

	UserHomes      string
	UserIsHomeless string

	NavbarCalendar  string
	NavbarDashboard string

	Status map[Status]string
	Event  map[Event]string
}

func (l *Language) HomeUnavailableUntil(dv DateView) string {
	switch l.ID {
	case LanguageIDNO:
		if dv.Year > time.Now().Year()+2 {
			return fmt.Sprintf("Utilgjengelig på ubestemt tid.")
		}
		return fmt.Sprintf("Utilgjengelig til og med den %d. %s %d.", dv.Day, l.Months[dv.Month], dv.Year)
	case LanguageIDEN:
		fallthrough
	default:
		if dv.Year > time.Now().Year()+2 {
			return fmt.Sprintf("Unavailable until further notice.")
		}
		return fmt.Sprintf("Unavailable until %s %d %d.", l.Months[dv.Month], dv.Day, dv.Year)
	}
}

func (l *Language) HomeAvailableUntil(dv DateView) string {
	switch l.ID {
	case LanguageIDNO:
		return fmt.Sprintf("Blir utilgjengelig den %d. %s %d.", dv.Day, l.Months[dv.Month], dv.Year)
	case LanguageIDEN:
		fallthrough
	default:
		return fmt.Sprintf("Becomes unavailable from %s %d %d.", l.Months[dv.Month], dv.Day, dv.Year)
	}
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
	AdminDefaultIncludeTag:      "Vis ved innsjekk",
	AdminDisplayName:            "Navn",
	AdminEmailAddress:           "Epostaddresse",
	AdminInviteToBino:           "Inviter til Bino",
	AdminManageEvents:           "Konfigurer hendelsestyper",
	AdminManageGoogleDrive:      "Konfigurer Google Drive",
	AdminManageHomes:            "Konfigurer rehabhjem",
	AdminManageSpecies:          "Konfigurer arter",
	AdminManageStatuses:         "Konfigurer statuser",
	AdminManageTags:             "Konfigurer tagger",
	AdminManageInvites:          "Administrer invitasjoner",
	AdminManageUsers:            "Administrer brukere",
	AdminScrubUserData:          "Slett brukerdata",
	AdminScrubUserDataConfirm:   "Skriv inn brukerens email-addresse for å bekrefte at du vil slette brukerdataen",
	AdminNukeUser:               "Tilintetgjør bruker",
	AdminNukeUserConfirm:        "Skriv inn brukerens email-addresse for å bekrefte at du vil tilintetgjøre brukeren (dette sletter også alt innhold brukeren har opprettet)",
	AdminAbortedDueToWrongEmail: "Feil email-addresse innskrevet. Handlingen ble avbrutt.",
	AdminUserDeletionFailed:     "Kunne ikke slette brukeren. Kontakt administrator.",
	AdminUserWasDeleted:         "Brukeren ble slettet.",
	AdminExistingUsers:          "Brukere i Bino",
	AdminInviteUsers:            "Inviter brukere",
	AdminInviteExpires:          "Utløper",
	AdminPendingInvitations:     "Utsendte invitasjoner",
	AdminInvitationFailed:       "Kunne ikke invitere brukeren. Kontakt administrator.",
	AdminInvitationOKNoEmail:    "Eposten ble lagt til i listen, men det er ikke sendt ut en epost. Send personen en lenke til forsiden og be dem om å opprette en bruker.",
	AdminInviteCode:             "Kode",
	AdminRoot:                   "Admin",

	AuthLogOut: "Logg ut",

	CheckinCheckInPatient:  "Sjekk inn pasient",
	CheckinIHaveThePatient: "Pasienten er her",
	CheckinPatientName:     "Pasientens navn",
	CheckinYouAreHomeless:  "Du kan ikke sjekke inn pasienter ennå fordi du ikke er koblet til et rehabhjem.",

	DashboardNoPatientsInHome:      "Ingen pasienter",
	DashboardGoToJournal:           "Gå til pasientjournal i Google Drive",
	DashboardGoToPatientPage:       "Gå til pasientside",
	DashboardCheckOut:              "Sjekk ut",
	DashboardSearch:                "Søk i rehabhjem og pasienter",
	DashboardSearchFilter:          "Filtrer rehabhjem",
	DashboardSearchShowMine:        "Vis mitt rehabhjem",
	DashboardSearchShowFull:        "Vis fulle rehabhjem",
	DashboardSearchShowUnavailable: "Vis utilgjengelige rehabhjem",
	DashboardSelectHome:            "Velg rehabhjem",
	DashboardSelectCheckout:        "Velg status",
	DashboardSelectTag:             "Velg tagg",
	DashboardSelectSpecies:         "Velg art",
	DashboardNonPreferredSpecies:   "Andre arter",
	DashboardOtherHome:             "Andre rehabhjem",

	ErrorPageHead:         "Feilmelding",
	ErrorPageInstructions: "Det skjedde noe feil under lasting av siden. Feilen har blitt logget og vil bli undersøkt. Send melding til administrator for hjelp. Den tekniske feilmeldingen følger under.",
	ErrorSettingLanguage:  "Kunne ikke oppdatere språk",

	FooterPrivacy:    "Personvern",
	FooterSourceCode: "Kildekode",

	FormerPatients: "Tidligere pasienter",

	GDriveBaseDir:                          "Journalmappe",
	GDriveSelectFolder:                     "Velg mappe",
	GDriveSelectedFolder:                   "Valgt",
	GDriveSelectFolderInstruction:          "Her kan du velge hvilken mappe nye journaler skal opprettes i.",
	GDriveReloadFolders:                    "Hent mapper fra Google Drive på nytt",
	GDriveTemplate:                         "Mal",
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
	GDriveUserInvited:                      "Brukeren ble invitert til mappen",
	GDriveCreateJournalForPatient:          "Opprett pasientjournal i Google Drive",
	GDriveSelectExistingJournalInstruction: "Eller velg en eksisterende journal i Google Drive:",
	GDriveNoJournalForPatient:              "Det er ikke koblet noen journal til pasienten.",
	GDriveCreateJournalFailed:              "Pasienten ble lagt til, men kunne ikke opprette pasientjournal i Google Drive",

	GenericAdd:      "Legg til",
	GenericAge:      "Alder",
	GenericAvatar:   "Profilbilde",
	GenericCancel:   "Avbryt",
	GenericConfirm:  "Bekreft",
	GenericDelete:   "Slett",
	GenericDetails:  "Detaljer",
	GenericEmail:    "Email",
	GenericFrom:     "Fra",
	GenericTo:       "Til",
	GenericGoBack:   "Tilbake",
	GenericHome:     "Rehabhjem",
	GenericJournal:  "Journal",
	GenericLatin:    "Latin",
	GenericMove:     "Flytt",
	GenericMoveTo:   "Flytt til",
	GenericName:     "Navn",
	GenericNone:     "Ingen",
	GenericNever:    "Aldri",
	GenericNote:     "Notis",
	GenericNotFound: "Ikke funnet",
	GenericSpecies:  "Art",
	GenericStatus:   "Status",
	GenericTags:     "Tagger",
	GenericUpdate:   "Oppdater",
	GenericFailed:   "Noe gikk galt. Kontakt administrator.",
	GenericSuccess:  "Handlingen ble utført.",
	GenericMessage:  "Melding",
	GenericURL:      "URL",

	HomesAddToHome:                 "Legg til",
	HomesArchiveHome:               "Arkiver rehabhjem",
	HomesCreateHome:                "Opprett nytt rehabhjem",
	HomesCreateHomeNote:            "Navnet er som regel navnet på en person, men det kan være flere personer i ett rehabhjem, og en person kan være del av flere rehabhjem.",
	HomesEmptyHome:                 "Det er ingen brukere i dette rehabhjemmet.",
	HomesHomeName:                  "Rehabhjem",
	HomesRemoveFromCurrent:         "Fjern fra dette rehabhjemmet",
	HomesUnassignedUsers:           "Brukere som ikke er koblet til noe rehabhjem",
	HomesViewHomes:                 "Rehabhjem",
	HomesPatients:                  "Pasienter",
	HomesUsers:                     "Brukere",
	HomeCapacity:                   "Kapasitet",
	HomePreferredSpecies:           "Favoritter",
	HomeAvailability:               "Tilgjengelighet",
	HomePeriodInvalid:              "Perioden er ugyldig.",
	HomeAvailableIndefinitely:      "Tilgjengelig.",
	HomeUnavailableIndefinitely:    "Utilgjengelig på ubestemt tid.",
	HomeUnavailableFromInstruction: "Datoen du blir utilgjengelig.",
	HomeUnavailableToInstruction:   "Siste dato du er utilgjengelig.",

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

	UserHomes:      "Tilkoblede rehabhjem",
	UserIsHomeless: "Ingen tilkoblede rehabhjem",

	Status: map[Status]string{
		StatusUnknown:                        "Ukjent",
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
		EventJournalCreated:                 "Opprettet journal i Google Drive",
		EventJournalAttached:                "Koblet til journal i Google Drive",
		EventJournalDetached:                "Koblet fra journal i Google Drive",
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
	AdminDefaultIncludeTag:      "Show at check-in",
	AdminDisplayName:            "Name",
	AdminEmailAddress:           "Email address",
	AdminInviteToBino:           "Invite to Bino",
	AdminManageEvents:           "Manage event types",
	AdminManageGoogleDrive:      "Configure Google Drive",
	AdminManageHomes:            "Manage rehab homes",
	AdminManageSpecies:          "Manage species",
	AdminManageStatuses:         "Manage statuses",
	AdminManageTags:             "Manage tags",
	AdminManageInvites:          "Invitations",
	AdminManageUsers:            "Manage users",
	AdminScrubUserData:          "Delete user data",
	AdminScrubUserDataConfirm:   "Write the user's email address to confirm that you want to scrub all user data for this user",
	AdminNukeUser:               "Destroy user",
	AdminNukeUserConfirm:        "Write the user's email address to confirm that you want to destroy the user record (this also removes all content created by the user)",
	AdminAbortedDueToWrongEmail: "Wrong email address. The action was cancelled.",
	AdminUserDeletionFailed:     "Failed to delete the user. Contact site administrator.",
	AdminUserWasDeleted:         "The user was deleted.",
	AdminExistingUsers:          "Bino users",
	AdminInviteUsers:            "Invite new users",
	AdminInviteExpires:          "Expires",
	AdminPendingInvitations:     "Pending invitations",
	AdminInvitationFailed:       "Failed to invite user. Contact site administrator.",
	AdminInvitationOKNoEmail:    "The user was added to the list of invited user. No email was sent; send them a link to the main page and ask them to log in.",
	AdminInviteCode:             "Code",
	AdminRoot:                   "Admin",

	AuthLogOut: "Log out",

	CheckinCheckInPatient:  "Check in",
	CheckinIHaveThePatient: "The patient is here",
	CheckinPatientName:     "Name of the patient",
	CheckinYouAreHomeless:  "You can't check in patients yet because you're not connected to a rehab home.",

	DashboardNoPatientsInHome:      "No patients",
	DashboardGoToJournal:           "Go to patient journal in Google Drive",
	DashboardGoToPatientPage:       "Go to patient page",
	DashboardCheckOut:              "Checkout",
	DashboardSearch:                "Search in homes and patients",
	DashboardSearchFilter:          "Filter homes",
	DashboardSearchShowMine:        "Show my home",
	DashboardSearchShowFull:        "Show full homes",
	DashboardSearchShowUnavailable: "Show unavailable homes",
	DashboardSelectHome:            "Select home",
	DashboardSelectCheckout:        "Select status",
	DashboardSelectTag:             "Select tag",
	DashboardSelectSpecies:         "Select species",
	DashboardNonPreferredSpecies:   "Other species",
	DashboardOtherHome:             "Other homes",

	ErrorPageHead:         "Error",
	ErrorPageInstructions: "An error occurred while loading the page. The error has been logged and will be investigated. Send a message to the site admin for help. The technical error message is as follows.",
	ErrorSettingLanguage:  "Failed to update language",

	FooterPrivacy:    "Privacy",
	FooterSourceCode: "Source code",

	FormerPatients: "Former patients",

	GDriveBaseDir:                          "Select journal folder",
	GDriveSelectFolder:                     "Select folder",
	GDriveSelectedFolder:                   "Selected",
	GDriveSelectFolderInstruction:          "Select the folder in which new patient journals will be created.",
	GDriveReloadFolders:                    "Reload folders from Google Drive",
	GDriveTemplate:                         "Choose template",
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
	GDriveCreateJournalForPatient:          "Create journal in Google Drive",
	GDriveSelectExistingJournalInstruction: "Or connect an existing journal in Google Drive:",
	GDriveNoJournalForPatient:              "No journal found",
	GDriveCreateJournalFailed:              "Patient was added, I couldn't create the journal in Google Drive",

	GenericAdd:      "Add",
	GenericAge:      "Age",
	GenericAvatar:   "Avatar",
	GenericCancel:   "Cancel",
	GenericConfirm:  "Confirm",
	GenericDelete:   "Delete",
	GenericDetails:  "Details",
	GenericEmail:    "Email",
	GenericFrom:     "From",
	GenericTo:       "To",
	GenericGoBack:   "Go back",
	GenericHome:     "Home",
	GenericJournal:  "Journal",
	GenericLatin:    "Latin",
	GenericMove:     "Move",
	GenericMoveTo:   "Move to",
	GenericName:     "Name",
	GenericNone:     "None",
	GenericNever:    "Never",
	GenericNote:     "Note",
	GenericNotFound: "Not found",
	GenericSpecies:  "Species",
	GenericStatus:   "Status",
	GenericTags:     "Tags",
	GenericUpdate:   "Update",
	GenericFailed:   "Something went wrong. Contact the site administrator.",
	GenericSuccess:  "Success.",
	GenericMessage:  "Message",
	GenericURL:      "URL",

	HomesAddToHome:                 "Add",
	HomesArchiveHome:               "Archive rehab home",
	HomesCreateHome:                "Create new rehab home",
	HomesCreateHomeNote:            "The name is usually that of a person, but there can be multiple people in a rehab home, and one person can be associated with several rehab homes.",
	HomesEmptyHome:                 "There are no users in this rehab home.",
	HomesHomeName:                  "Name",
	HomesRemoveFromCurrent:         "Remove from this rehab home",
	HomesUnassignedUsers:           "Users that are not associated with any rehab homes",
	HomesViewHomes:                 "Rehab homes",
	HomesPatients:                  "Patients",
	HomesUsers:                     "Users",
	HomeCapacity:                   "Capacity",
	HomePreferredSpecies:           "Favorites",
	HomeAvailability:               "Availability",
	HomePeriodInvalid:              "Invalid period.",
	HomeAvailableIndefinitely:      "Available.",
	HomeUnavailableIndefinitely:    "Unavailable until further notice.",
	HomeUnavailableFromInstruction: "The date when you become unavailable.",
	HomeUnavailableToInstruction:   "The last date when you are unavailable.",

	LanguageUpdateFailed: "Failed to update language",

	NavbarCalendar:  "Calendar",
	NavbarDashboard: "Dashboard",

	PatientRegisteredTime: "Registrert",
	PatientCheckedOutTime: "Checked out",
	PatientEventTime:      "Time",
	PatientEventEvent:     "Event",
	PatientEventNote:      "Note",
	PatientEventUser:      "User",
	PatientEventHome:      "Home",

	UserHomes:      "Associated rehab homes",
	UserIsHomeless: "No associated rehab homes",

	Status: map[Status]string{
		StatusUnknown:                        "Unknown",
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
		EventJournalCreated:                 "Created journal",
		EventJournalAttached:                "Linked journal",
		EventJournalDetached:                "Unlinked journal",
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

func (l *Language) FormatTimeRelWithAbsFallback(t time.Time) string {
	if t.IsZero() {
		return l.GenericNever
	}

	rel := l.FormatTimeRel(t)
	if rel != "" {
		return rel
	}
	return l.FormatTimeAbs(t)
}

func (l *Language) FormatTimeAbsWithRelParen(t time.Time) string {
	if t.IsZero() {
		return l.GenericNever
	}

	abs := l.FormatTimeAbs(t)
	rel := l.FormatTimeRel(t)
	if rel == "" {
		return abs
	}
	return fmt.Sprintf("%s (%s)", abs, rel)
}

func (l *Language) FormatTimeAbs(t time.Time) string {
	if t.IsZero() {
		return l.GenericNever
	}

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
	if t.IsZero() {
		return l.GenericNever
	}

	now := time.Now()
	diff := now.Sub(t)

	switch l.ID {
	case LanguageIDNO:
		if diff < -356*24*time.Hour {
			return ""
		}
		if diff < -2*24*time.Hour {
			return fmt.Sprintf("om %d dager", -int(diff.Hours()/24))
		}
		if diff < -24*time.Hour {
			return fmt.Sprintf("om 1 dag")
		}
		if diff < -2*time.Hour {
			return fmt.Sprintf("om %d timer", -int(diff.Hours()))
		}
		if diff < -time.Hour {
			return fmt.Sprintf("om 1 time")
		}
		if diff < -2*time.Minute {
			return fmt.Sprintf("om %d minutter", -int(diff.Minutes()))
		}
		if diff < -time.Minute {
			return fmt.Sprintf("om 1 minutt")
		}
		if diff < -2*time.Second {
			return fmt.Sprintf("om %d sekunder", -int(diff.Seconds()))
		}
		if diff < -time.Second {
			return fmt.Sprintf("om 1 sekund")
		}
		if diff < time.Second {
			return fmt.Sprintf("akkurat nå")
		}
		if diff < 2*time.Second {
			return fmt.Sprintf("for 1 sekund siden")
		}
		if diff < time.Minute {
			return fmt.Sprintf("for %d sekunder siden", int(diff.Seconds()))
		}
		if diff < 2*time.Minute {
			return fmt.Sprintf("for 1 minutt siden")
		}
		if diff < time.Hour {
			return fmt.Sprintf("for %d minutter siden", int(diff.Minutes()))
		}
		if diff < 2*time.Hour {
			return fmt.Sprintf("for 1 time siden")
		}
		if diff < 24*time.Hour {
			return fmt.Sprintf("for %d timer siden", int(diff.Hours()))
		}
		if diff < 2*24*time.Hour {
			return fmt.Sprintf("for 1 dag siden")
		}
		if diff < 356*24*time.Hour {
			return fmt.Sprintf("for %d dager siden", int(diff.Hours()/24))
		}
	case LanguageIDEN:
		if diff < -356*24*time.Hour {
			return ""
		}
		if diff < -2*24*time.Hour {
			return fmt.Sprintf("in %d days", -int(diff.Hours()/24))
		}
		if diff < -24*time.Hour {
			return fmt.Sprintf("in 1 day")
		}
		if diff < -2*time.Hour {
			return fmt.Sprintf("in %d hours", -int(diff.Hours()))
		}
		if diff < -time.Hour {
			return fmt.Sprintf("in 1 hour")
		}
		if diff < -2*time.Minute {
			return fmt.Sprintf("in %d minutes", -int(diff.Minutes()))
		}
		if diff < -time.Minute {
			return fmt.Sprintf("in 1 minute")
		}
		if diff < -2*time.Second {
			return fmt.Sprintf("in %d seconds", -int(diff.Seconds()))
		}
		if diff < -time.Second {
			return fmt.Sprintf("in 1 second")
		}
		if diff < time.Second {
			return fmt.Sprintf("just now")
		}
		if diff < 2*time.Second {
			return fmt.Sprintf("1 second ago")
		}
		if diff < time.Minute {
			return fmt.Sprintf("%d seconds ago", int(diff.Seconds()))
		}
		if diff < 2*time.Minute {
			return fmt.Sprintf("1 minute ago")
		}
		if diff < time.Hour {
			return fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
		}
		if diff < 2*time.Hour {
			return fmt.Sprintf("1 hour ago")
		}
		if diff < 24*time.Hour {
			return fmt.Sprintf("%d hours ago", int(diff.Hours()))
		}
		if diff < 2*24*time.Hour {
			return fmt.Sprintf("1 day ago")
		}
		if diff < 356*24*time.Hour {
			return fmt.Sprintf("%d days ago", int(diff.Hours()/24))
		}
	}
	return ""
}

func (l *Language) AdminDefaultInviteMessage(inviter string) string {
	switch l.ID {
	case LanguageIDNO:
		return fmt.Sprintf("%s har invitert deg til å opprette en bruker i Bino.", inviter)
	case LanguageIDEN:
		fallthrough
	default:
		return fmt.Sprintf("%s has invited you to create a user in Bino.", inviter)
	}
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
