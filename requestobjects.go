package main

import (
	"context"
	"fmt"
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
	BuildKey string
	User     UserData
}

func (cd *CommonData) StaticFile(name string) string {
	return fmt.Sprintf("/static/%s/%s", cd.BuildKey, name)
}

func (cd *CommonData) Lang() LanguageID {
	return cd.User.Language.ID
}

func (cd *CommonData) Lang32() int32 {
	return int32(cd.User.Language.ID)
}

type UserData struct {
	AppuserID       int32
	DisplayName     string
	PreferredHomeID int32
	Homes           []int32
	Email           string
	Language        *Language
	LoggingConsent  bool
	AvatarURL       string
	HasAvatarURL    bool
}

type LanguageView struct {
	ID       int32
	Emoji    string
	SelfName string
}
