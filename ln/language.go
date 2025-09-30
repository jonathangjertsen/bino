//go:generate go-enum --noprefix
package ln

// ENUM(
//
// LogOut,
// PatientName,
// Species,
// Tags,
// CheckInPatient,
// ManageSpecies,
// Latin,
// Update,
// Add,
// ManageHomes,
// Dashboard,
// ManageTags,
// Calendar,
// ManageStatuses,
// ManageEvents,
// AdminRoot,
// ViewHomes,
// ErrorPageHead,
// ErrorPageInstructions,
//
// )
type L int

var NO = []string{
	LogOut:                "Logg ut",
	PatientName:           "Pasientens navn",
	Species:               "Art",
	Tags:                  "Tagger",
	CheckInPatient:        "Sjekk in pasient",
	ManageSpecies:         "Konfigurer arter",
	Latin:                 "Latin",
	Update:                "Oppdater",
	Add:                   "Legg til",
	ManageHomes:           "Konfigurer rehabhjem",
	Dashboard:             "Hovedside",
	ManageTags:            "Konfigurer tagger",
	Calendar:              "Kalender",
	ManageStatuses:        "Konfigurer statuser",
	ManageEvents:          "Konfigurer hendelsestyper",
	AdminRoot:             "Konfigurering",
	ViewHomes:             "Rehabhjem",
	ErrorPageHead:         "Feilmelding",
	ErrorPageInstructions: "Det skjedde noe feil under lasting av siden. Feilen har blitt logget og vil bli undersøkt. Send melding til administrator for hjelp. Den tekniske feilmeldingen følger under.",
}

var EN = []string{
	LogOut:                "Log out",
	PatientName:           "Name of the patient",
	Species:               "Species",
	Tags:                  "Tags",
	CheckInPatient:        "Check in",
	ManageSpecies:         "Manage species",
	Latin:                 "Latin",
	Update:                "Update",
	Add:                   "Add",
	ManageHomes:           "Manage rehab homes",
	Dashboard:             "Dashboard",
	ManageTags:            "Manage tags",
	Calendar:              "Calendar",
	ManageStatuses:        "Manage statuses",
	ManageEvents:          "Manage event types",
	AdminRoot:             "Admin",
	ViewHomes:             "Rehab homes",
	ErrorPageHead:         "Error",
	ErrorPageInstructions: "An error occurred while loading the page. The error has been logged and will be investigated. Send a message to the site admin for help. The technical error message is as follows.",
}

var LANG = [][]string{
	nil,
	NO,
	EN,
}

func Ln(id int32, key L) string {
	if int(id) > len(LANG) {
		return key.String()
	}
	lang := LANG[id]
	if int(key) > len(lang) {
		return key.String()
	}
	return lang[key]
}
