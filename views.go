//go:generate go tool go-enum --no-iota --values
package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func FormID(prefix, field string, id int32) string {
	return fmt.Sprintf("%s%s-%d", prefix, field, id)
}

// ---- Home

type HomeView struct {
	Home Home

	// Optional
	Patients           []PatientView
	Users              []UserView
	PreferredSpecies   []SpeciesView
	UnavailablePeriods []PeriodView
}

// ENUM(AvailableIndefinitely, AvailableUntil, UnavailableUntil, UnavailableIndefinitely)
type Availability int

func HomeURL(id int32) string {
	return fmt.Sprintf("/home/%d", id)
}

func (hv HomeView) URL() string {
	return HomeURL(hv.Home.ID)
}

func (hv HomeView) URLSuffix(suffix string) string {
	return fmt.Sprintf("/home/%d/%s", hv.Home.ID, suffix)
}

func (h HomeView) OccupancyClass() string {
	if len(h.Patients) < int(h.Home.Capacity) {
		return "text-success"
	}
	if len(h.Patients) == int(h.Home.Capacity) {
		return "text-warning"
	}
	return "text-danger"
}

func (h Home) ToHomeView() HomeView {
	return HomeView{
		Home: h,
	}
}

func (hv HomeView) AvailabilityDate() (Availability, DateView) {
	tomorrow := DateViewFromTime(time.Now().AddDate(0, 0, 1))
	for _, pv := range hv.UnavailablePeriods {
		if pv.From.Before(tomorrow) && tomorrow.Before(pv.To) {
			if pv.To.Year > tomorrow.Year+2 {
				return AvailabilityUnavailableIndefinitely, pv.To
			}
			return AvailabilityUnavailableUntil, pv.To
		}
		if tomorrow.Before(pv.From) {
			return AvailabilityAvailableUntil, pv.From
		}
	}
	return AvailabilityAvailableIndefinitely, tomorrow
}

func (hv HomeView) AvailabilityString(language *Language) (Availability, string) {
	availability, dv := hv.AvailabilityDate()
	switch availability {
	case AvailabilityAvailableIndefinitely:
		return availability, language.HomeAvailableIndefinitely
	case AvailabilityAvailableUntil:
		return availability, language.HomeAvailableUntil(dv)
	case AvailabilityUnavailableIndefinitely:
		return availability, language.HomeUnavailableIndefinitely
	case AvailabilityUnavailableUntil:
		return availability, language.HomeUnavailableUntil(dv)
	}
	return availability, language.HomeAvailableIndefinitely
}

// ---- Patient

type PatientView struct {
	ID           int32
	Status       int32
	Name         string
	Species      string
	Tags         []TagView
	JournalURL   string
	TimeCheckin  time.Time
	TimeCheckout time.Time
}

func PatientURL(id int32) string {
	return fmt.Sprintf("/patient/%d", id)
}

func (pv PatientView) CollapseID(prefix string) string {
	return fmt.Sprintf("%spatient-collapsible-%d", prefix, pv.ID)
}

func (pv PatientView) CheckoutNoteID(prefix string) string {
	return fmt.Sprintf("%spatient-checkout-note-%d", prefix, pv.ID)
}

func (pv PatientView) CheckoutStatusID(prefix string) string {
	return fmt.Sprintf("%spatient-checkout-status-%d", prefix, pv.ID)
}

func (pv PatientView) AttachJournalID(prefix string) string {
	return fmt.Sprintf("%spatient-attach-journal-%d", prefix, pv.ID)
}

func (pv PatientView) CardID(prefix string) string {
	return fmt.Sprintf("%spatient-card_%d", prefix, pv.ID)
}

func (pv PatientView) URL() string {
	return PatientURL(pv.ID)
}

func (pv PatientView) URLSuffix(suffix string) string {
	return fmt.Sprintf("/patient/%d/%s", pv.ID, suffix)
}

func (in GetFormerPatientsRow) ToPatientView() PatientView {
	return PatientView{
		ID:           in.ID,
		Status:       in.Status,
		Name:         in.Name,
		Species:      in.Species,
		TimeCheckin:  in.TimeCheckin.Time,
		TimeCheckout: in.TimeCheckout.Time,
	}
}

// ---- Tag

type TagView struct {
	ID        int32
	PatientID int32
	Name      string
}

func (tv TagView) URL() string {
	return fmt.Sprintf("/patient/%d/tag/%d", tv.PatientID, tv.ID)
}

func (in GetTagsForCurrentPatientsForHomeRow) ToTagView() TagView {
	return TagView{
		ID:        in.TagID,
		PatientID: in.PatientID,
		Name:      in.Name,
	}
}

func (in GetTagsForActivePatientsRow) ToTagView() TagView {
	return TagView{
		ID:        in.TagID,
		PatientID: in.PatientID,
		Name:      in.Name,
	}
}

// ---- User

type UserView struct {
	ID           int32
	Name         string
	Email        string
	AvatarURL    string
	HasAvatarURL bool
	AccessLevel  AccessLevel

	// Optional
	Homes []HomeView
}

func (u *UserView) Valid() bool {
	return u.ID > 0
}

func (u *UserView) URL() string {
	return fmt.Sprintf("/user/%d", u.ID)
}

func (u *UserView) URLSuffix(suffix string) string {
	return fmt.Sprintf("/user/%d/%s", u.ID, suffix)
}

func (u *UserView) HasAccess(al AccessLevel) bool {
	return u.AccessLevel >= al
}

func (user GetAppusersRow) ToUserView() UserView {
	return UserView{
		ID:           user.ID,
		Name:         user.DisplayName,
		Email:        user.Email,
		AvatarURL:    user.AvatarUrl.String,
		HasAvatarURL: user.AvatarUrl.Valid,
		AccessLevel:  AccessLevel(user.AccessLevel),
	}
}

func (user Appuser) ToUserView() UserView {
	return UserView{
		ID:           user.ID,
		Name:         user.DisplayName,
		Email:        user.Email,
		AvatarURL:    user.AvatarUrl.String,
		HasAvatarURL: user.AvatarUrl.Valid,
		AccessLevel:  AccessLevel(user.AccessLevel),
	}
}

func (user GetUserRow) ToUserView() UserView {
	return UserView{
		ID:           user.ID,
		Name:         user.DisplayName,
		Email:        user.Email,
		AvatarURL:    user.AvatarUrl.String,
		HasAvatarURL: user.AvatarUrl.Valid,
		AccessLevel:  AccessLevel(user.AccessLevel),
	}
}

// ---- Invitation

type InvitationView struct {
	ID      string
	Email   string
	Created time.Time
	Expires time.Time
}

func (inv InvitationView) DeleteURL() string {
	return fmt.Sprintf("/invite/%s/delete", inv.ID)
}

func (inv Invitation) ToInvitationView() InvitationView {
	return InvitationView{
		ID:      inv.ID,
		Email:   inv.Email.String,
		Expires: inv.Expires.Time,
		Created: inv.Created.Time,
	}
}

// ---- Google Drive Item

type GDriveItemView struct {
	Item GDriveItem
}

// ---- Google Drive Permission

type GDrivePermissionView struct {
	Permission GDrivePermission
	BinoUser   UserView
}

// ---- Preferred species

type SpeciesView struct {
	ID        int32
	Name      string
	Preferred bool
}

func (in GetPreferredSpeciesForHomeRow) ToSpeciesView() SpeciesView {
	return SpeciesView{
		ID:        in.SpeciesID,
		Name:      in.Name,
		Preferred: true,
	}
}

func (in GetSpeciesWithLanguageRow) ToSpeciesView(preferred bool) SpeciesView {
	return SpeciesView{
		ID:        in.SpeciesID,
		Name:      in.Name,
		Preferred: preferred,
	}
}

// ---- Period

type PeriodView struct {
	ID     int32
	HomeID int32
	From   DateView
	To     DateView
	Note   string
}

func (pv PeriodView) DeleteURL() string {
	return fmt.Sprintf("/period/%d/delete", pv.ID)
}

func (in HomeUnavailable) ToPeriodView() PeriodView {
	return PeriodView{
		ID:     in.ID,
		HomeID: in.HomeID,
		From:   DateViewFromPGDate(in.FromDate),
		To:     DateViewFromPGDate(in.ToDate),
		Note:   in.Note.String,
	}
}

type DateView struct {
	Year  int
	Month time.Month
	Day   int // 1-31
}

func DateViewFromPGDate(pg pgtype.Date) DateView {
	if pg.Valid {
		return DateViewFromTime(pg.Time)
	}
	return DateView{}
}

func DateViewFromTime(t time.Time) DateView {
	if t.IsZero() {
		return DateView{}
	}
	return DateView{
		Year:  t.Year(),
		Month: t.Month(),
		Day:   t.Day(),
	}
}

func (dv DateView) Valid() bool {
	return dv.Day != 0 && dv.Month >= 1 && dv.Month <= 12 && dv.Year > 0 && dv.Year < 10000
}

func (dv DateView) ToTime() time.Time {
	return time.Date(dv.Year, dv.Month, dv.Day, 0, 0, 0, 0, time.UTC)
}

func (dv DateView) ToPGDate() pgtype.Date {
	return pgtype.Date{
		Time:  dv.ToTime(),
		Valid: dv.Valid(),
	}
}

func (dv DateView) Before(other DateView) bool {
	if dv.Year < other.Year {
		return true
	}
	if dv.Year > other.Year {
		return false
	}
	if dv.Month < other.Month {
		return true
	}
	if dv.Month > other.Month {
		return false
	}
	return dv.Day < other.Day
}

// ---- Patient page

type PatientPageView struct {
	Patient PatientView
	Home    *HomeView
	Tags    []GetTagsWithLanguageCheckinRow
	Events  []EventView
	Homes   []HomeView
}

// ---- Event

type EventView struct {
	Row     GetEventsForPatientRow
	TimeAbs string
	TimeRel string
	Time    time.Time
	Home    HomeView
	User    UserView
}

func (ev *EventView) SetNoteURL() string {
	return fmt.Sprintf("/event/%d/set-note", ev.Row.ID)
}

// ---- Match

// ENUM(journal, patient)
type MatchType string

type MatchView struct {
	URL           string
	Type          MatchType
	HeaderRuns    []HighlightRun
	BodyFragments []HighlightFragment

	ExtraData string
}

func parseJSON[T any](extraData string) *T {
	var out T
	if err := json.Unmarshal([]byte(extraData), &out); err != nil {
		return nil
	}
	return &out
}

type HighlightFragment struct {
	Runs []HighlightRun
}

func SplitFragments(runs []HighlightRun) []HighlightFragment {
	var frags []HighlightFragment
	var current []HighlightRun

	for _, r := range runs {
		if strings.Contains(r.Text, "[CUT]") {
			parts := strings.Split(r.Text, "[CUT]")
			for i, part := range parts {
				if part != "" {
					current = append(current, HighlightRun{Text: part, Hit: r.Hit})
				}
				// every [CUT] ends the current fragment
				if i < len(parts)-1 {
					if len(current) > 0 {
						frags = append(frags, HighlightFragment{Runs: current})
						current = nil
					}
				}
			}
		} else {
			current = append(current, r)
		}
	}

	if len(current) > 0 {
		frags = append(frags, HighlightFragment{Runs: current})
	}

	return frags
}

func (in *SearchBasicRow) ToMatchView() MatchView {
	headerRuns := ParseHeadline(in.HeaderHeadline)
	bodyRuns := ParseHeadline(in.BodyHeadline)

	return MatchView{
		URL:           in.AssociatedUrl.String,
		Type:          MatchType(in.Ns),
		HeaderRuns:    headerRuns,
		BodyFragments: SplitFragments(bodyRuns),
		ExtraData:     in.ExtraData.String,
	}
}

func (in *SearchAdvancedRow) ToMatchView(q string) MatchView {
	headerRuns := ParseHeadline(in.HeaderHeadline)
	if !hasHit(headerRuns) {
		headerRuns = HighlightFallback(in.Header, q)
	}

	bodyRuns := ParseHeadline(in.BodyHeadline)
	if !hasHit(bodyRuns) {
		bodyRuns = HighlightFallback(in.Body, q)
	}

	return MatchView{
		URL:           in.AssociatedUrl.String,
		Type:          MatchType(in.Ns),
		HeaderRuns:    headerRuns,
		BodyFragments: SplitFragments(bodyRuns),
		ExtraData:     in.ExtraData.String,
	}
}

func hasHit(runs []HighlightRun) bool {
	for _, r := range runs {
		if r.Hit {
			return true
		}
	}
	return false
}

type HighlightRun struct {
	Text string
	Hit  bool
}

func ParseHeadline(s string) []HighlightRun {
	const start = "[START]"
	const stop = "[STOP]"
	var out []HighlightRun
	i := 0
	for i < len(s) {
		ix := strings.Index(s[i:], start)
		if ix < 0 {
			out = append(out, HighlightRun{Text: s[i:], Hit: false})
			break
		}
		ix += i
		if ix > i {
			out = append(out, HighlightRun{Text: s[i:ix], Hit: false})
		}
		j := strings.Index(s[ix+len(start):], stop)
		if j < 0 {
			out = append(out, HighlightRun{Text: s[ix:], Hit: false})
			break
		}
		j += ix + len(start)
		out = append(out, HighlightRun{Text: s[ix+len(start) : j], Hit: true})
		i = j + len(stop)
	}
	return out
}

func HighlightFallback(text, query string) []HighlightRun {
	const context = 40 // number of chars of context on each side
	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)
	var out []HighlightRun

	pos := 0
	for {
		i := strings.Index(lowerText[pos:], lowerQuery)
		if i < 0 {
			break
		}
		i += pos
		start := max(0, i-context)
		end := min(len(text), i+len(query)+context)

		// Add ellipsis if we skipped earlier text
		if len(out) == 0 && start > 0 {
			out = append(out, HighlightRun{Text: "[CUT]", Hit: false})
		}

		out = append(out,
			HighlightRun{Text: text[start:i], Hit: false},
			HighlightRun{Text: text[i : i+len(query)], Hit: true},
		)

		pos = end
		if pos < len(text) {
			out = append(out, HighlightRun{Text: text[i+len(query) : pos], Hit: false})
		}

		if pos < len(text) {
			out = append(out, HighlightRun{Text: "[CUT]", Hit: false})
		}

		// Limit to a few snippets
		if len(out) > 5 {
			break
		}
	}

	if len(out) == 0 {
		// no match, take leading snippet
		if len(text) > 2*context {
			return []HighlightRun{{Text: text[:2*context] + "[CUT]", Hit: false}}
		}
		return []HighlightRun{{Text: text, Hit: false}}
	}

	return out
}
