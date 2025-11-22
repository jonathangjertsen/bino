//go:generate go tool go-enum --no-iota --noprefix --prefix=Cap
package main

import "slices"

// ENUM(
// ViewAllActivePatients,
// ViewAllFormerPatients,
// ViewAllHomes,
// ViewAllUsers,
// ViewCalendar,
// Search,
// SetOwnPreferences,
// CheckInPatient,
// ManageOwnPatients,
// ManageAllPatients,
// ManageOwnHomes,
// ManageAllHomes,
// CreatePatientJournal,
// ManageSpecies,
// ManageUsers,
// DeleteUsers,
// ViewAdminTools,
// ViewGDriveSettings,
// InviteToGDrive,
// InviteToBino,
// UseImportTool,
// Debug,
// UploadFile,
// )
type Capability int32

var RequiredAccessLevel = map[Capability]AccessLevel{
	CapViewAllActivePatients: AccessLevelNone,
	CapViewAllFormerPatients: AccessLevelNone,
	CapViewAllHomes:          AccessLevelNone,
	CapViewAllUsers:          AccessLevelNone,
	CapViewCalendar:          AccessLevelNone,
	CapSearch:                AccessLevelNone,
	CapSetOwnPreferences:     AccessLevelNone,

	CapCheckInPatient:       AccessLevelRehabber,
	CapManageOwnPatients:    AccessLevelRehabber,
	CapManageOwnHomes:       AccessLevelRehabber,
	CapCreatePatientJournal: AccessLevelRehabber,
	CapViewGDriveSettings:   AccessLevelRehabber,
	CapManageAllPatients:    AccessLevelRehabber,
	CapUploadFile:           AccessLevelRehabber,

	CapViewAdminTools: AccessLevelCoordinator,
	CapManageAllHomes: AccessLevelCoordinator,
	CapManageSpecies:  AccessLevelCoordinator,
	CapUseImportTool:  AccessLevelCoordinator,

	CapManageUsers:    AccessLevelAdmin,
	CapDeleteUsers:    AccessLevelAdmin,
	CapInviteToGDrive: AccessLevelAdmin,
	CapInviteToBino:   AccessLevelAdmin,
	CapDebug:          AccessLevelAdmin,
}

var AccessLevelToCapabilities = func() (out struct {
	None        []Capability
	Rehabber    []Capability
	Coordinator []Capability
	Admin       []Capability
}) {
	for cap, al := range RequiredAccessLevel {
		switch al {
		case AccessLevelNone:
			out.None = append(out.None, cap)
		case AccessLevelRehabber:
			out.Rehabber = append(out.Rehabber, cap)
		case AccessLevelCoordinator:
			out.Coordinator = append(out.Coordinator, cap)
		case AccessLevelAdmin:
			out.Admin = append(out.Admin, cap)
		}
	}
	slices.Sort(out.None)
	slices.Sort(out.Rehabber)
	slices.Sort(out.Coordinator)
	slices.Sort(out.Admin)
	return out
}()
