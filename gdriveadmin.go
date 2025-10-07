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

func (s *Server) folders(w http.ResponseWriter, r *http.Request) (GDriveFoldersResult, error) {
	ctx := r.Context()

	var folders GDriveFoldersResult
	if !s.CacheGet(ctx, cacheKeyGDriveDirs, &folders) {
		g, err := s.getDriveService(r)
		if err != nil {
			return folders, fmt.Errorf("connecting to Google Drive: %w", err)
		}

		folders.Folders, err = s.GetFolders(ctx, g)
		if err != nil {
			return folders, fmt.Errorf("listing folders: %w", err)
		}
		folders.Time = time.Now()

		_ = s.CacheSet(ctx, cacheKeyGDriveDirs, folders)
	}

	folders.Valid = true
	return folders, nil
}

func (s *Server) getGDriveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	var selectedDir GDriveItem
	s.CacheGet(ctx, cacheKeyGDriveBaseDir, &selectedDir)

	var selectedTemplate GDriveItem
	s.CacheGet(ctx, cacheKeyGDriveTemplate, &selectedTemplate)

	folders, err := s.folders(w, r)
	if err != nil {
		commonData.Error(commonData.User.Language.GDriveLoadFoldersFailed, err)
	}

	s.GDrivePage(ctx, commonData, folders, selectedDir, nil, selectedTemplate).Render(ctx, w)
}

func (s *Server) reloadGDriveFoldersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.Queries.CacheDelete(ctx, cacheKeyGDriveDirs)
	s.redirectToReferer(w, r)
}

func (s *Server) setGDriveBaseFolderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	newDir, err := s.getPathValue(r, "id")
	if err != nil {
		s.renderError(w, r, commonData, err)
		return
	}

	var oldDir GDriveItem
	s.CacheGet(ctx, cacheKeyGDriveBaseDir, &oldDir)

	if oldDir.ID == newDir {
		s.redirectToReferer(w, r)
		return
	}

	g, err := s.getDriveService(r)
	if err != nil {
		s.renderError(w, r, commonData, fmt.Errorf("getting drive service: %w", err))
		return
	}

	newDirItem, err := s.GetFile(ctx, g, newDir)
	if err != nil {
		s.renderError(w, r, commonData, fmt.Errorf("looking up file: %w", err))
		return
	}

	if err := s.CacheSet(ctx, cacheKeyGDriveBaseDir, newDirItem); err != nil {
		s.renderError(w, r, commonData, fmt.Errorf("updating base dir: %w", err))
		return
	}

	commonData.Success(commonData.User.Language.GDriveBaseDirUpdated)

	// Have to reset the template since it's not in the same dir anymore
	s.Queries.CacheDelete(ctx, cacheKeyGDriveTemplate)

	s.redirectToReferer(w, r)
}

func (s *Server) gdriveFindTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	var selectedDir GDriveItem
	if !s.CacheGet(ctx, cacheKeyGDriveBaseDir, &selectedDir) {
		s.renderError(w, r, commonData, errors.New(commonData.User.Language.GDriveNoBaseDirSelected))
		return
	}
	var selectedTemplate GDriveItem
	s.CacheGet(ctx, cacheKeyGDriveTemplate, &selectedTemplate)

	folders, err := s.folders(w, r)
	if err != nil {
		s.renderError(w, r, commonData, err)
		return
	}

	filter, err := s.getFormValue(r, "filter")
	if err != nil {
		s.renderError(w, r, commonData, err)
		return
	}
	g, err := s.getDriveService(r)
	if err != nil {
		s.renderError(w, r, commonData, fmt.Errorf("connecting to Google Drive: %w", err))
		return
	}

	files, err := s.GetFiles(ctx, g, selectedDir.ID, filter)
	if err != nil {
		s.renderError(w, r, commonData, fmt.Errorf("listing documents: %w", err))
		return
	}

	s.GDrivePage(ctx, commonData, folders, selectedDir, files, selectedTemplate).Render(ctx, w)
}

func (s *Server) getExtraBinoUsers(ctx context.Context, selectedDir GDriveItem) map[string]UserView {
	extraUsers := s.getUserViews(ctx)
	for _, perm := range selectedDir.Permissions {
		if slices.Contains([]string{"owner", "organizer", "fileOrganizer", "writer"}, perm.Role) {
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

	name, err := s.getFormValue(r, "name")
	if err != nil {
		s.renderError(w, r, commonData, err)
		return
	}

	s.CacheSet(ctx, cacheKeyGDriveTemplate, GDriveItem{ID: id, Name: name})

	commonData.Success(commonData.User.Language.GDriveTemplateUpdated)

	s.redirect(w, r, "/gdrive")
}
