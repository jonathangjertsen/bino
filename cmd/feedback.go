//go:generate go tool go-enum --no-iota --values --noprefix --prefix=FB
package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Feedback struct {
	Items    []FeedbackItem
	NSkipped int
}

func (fb Feedback) CSSClass() string {
	var maxFBT FeedbackType
	for _, item := range fb.Items {
		if item.Type > maxFBT {
			maxFBT = item.Type
		}
	}
	return maxFBT.CSSClass()
}

type FeedbackItem struct {
	Message string
	Type    FeedbackType
}

// ENUM(
//
//	Info = 0,
//	Success = 1,
//	Warning = 2,
//	Error = 3,
//
// )
type FeedbackType int32

func (fbt FeedbackType) CSSClass() string {
	return "feedback-" + strings.ToLower(fbt.String())
}

func (server *Server) setFeedbackCookie(w http.ResponseWriter, r *http.Request) {
	// Try to get user data
	cd, err := LoadCommonData(r.Context())
	if err != nil {
		return
	}

	// Try to marshal the feedback we have now
	str, err := json.Marshal(cd.Feedback)
	if err != nil {
		cd.Log("marshalling feedback: %v", err)
		return
	}

	// Set the cookie
	sess, _ := server.Cookies.Get(r, "feedback")
	sess.Values["json"] = string(str)
	sess.Options.MaxAge = 5
	if err := sess.Save(r, w); err != nil {
		cd.Log("saving feedback cookie: %v", err)
	}
}

func (server *Server) eatFeedbackCookie(w http.ResponseWriter, r *http.Request) {
	// Try to get user data
	cd, err := LoadCommonData(r.Context())
	if err != nil {
		return
	}

	// Get the feedback cookie and destroy it
	sess, err := server.Cookies.Get(r, "feedback")
	if err != nil {
		cd.Log("failed to decode cookie session: %v", err)
		return
	}

	// Try to get the cookie and then eat it
	str, ok := sess.Values["json"].(string)
	sess.Options.MaxAge = -1
	if err := sess.Save(r, w); err != nil {
		cd.Log("saving feedback cookie: %v", err)
	}
	if !ok {
		return
	}

	// Try to unmarshal it
	var feedback Feedback
	if err := json.Unmarshal([]byte(str), &feedback); err != nil {
		cd.Log("failed to unmarshal feedback cookie: %v", err)
		return
	}

	// Prepend the feedback from the cookie, since that happened before this request
	cd.Feedback.Items = append(feedback.Items, cd.Feedback.Items...)
	if len(cd.Feedback.Items) > 10 {
		cd.Feedback.Items = cd.Feedback.Items[:10]
		cd.Feedback.NSkipped += len(cd.Feedback.Items) - 10
	}
}
