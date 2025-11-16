//go:generate go tool go-enum --no-iota --values
package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/docs/v1"
)

// ENUM(
//
//	YYYY,
//	MM,
//	DD,
//	Name,
//	Species,
//	BinoURL,
//
// )
type Template string

type GDriveJournal struct {
	Item    GDriveItem
	Content string
}

type GDriveTemplateVars struct {
	Time    time.Time
	Name    string
	Species string
	BinoURL string
}

func (vars *GDriveTemplateVars) ApplyToString(s string) string {
	for _, t := range TemplateValues() {
		s = strings.ReplaceAll(s, t.String(), vars.Replacement(t))
	}
	return s
}

func (vars *GDriveTemplateVars) Replacement(template Template) string {
	switch template {
	case TemplateYYYY:
		return fmt.Sprintf("%d", vars.Time.Year())
	case TemplateMM:
		return fmt.Sprintf("%02d", vars.Time.Month())
	case TemplateDD:
		return fmt.Sprintf("%02d", vars.Time.Day())
	case TemplateName:
		return vars.Name
	case TemplateSpecies:
		return vars.Species
	case TemplateBinoURL:
		return vars.BinoURL
	default:
		return template.String()
	}
}

func (vars *GDriveTemplateVars) ReplaceRequests() *docs.BatchUpdateDocumentRequest {
	return &docs.BatchUpdateDocumentRequest{
		Requests: SliceToSlice(TemplateValues(), func(t Template) *docs.Request {
			return &docs.Request{
				ReplaceAllText: &docs.ReplaceAllTextRequest{
					ContainsText: &docs.SubstringMatchCriteria{
						MatchCase: true,
						Text:      t.String(),
					},
					ReplaceText: vars.Replacement(t),
				},
			}
		}),
	}
}

func (gdj *GDriveJournal) Validate() error {
	errs := []error{}
	for _, k := range TemplateValues() {
		if !strings.Contains(gdj.Content, k.String()) {
			errs = append(errs, fmt.Errorf("template is missing variable '%s'", k))
		}
	}
	return errors.Join(errs...)
}
