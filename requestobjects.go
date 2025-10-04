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
	Languages []LanguageView
}

type UserData struct {
	AppuserID       int32
	DisplayName     string
	PreferredHomeID int32
	Homes           []int32
	Email           string
	Language        *ln.Language
	LoggingConsent  bool
}

type LanguageView struct {
	ID       int32
	Emoji    string
	SelfName string
}
