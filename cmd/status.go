//go:generate go tool go-enum --no-iota --values
package main

// ENUM(
//
//	Unknown                        = 0,
//	Admitted                       = 2,
//	Released                       = 3,
//	Dead                           = 4,
//	Euthanized                     = 5,
//	TransferredOutsideOrganization = 6,
//	Adopted                        = 7,
//	Deleted                        = 8,
//
// )
type Status int32

var IsCheckoutStatus = map[Status]bool{
	StatusUnknown:                        false,
	StatusAdmitted:                       false,
	StatusAdopted:                        true,
	StatusReleased:                       true,
	StatusTransferredOutsideOrganization: true,
	StatusDead:                           true,
	StatusEuthanized:                     true,
	StatusDeleted:                        true,
}
