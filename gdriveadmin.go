package main

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"google.golang.org/api/drive/v3"
)

func (s *Server) getGDriveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	g, err := s.getDriveService(r)
	if err != nil {
		s.renderError(w, r, commonData, fmt.Errorf("getting drive service: %w", err))
		return
	}

	var selectedDirFile GDriveItemView
	selectedDirFile, err = s.GetFile(ctx, g, s.Config.GoogleDrive.JournalFolder)
	if err != nil {
		commonData.Error(commonData.User.Language.GDriveLoadFoldersFailed, err)
	}

	var selectedTemplateFile GDriveItemView
	selectedTemplateFile, err = s.GetFile(ctx, g, s.Config.GoogleDrive.TemplateFile)
	if err != nil {
		commonData.Error(commonData.User.Language.GDriveLoadFoldersFailed, err)
	}

	s.GDrivePage(ctx, commonData, selectedDirFile, nil, selectedTemplateFile).Render(ctx, w)
}

func (s *Server) getExtraBinoUsers(ctx context.Context, selectedDir GDriveItemView) map[string]UserView {
	extraUsers := s.getUserViews(ctx)
	for _, perm := range selectedDir.Permissions {
		hasAccess := slices.Contains([]string{"owner", "organizer", "fileOrganizer", "writer"}, perm.Permission.Role)
		if hasAccess && perm.BinoUser.Valid() {
			delete(extraUsers, perm.BinoUser.Email)
		}
	}
	return extraUsers
}

func (s *Server) gdriveInviteUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	email, err := s.getPathValue(r, "email")
	if err != nil {
		s.renderError(w, r, commonData, err)
		return
	}

	g, err := s.getDriveService(r)
	if err != nil {
		s.renderError(w, r, commonData, fmt.Errorf("connecting to Google Drive: %w", err))
		return
	}

	if err := s.InviteUser(ctx, g, s.Config.GoogleDrive.JournalFolder, email); err != nil {
		s.renderError(w, r, commonData, fmt.Errorf("inviting user: %w", err))
		return
	}

	commonData.Success(commonData.User.Language.GDriveUserInvited)

	s.redirect(w, r, "/gdrive")
}

func (server *Server) GetFile(ctx context.Context, g *GDrive, id string) (GDriveItemView, error) {
	item, err := g.GetFile(id)
	if err != nil {
		return GDriveItemView{}, err
	}
	return GDriveItemViewFromFile(ctx, item, item.file, server.getUserViews(ctx)), nil
}

func (server *Server) InviteUser(ctx context.Context, g *GDrive, file string, email string) error {
	item, err := server.GetFile(ctx, g, file)
	if err != nil {
		return err
	}
	call := g.Service.Permissions.Create(item.Item.ID, &drive.Permission{
		Type:         "user",
		EmailAddress: email,
		Role:         "writer",
	}).SendNotificationEmail(true)

	if g.DriveBase != "" {
		call = call.
			SupportsAllDrives(true)
	}

	_, err = call.Do()
	if err != nil {
		return err
	}

	return nil
}

func GDriveItemViewFromFile(ctx context.Context, item GDriveItem, f *drive.File, users map[string]UserView) GDriveItemView {
	cd := MustLoadCommonData(ctx)
	if !item.Valid {
		return GDriveItemView{}
	}
	var capabilities drive.FileCapabilities
	if f.Capabilities != nil {
		capabilities = *f.Capabilities
	}
	return GDriveItemView{
		Item: item,
		Permissions: SliceToSlice(item.Permissions, func(p GDrivePermission) GDrivePermissionView {
			var binoUser UserView
			if users != nil {
				binoUser = users[p.Email]
			}
			return GDrivePermissionView{
				Permission: GDrivePermission{
					DisplayName: p.DisplayName,
					Email:       p.Email,
					Role:        p.Role,
				},
				BinoUser: binoUser,
			}
		}),
		RequestingUserCapabilities: capabilities,
		RequestingUser:             cd.User.AppuserID,
	}
}
