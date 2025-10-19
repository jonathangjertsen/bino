package main

import "fmt"

// ---- Home

type HomeView struct {
	Home     Home
	Patients []PatientView
	Users    []UserView
}

func (hv HomeView) URL() string {
	return fmt.Sprintf("/home/%d", hv.Home.ID)
}

// ---- Patient

type PatientView struct {
	ID         int32
	Status     int32
	Name       string
	Species    string
	Tags       []TagView
	JournalURL string
}

func (pv PatientView) CollapseID(prefix string) string {
	return fmt.Sprintf("%spatient-collapsible-%d", prefix, pv.ID)
}

func (pv PatientView) CheckoutNoteID(prefix string) string {
	return fmt.Sprintf("%spatient-checkout-note-%d", prefix, pv.ID)
}

func (pv PatientView) URL() string {
	return fmt.Sprintf("/patient/%d", pv.ID)
}

func (pv PatientView) URLSuffix(suffix string) string {
	return fmt.Sprintf("/patient/%d/%s", pv.ID, suffix)
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
}

func (u UserView) Valid() bool {
	return u.ID > 0
}

func (u UserView) URL() string {
	return fmt.Sprintf("/user/%d", u.ID)
}

func (user GetAppusersRow) ToUserView() UserView {
	return UserView{
		ID:           user.ID,
		Name:         user.DisplayName,
		Email:        user.Email,
		AvatarURL:    user.AvatarUrl.String,
		HasAvatarURL: user.AvatarUrl.Valid,
	}
}
