//go:generate go tool go-enum --no-iota
package main

// ENUM(
//
//	Unknown                        = 0,
//	Registered                     = 1,
//	Admitted                       = 2,
//	Adopted                        = 3,
//	Released                       = 4,
//	TransferredToOtherHome         = 5,
//	TransferredOutsideOrganization = 6,
//	Died                           = 7,
//	Euthanized                     = 8,
//
// )
type Event int32
