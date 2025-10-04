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

	ErrorPageHead         string
	ErrorPageInstructions string

	FooterPrivacy    string
	FooterSourceCode string

	GenericAdd     string
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

	AuthLogOut:             "Logg ut",
	CheckinPatientName:     "Pasientens navn",
	GenericSpecies:         "Art",
	GenericTags:            "Tagger",
	CheckinCheckInPatient:  "Sjekk inn pasient",
	AdminManageSpecies:     "Konfigurer arter",
	GenericLatin:           "Latin",
	GenericUpdate:          "Oppdater",
	GenericAdd:             "Legg til",
	AdminManageHomes:       "Konfigurer rehabhjem",
	NavbarDashboard:        "Hovedside",
	AdminManageTags:        "Konfigurer tagger",
	NavbarCalendar:         "Kalender",
	AdminManageStatuses:    "Konfigurer statuser",
	AdminManageEvents:      "Konfigurer hendelsestyper",
	AdminRoot:              "Konfigurering",
	HomesViewHomes:         "Rehabhjem",
	ErrorPageHead:          "Feilmelding",
	ErrorPageInstructions:  "Det skjedde noe feil under lasting av siden. Feilen har blitt logget og vil bli undersøkt. Send melding til administrator for hjelp. Den tekniske feilmeldingen følger under.",
	AdminDefaultIncludeTag: "Vis ved innsjekk",
	HomesUnassignedUsers:   "Brukere som ikke er koblet til noe rehabhjem",
	AdminDisplayName:       "Navn",
	AdminEmailAddress:      "Epostaddresse",
	HomesAddToHome:         "Legg til",
	HomesCreateHome:        "Opprett nytt rehabhjem",
	HomesHomeName:          "Rehabhjem",
	HomesCreateHomeNote:    "Navnet er som regel navnet på en person, men det kan være flere personer i ett rehabhjem, og en person kan være del av flere rehabhjem.",
	HomesEmptyHome:         "Det er ingen brukere i dette rehabhjemmet.",
	CheckinYouAreHomeless:  "Du kan ikke sjekke inn pasienter ennå fordi du ikke er koblet til et rehabhjem.",
	HomesRemoveFromCurrent: "Fjern fra dette rehabhjemmet",
	GenericMove:            "Flytt",
	GenericMoveTo:          "Flytt til",
	GenericNone:            "Ingen",
	HomesArchiveHome:       "Arkiver rehabhjem",
	FooterPrivacy:          "Personvern",
	FooterSourceCode:       "Kildekode",
	CheckinIHaveThePatient: "Pasienten er her",
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
	HomesUnassignedUsers:   "Users that are not associated with any rehab homes",
	AuthLogOut:             "Log out",
	CheckinCheckInPatient:  "Check in",
	CheckinIHaveThePatient: "The patient is here",
	CheckinPatientName:     "Name of the patient",
	ErrorPageHead:          "Error",
	ErrorPageInstructions:  "An error occurred while loading the page. The error has been logged and will be investigated. Send a message to the site admin for help. The technical error message is as follows.",
	FooterPrivacy:          "Privacy",
	FooterSourceCode:       "Source code",
	GenericAdd:             "Add",
	GenericLatin:           "Latin",
	GenericMove:            "Move",
	GenericMoveTo:          "Move to",
	GenericNone:            "None",
	GenericSpecies:         "Species",
	GenericTags:            "Tags",
	GenericUpdate:          "Update",
	HomesArchiveHome:       "Archive rehab home",
	HomesAddToHome:         "Add",
	CheckinYouAreHomeless:  "You can't check in patients yet because you're not connected to a rehab home.",
	HomesCreateHome:        "Create new rehab home",
	HomesCreateHomeNote:    "The name is usually that of a person, but there can be multiple people in a rehab home, and one person can be associated with several rehab homes.",
	HomesEmptyHome:         "There are no users in this rehab home.",
	HomesHomeName:          "Name",
	HomesRemoveFromCurrent: "Remove from this rehab home",
	HomesViewHomes:         "Rehab homes",
	NavbarCalendar:         "Calendar",
	NavbarDashboard:        "Dashboard",
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
