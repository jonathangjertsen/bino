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
	// Cached result of queries that might be called more than once
	QueryCache struct {
		AllUsers         map[string]UserView
		CanCreateJournal bool
	}
	Feedback Feedback
}

func (cd *CommonData) Error(msg string, err error) {
	cd.Log("Showed user ERROR: err=%v, message to user=%s", err, msg)
	cd.SetFeedback(FBError, msg)
}

func (cd *CommonData) Warning(msg string, err error) {
	cd.Log("Showed user WARNING: err=%v, message to user=%s", err, msg)
	cd.SetFeedback(FBWarning, msg)
}

func (cd *CommonData) Success(msg string) {
	cd.SetFeedback(FBSuccess, msg)
}

func (cd *CommonData) Info(msg string) {
	cd.SetFeedback(FBInfo, msg)
}

func (cd *CommonData) SetFeedback(fbt FeedbackType, msg string) {
	if n := len(cd.Feedback.Items); n < 10 {
		// Filter dupes
		for i := range n {
			if cd.Feedback.Items[i].Message == msg {
				cd.Feedback.NSkipped++
				return
			}
		}

		cd.Feedback.Items = append(cd.Feedback.Items, FeedbackItem{
			Message: msg,
			Type:    fbt,
		})
	} else {
		cd.Feedback.NSkipped++
	}
}

func (server *Server) getUserViews(ctx context.Context) map[string]UserView {
	commonData := MustLoadCommonData(ctx)
	if commonData.QueryCache.AllUsers == nil {
		users, err := server.Queries.GetAppusers(ctx)
		if err == nil {
			commonData.QueryCache.AllUsers = SliceToMap(users, func(user GetAppusersRow) (string, UserView) {
				return user.Email, user.ToUserView()
			})
		} else {
			LogCtx(ctx, "GetAppusers failed: %w", err)
		}
	}
	return commonData.QueryCache.AllUsers
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
	HasGDriveAccess bool
}

type LanguageView struct {
	ID       int32
	Emoji    string
	SelfName string
}
