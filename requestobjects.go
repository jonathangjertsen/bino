package main

import (
	"context"
	"fmt"

	"github.com/jonathangjertsen/bino/ln"
)

type ctxKey int32

const (
	ctxKeyCommonData ctxKey = iota
)

func WithCommonData(ctx context.Context, cd *CommonData) context.Context {
	return context.WithValue(ctx, ctxKeyCommonData, cd)
}

func LoadCommonData(ctx context.Context) (*CommonData, error) {
	cd, ok := ctx.Value(ctxKeyCommonData).(*CommonData)
	if !ok {
		return nil, fmt.Errorf("no CommonData in ctx")
	}
	return cd, nil
}

func MustLoadCommonData(ctx context.Context) *CommonData {
	cd, err := LoadCommonData(ctx)
	if err != nil {
		panic(err)
	}
	return cd
}

type CommonData struct {
	BuildKey  string
	User      UserData
	Languages []Language
}

func (cd *CommonData) Ln(key ln.L) string {
	return cd.User.Ln(key)
}

type UserData struct {
	AppuserID       int32
	DisplayName     string
	PreferredHomeID int32
	Homes           []int32
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
