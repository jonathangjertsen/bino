//go:generate go tool go-enum --no-iota --values --noprefix --prefix=FB
package main

import (
	"encoding/json"
	"fmt"
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

func (server *Server) setCookie(w http.ResponseWriter, r *http.Request, key string, value any) error {
	// Try to marshal the feedback we have now
	str, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshalling feedback: %w", err)
	}

	// Set cookie
	sess, _ := server.Cookies.Get(r, key)
	sess.Values["json"] = string(str)
	sess.Options.MaxAge = 3600
	if err := sess.Save(r, w); err != nil {
		return fmt.Errorf("saving cookie for '%s': %w", key, err)
	}

	return nil
}

func (server *Server) getCookie(w http.ResponseWriter, r *http.Request, key string, value any) (bool, error) {
	// Get the feedback cookie and destroy it
	sess, err := server.Cookies.Get(r, key)
	if err != nil {
		return false, fmt.Errorf("failed to decode cookie session: %v", err)
	}

	// Try to get the cookie
	str, ok := sess.Values["json"].(string)
	if !ok {
		return false, nil
	}

	// Try to unmarshal it
	if err := json.Unmarshal([]byte(str), &value); err != nil {
		return false, fmt.Errorf("failed to unmarshal cookie '%s': %w", key, err)
	}

	return true, nil
}

func (server *Server) eatCookie(w http.ResponseWriter, r *http.Request, key string, value any) (bool, error) {
	// Get the feedback cookie and destroy it
	sess, err := server.Cookies.Get(r, key)
	if err != nil {
		return false, fmt.Errorf("failed to decode cookie session: %v", err)
	}

	// Try to get the cookie, eat it regardless
	str, ok := sess.Values["json"].(string)
	sess.Options.MaxAge = -1
	if err := sess.Save(r, w); err != nil {
		return false, fmt.Errorf("deleting cookie '%s': %v", key, err)
	}
	if !ok {
		return false, nil
	}

	// Try to unmarshal it
	if err := json.Unmarshal([]byte(str), &value); err != nil {
		return false, fmt.Errorf("failed to unmarshal cookie '%s': %w", key, err)
	}

	return true, nil
}

func (server *Server) deleteCookie(w http.ResponseWriter, r *http.Request, key string) {
	sess, err := server.Cookies.Get(r, key)
	if err != nil {
		return
	}
	sess.Options.MaxAge = -1
	_ = sess.Save(r, w)
}

func (server *Server) setFeedbackCookie(w http.ResponseWriter, r *http.Request) {
	cd, err := LoadCommonData(r.Context())
	if err != nil {
		cd.Log("no common data: %w", err)
		return
	}

	if err := server.setCookie(w, r, "feedback", cd.Feedback); err != nil {
		cd.Log("%v", err)
		return
	}
}

func (server *Server) eatFeedbackCookie(w http.ResponseWriter, r *http.Request) {
	cd, err := LoadCommonData(r.Context())
	if err != nil {
		cd.Log("no common data: %w", err)
		return
	}

	var feedback Feedback
	if _, err := server.eatCookie(w, r, "feedback", &feedback); err != nil {
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
