package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"
)

type GoogleDriveConfig struct {
	ClientID string
}

func (s *Server) folders(w http.ResponseWriter, r *http.Request, g *GDrive) (GDriveFoldersResult, error) {
	ctx := r.Context()

	var folders GDriveFoldersResult
	var err error
	folders.Folders, err = s.GetFolders(ctx, g)
	if err != nil {
		return folders, fmt.Errorf("listing folders: %w", err)
	}
	folders.Time = time.Now()
	folders.Valid = true
	return folders, nil
}

func (s *Server) getGDriveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	g, err := s.getDriveService(r)
	if err != nil {
		s.renderError(w, r, commonData, fmt.Errorf("getting drive service: %w", err))
		return
	}

	var selectedDir string
	s.CacheGet(ctx, cacheKeyGDriveBaseDir, &selectedDir)

	var selectedDirFile GDriveItem
	if selectedDir != "" {
		var err error
		selectedDirFile, err = s.GetFile(ctx, g, selectedDir)
		if err != nil {
			commonData.Error(commonData.User.Language.GDriveLoadFoldersFailed, err)
		}
	}

	var selectedTemplate string
	s.CacheGet(ctx, cacheKeyGDriveTemplate, &selectedTemplate)

	var selectedTemplateFile GDriveItem
	if selectedTemplate != "" {
		var err error
		selectedTemplateFile, err = s.GetFile(ctx, g, selectedTemplate)
		if err != nil {
			commonData.Error(commonData.User.Language.GDriveLoadFoldersFailed, err)
		}
	}

	folders, err := s.folders(w, r, g)
	if err != nil {
		commonData.Error(commonData.User.Language.GDriveLoadFoldersFailed, err)
	}

	s.GDrivePage(ctx, commonData, folders, selectedDirFile, nil, selectedTemplateFile).Render(ctx, w)
}

func (s *Server) setGDriveBaseFolderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	newDir, err := s.getPathValue(r, "id")
	if err != nil {
		s.renderError(w, r, commonData, err)
		return
	}

	var oldDir string
	s.CacheGet(ctx, cacheKeyGDriveBaseDir, &oldDir)

	if oldDir == newDir {
		s.redirectToReferer(w, r)
		return
	}

	if err := s.CacheSet(ctx, cacheKeyGDriveBaseDir, newDir); err != nil {
		s.renderError(w, r, commonData, fmt.Errorf("updating base dir: %w", err))
		return
	}

	commonData.Success(commonData.User.Language.GDriveBaseDirUpdated)

	// Have to reset the template since it's not in the same dir anymore
	s.Queries.CacheDelete(ctx, cacheKeyGDriveTemplate)
	s.Queries.ClearAllUserGDriveAccess(ctx)

	s.redirectToReferer(w, r)
}

func (s *Server) gdriveFindTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	g, err := s.getDriveService(r)
	if err != nil {
		s.renderError(w, r, commonData, fmt.Errorf("connecting to Google Drive: %w", err))
		return
	}

	var selectedDir string
	s.CacheGet(ctx, cacheKeyGDriveBaseDir, &selectedDir)

	var selectedDirFile GDriveItem
	if selectedDir != "" {
		var err error
		selectedDirFile, err = s.GetFile(ctx, g, selectedDir)
		if err != nil {
			commonData.Error(commonData.User.Language.GDriveLoadFoldersFailed, err)
		}
	}

	var selectedTemplate string
	s.CacheGet(ctx, cacheKeyGDriveTemplate, &selectedTemplate)

	var selectedTemplateFile GDriveItem
	if selectedTemplate != "" {
		var err error
		selectedTemplateFile, err = s.GetFile(ctx, g, selectedTemplate)
		if err != nil {
			commonData.Error(commonData.User.Language.GDriveLoadFoldersFailed, err)
		}
	}

	folders, err := s.folders(w, r, g)
	if err != nil {
		s.renderError(w, r, commonData, err)
		return
	}

	filter, err := s.getFormValue(r, "filter")
	if err != nil {
		s.renderError(w, r, commonData, err)
		return
	}

	files, err := s.GetFiles(ctx, g, selectedDir, filter)
	if err != nil {
		s.renderError(w, r, commonData, fmt.Errorf("listing documents: %w", err))
		return
	}

	s.GDrivePage(ctx, commonData, folders, selectedDirFile, files, selectedTemplateFile).Render(ctx, w)
}

func (s *Server) getExtraBinoUsers(ctx context.Context, selectedDir GDriveItem) map[string]UserView {
	extraUsers := s.getUserViews(ctx)
	for _, perm := range selectedDir.Permissions {
		hasAccess := slices.Contains([]string{"owner", "organizer", "fileOrganizer", "writer"}, perm.Role)
		if hasAccess && perm.BinoUser.Valid() {
			delete(extraUsers, perm.BinoUser.Email)
		}
	}
	return extraUsers
}

func (s *Server) gdriveSetTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	id, err := s.getPathValue(r, "id")
	if err != nil {
		s.renderError(w, r, commonData, err)
		return
	}

	g, err := s.getDriveService(r)
	if err != nil {
		s.renderError(w, r, commonData, fmt.Errorf("connecting to Google Drive: %w", err))
		return
	}

	gdriveItem, err := s.GetFile(ctx, g, id)
	if err != nil {
		s.renderError(w, r, commonData, fmt.Errorf("getting selected template file: %w", err))
		return
	}
	if s.Transaction(ctx, func(ctx context.Context, q *Queries) error {
		extraUsers := s.getUserViews(ctx)
		for _, perm := range gdriveItem.Permissions {
			hasAccess := slices.Contains([]string{"owner", "organizer", "fileOrganizer", "writer", "commenter", "reader"}, perm.Role)
			if hasAccess && perm.BinoUser.Valid() {
				if err := q.SetUserGDriveAccess(ctx, SetUserGDriveAccessParams{
					ID:              perm.BinoUser.ID,
					HasGdriveAccess: true,
				}); err != nil {
					return err
				}
				delete(extraUsers, perm.BinoUser.Email)
			}
		}
		for _, user := range extraUsers {
			if err := q.SetUserGDriveAccess(ctx, SetUserGDriveAccessParams{
				ID:              user.ID,
				HasGdriveAccess: false,
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		s.renderError(w, r, commonData, fmt.Errorf("updating users with access: %w", err))
		return
	}

	s.CacheSet(ctx, cacheKeyGDriveTemplate, id)

	commonData.Success(commonData.User.Language.GDriveTemplateUpdated)

	s.redirect(w, r, "/gdrive")
}

func (s *Server) gdriveInviteUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	email, err := s.getPathValue(r, "email")
	if err != nil {
		s.renderError(w, r, commonData, err)
		return
	}

	var selectedDir string
	if !s.CacheGet(ctx, cacheKeyGDriveBaseDir, &selectedDir) {
		s.renderError(w, r, commonData, errors.New(commonData.User.Language.GDriveNoBaseDirSelected))
		return
	}

	g, err := s.getDriveService(r)
	if err != nil {
		s.renderError(w, r, commonData, fmt.Errorf("connecting to Google Drive: %w", err))
		return
	}

	if err := s.InviteUser(ctx, g, selectedDir, email); err != nil {
		s.renderError(w, r, commonData, fmt.Errorf("inviting user: %w", err))
		return
	}

	commonData.Success(commonData.User.Language.GDriveUserInvited)

	s.redirect(w, r, "/gdrive")
}
