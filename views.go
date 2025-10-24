package main

import (
	"fmt"
	"time"

	"google.golang.org/api/drive/v3"
)

// ---- Home

type HomeView struct {
	Home Home

	// Optional
	Patients []PatientView
	Users    []UserView
}

func (hv HomeView) URL() string {
	return fmt.Sprintf("/home/%d", hv.Home.ID)
}

func (h Home) ToHomeView() HomeView {
	return HomeView{
		Home: h,
	}
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

func (in GetFormerPatientsRow) ToPatientView() PatientView {
	return PatientView{
		ID:      in.ID,
		Status:  in.Status,
		Name:    in.Name,
		Species: in.Species,
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

	// Optional
	Homes []HomeView
}

func (u UserView) Valid() bool {
	return u.ID > 0
}

func (u UserView) URL() string {
	return fmt.Sprintf("/user/%d", u.ID)
}

func (u UserView) URLSuffix(suffix string) string {
	return fmt.Sprintf("/user/%d/%s", u.ID, suffix)
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

func (user Appuser) ToUserView() UserView {
	return UserView{
		ID:           user.ID,
		Name:         user.DisplayName,
		Email:        user.Email,
		AvatarURL:    user.AvatarUrl.String,
		HasAvatarURL: user.AvatarUrl.Valid,
	}
}

func (user GetUserRow) ToUserView() UserView {
	return UserView{
		ID:           user.ID,
		Name:         user.DisplayName,
		Email:        user.Email,
		AvatarURL:    user.AvatarUrl.String,
		HasAvatarURL: user.AvatarUrl.Valid,
	}
}

// ---- User

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
	Item                       GDriveItem
	RequestingUser             int32
	RequestingUserCapabilities drive.FileCapabilities
	Permissions                []GDrivePermissionView
}

func (gdf GDriveItemView) HTMLIDSelectBaseFolder(prefix string) string {
	return fmt.Sprintf("%sset-base-folder-%s", prefix, gdf.Item.ID)
}

func (gdf GDriveItemView) URLSelectBaseFolder() string {
	return fmt.Sprintf("/gdrive/set-base-folder/%s", gdf.Item.ID)
}

func (gdf GDriveItemView) HTMLIDSelectTemplate(prefix string) string {
	return fmt.Sprintf("%sset-template-%s", prefix, gdf.Item.ID)
}

func (gdf GDriveItemView) URLSelectTemplate() string {
	return fmt.Sprintf("/gdrive/set-template/%s", gdf.Item.ID)
}

// ---- Google Drive Permission

type GDrivePermissionView struct {
	Permission GDrivePermission
	BinoUser   UserView
}
