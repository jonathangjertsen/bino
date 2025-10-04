//go:generate go tool go-enum --no-iota
package main

// ENUM(
//
//	Unknown                        = 0,
//	PendingAdmission               = 1,
//	Admitted                       = 2,
//	Adopted                        = 3,
//	Released                       = 4,
//	TransferredOutsideOrganization = 5,
//	Dead                           = 6,
//
// )
type Status int32
