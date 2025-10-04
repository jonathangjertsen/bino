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
// DefaultIncludeTag,
// UnassignedUsers,
// DisplayName,
// EmailAddress,
// AddToHome,
// CreateHome,
// HomeName,
// CreateHomeNote,
// EmptyHome,
// AddUserToHome,
// CannotAddPatientBecauseUserIsHomeless,
// RemoveFromCurrent,
// Move,
// MoveTo,
// None,
// ArchiveHome,
// Privacy,
// SourceCode,
// IHaveThePatient,
//
// )
type L int

var NO = []string{
	LogOut:                                "Logg ut",
	PatientName:                           "Pasientens navn",
	Species:                               "Art",
	Tags:                                  "Tagger",
	CheckInPatient:                        "Sjekk inn pasient",
	ManageSpecies:                         "Konfigurer arter",
	Latin:                                 "Latin",
	Update:                                "Oppdater",
	Add:                                   "Legg til",
	ManageHomes:                           "Konfigurer rehabhjem",
	Dashboard:                             "Hovedside",
	ManageTags:                            "Konfigurer tagger",
	Calendar:                              "Kalender",
	ManageStatuses:                        "Konfigurer statuser",
	ManageEvents:                          "Konfigurer hendelsestyper",
	AdminRoot:                             "Konfigurering",
	ViewHomes:                             "Rehabhjem",
	ErrorPageHead:                         "Feilmelding",
	ErrorPageInstructions:                 "Det skjedde noe feil under lasting av siden. Feilen har blitt logget og vil bli undersøkt. Send melding til administrator for hjelp. Den tekniske feilmeldingen følger under.",
	DefaultIncludeTag:                     "Vis ved innsjekk",
	UnassignedUsers:                       "Brukere som ikke er koblet til noe rehabhjem",
	DisplayName:                           "Navn",
	EmailAddress:                          "Epostaddresse",
	AddToHome:                             "Legg til",
	CreateHome:                            "Opprett nytt rehabhjem",
	HomeName:                              "Rehabhjem",
	CreateHomeNote:                        "Navnet er som regel navnet på en person, men det kan være flere personer i ett rehabhjem, og en person kan være del av flere rehabhjem.",
	EmptyHome:                             "Det er ingen brukere i dette rehabhjemmet.",
	CannotAddPatientBecauseUserIsHomeless: "Du kan ikke sjekke inn pasienter ennå fordi du ikke er koblet til et rehabhjem.",
	RemoveFromCurrent:                     "Fjern fra dette rehabhjemmet",
	Move:                                  "Flytt",
	MoveTo:                                "Flytt til",
	None:                                  "Ingen",
	ArchiveHome:                           "Arkiver rehabhjem",
	Privacy:                               "Personvern",
	SourceCode:                            "Kildekode",
	IHaveThePatient:                       "Patienten er her",
}

var EN = []string{
	LogOut:                                "Log out",
	PatientName:                           "Name of the patient",
	Species:                               "Species",
	Tags:                                  "Tags",
	CheckInPatient:                        "Check in",
	ManageSpecies:                         "Manage species",
	Latin:                                 "Latin",
	Update:                                "Update",
	Add:                                   "Add",
	ManageHomes:                           "Manage rehab homes",
	Dashboard:                             "Dashboard",
	ManageTags:                            "Manage tags",
	Calendar:                              "Calendar",
	ManageStatuses:                        "Manage statuses",
	ManageEvents:                          "Manage event types",
	AdminRoot:                             "Admin",
	ViewHomes:                             "Rehab homes",
	ErrorPageHead:                         "Error",
	ErrorPageInstructions:                 "An error occurred while loading the page. The error has been logged and will be investigated. Send a message to the site admin for help. The technical error message is as follows.",
	DefaultIncludeTag:                     "Show at check-in",
	UnassignedUsers:                       "Users that are not associated with any rehab homes",
	DisplayName:                           "Name",
	EmailAddress:                          "Email address",
	AddToHome:                             "Add",
	CreateHome:                            "Create new rehab home",
	HomeName:                              "Name",
	CreateHomeNote:                        "The name is usually that of a person, but there can be multiple people in a rehab home, and one person can be associated with several rehab homes.",
	EmptyHome:                             "There are no users in this rehab home.",
	CannotAddPatientBecauseUserIsHomeless: "You can't check in patients yet because you're not connected to a rehab home.",
	RemoveFromCurrent:                     "Remove from this rehab home",
	Move:                                  "Move",
	MoveTo:                                "Move to",
	None:                                  "None",
	ArchiveHome:                           "Archive rehab home",
	Privacy:                               "Privacy",
	SourceCode:                            "Source code",
	IHaveThePatient:                       "The patient is here",
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
