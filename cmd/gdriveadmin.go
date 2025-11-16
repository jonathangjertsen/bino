package main

import (
	"context"
	"fmt"
	"maps"
	"net/http"
)

func (s *Server) getGDriveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	info := s.GDriveWorker.GetGDriveConfigInfo()
	s.GDrivePage(ctx, commonData, info).Render(ctx, w)
}

func (s *Server) getExtraBinoUsers(ctx context.Context, selectedDir GDriveItem) map[string]UserView {
	users := s.getUserViews(ctx)
	extraUsers := maps.Clone(users)
	for _, perm := range selectedDir.Permissions {
		if !perm.CanWrite() {
			continue
		}
		binoUser, ok := users[perm.Email]
		if !ok || !binoUser.Valid() {
			continue
		}
		delete(extraUsers, perm.Email)
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

	if err := s.GDriveWorker.InviteUser(s.Config.GoogleDrive.JournalFolder, email, "writer"); err != nil {
		s.renderError(w, r, commonData, fmt.Errorf("inviting user: %w", err))
		return
	}

	commonData.Success(commonData.User.Language.GDriveUserInvited)

	s.redirect(w, r, "/gdrive")
}
