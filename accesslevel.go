//go:generate go tool go-enum --no-iota --values
package main

// ENUM(
//
//	None        = 0,
//	Rehabber    = 1,
//	Coordinator = 2,
//	Admin       = 3,
//
// )
type AccessLevel int32
