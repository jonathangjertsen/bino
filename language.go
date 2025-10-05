//go:generate go tool go-enum --no-iota --values
package main

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

	ErrorPageHead         string
	ErrorPageInstructions string

	FooterPrivacy    string
	FooterSourceCode string

	GenericAdd     string
	GenericDelete  string
	GenericDetails string
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

	NavbarCalendar  string
	NavbarDashboard string

	Status map[Status]string
}

var NO = &Language{
	ID:       LanguageIDNO,
	Emoji:    "🇳🇴",
	SelfName: "Norsk",

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

	ErrorPageHead:         "Feilmelding",
	ErrorPageInstructions: "Det skjedde noe feil under lasting av siden. Feilen har blitt logget og vil bli undersøkt. Send melding til administrator for hjelp. Den tekniske feilmeldingen følger under.",

	FooterPrivacy:    "Personvern",
	FooterSourceCode: "Kildekode",

	GenericAdd:     "Legg til",
	GenericDelete:  "Slett",
	GenericDetails: "Detaljer",
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

	NavbarCalendar:  "Kalender",
	NavbarDashboard: "Hovedside",

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
}

var EN = &Language{
	ID:       LanguageIDEN,
	Emoji:    "🇬🇧",
	SelfName: "English",

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

	ErrorPageHead:         "Error",
	ErrorPageInstructions: "An error occurred while loading the page. The error has been logged and will be investigated. Send a message to the site admin for help. The technical error message is as follows.",

	FooterPrivacy:    "Privacy",
	FooterSourceCode: "Source code",

	GenericAdd:     "Add",
	GenericDelete:  "Delete",
	GenericDetails: "Details",
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
