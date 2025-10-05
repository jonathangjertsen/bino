package ln

type Language struct {
	ID int32

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

	ErrorPageHead         string
	ErrorPageInstructions string

	FooterPrivacy    string
	FooterSourceCode string

	GenericAdd     string
	GenericDelete  string
	GenericLatin   string
	GenericMove    string
	GenericMoveTo  string
	GenericNone    string
	GenericSpecies string
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
}

var NO = &Language{
	ID: 1,

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

	ErrorPageHead:         "Feilmelding",
	ErrorPageInstructions: "Det skjedde noe feil under lasting av siden. Feilen har blitt logget og vil bli undersøkt. Send melding til administrator for hjelp. Den tekniske feilmeldingen følger under.",

	FooterPrivacy:    "Personvern",
	FooterSourceCode: "Kildekode",

	GenericAdd:     "Legg til",
	GenericDelete:  "Slett",
	GenericLatin:   "Latin",
	GenericMove:    "Flytt",
	GenericMoveTo:  "Flytt til",
	GenericNone:    "Ingen",
	GenericSpecies: "Art",
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
}

var EN = &Language{
	ID: 2,

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

	ErrorPageHead:         "Error",
	ErrorPageInstructions: "An error occurred while loading the page. The error has been logged and will be investigated. Send a message to the site admin for help. The technical error message is as follows.",

	FooterPrivacy:    "Privacy",
	FooterSourceCode: "Source code",

	GenericAdd:     "Add",
	GenericDelete:  "Delete",
	GenericLatin:   "Latin",
	GenericMove:    "Move",
	GenericMoveTo:  "Move to",
	GenericNone:    "None",
	GenericSpecies: "Species",
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
}

var Languages = map[int32]*Language{
	NO.ID: NO,
	EN.ID: EN,
}

func GetLanguage(id int32) *Language {
	lang, ok := Languages[id]
	if !ok {
		return EN
	}
	return lang
}
