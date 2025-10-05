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
//	TagAdded                       = 9,  // Associated ID is tag ID
//	TagRemoved                     = 10, // Associated ID is tag ID
//	StatusChanged                  = 11, // Associated ID is status
//	Deleted                        = 12,
//
// )
type Event int32
