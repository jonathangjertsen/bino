//go:generate go-enum
package ln

// ENUM(
//
// LogOut,
// PatientName,
// Species,
// Details,
// CheckInPatient,
//
// )
type Key int

var NO = []string{
	KeyLogOut:         "Logg ut",
	KeyPatientName:    "Pasientens navn",
	KeySpecies:        "Art",
	KeyDetails:        "Detaljer",
	KeyCheckInPatient: "Sjekk in pasient",
}

var EN = []string{
	KeyLogOut:         "Log out",
	KeyPatientName:    "Name of the patient",
	KeySpecies:        "Species",
	KeyDetails:        "Details",
	KeyCheckInPatient: "Check in",
}

var LANG = [][]string{
	nil,
	NO,
	EN,
}

func Ln(id int32, key Key) string {
	if int(id) > len(LANG) {
		return key.String()
	}
	lang := LANG[id]
	if int(key) > len(lang) {
		return key.String()
	}
	return lang[key]
}
