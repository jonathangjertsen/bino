package main

import "github.com/jonathangjertsen/bino/ln"

type UserData struct {
	AppuserID       int32
	DisplayName     string
	PreferredHomeID int32
	Email           string
	LanguageID      int32
	LoggingConsent  bool
}

func (ud *UserData) Ln(key ln.L) string {
	return ln.Ln(ud.LanguageID, key)
}

type Language struct {
	ID       int32
	Emoji    string
	SelfName string
}

type CommonData struct {
	User      UserData
	Languages []Language
}

func (cd *CommonData) Ln(key ln.L) string {
	return cd.User.Ln(key)
}
