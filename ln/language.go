//go:generate go-enum --noprefix
package ln

// ENUM(
//
// LogOut,
// PatientName,
// Species,
// Details,
// CheckInPatient,
// ManageSpecies,
// Latin,
// Update,
// AddSpecies,
//
// )
type L int

var NO = []string{
	LogOut:         "Logg ut",
	PatientName:    "Pasientens navn",
	Species:        "Art",
	Details:        "Detaljer",
	CheckInPatient: "Sjekk in pasient",
	ManageSpecies:  "Administrer arter",
	Latin:          "Latin",
	Update:         "Oppdater",
	AddSpecies:     "Legg til art",
}

var EN = []string{
	LogOut:         "Log out",
	PatientName:    "Name of the patient",
	Species:        "Species",
	Details:        "Details",
	CheckInPatient: "Check in",
	ManageSpecies:  "Manage species",
	Latin:          "Latin",
	Update:         "Update",
	AddSpecies:     "Add species",
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
