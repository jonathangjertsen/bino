package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

type GoogleDriveConfig struct {
	ClientID string
}

func (s *Server) folders(w http.ResponseWriter, r *http.Request) (GDriveFoldersResult, error) {
	ctx := r.Context()

	var folders GDriveFoldersResult
	if !s.CacheGet(ctx, cacheKeyGDriveDirs, &folders) {
		srv, err := s.getDriveService(r)
		if err != nil {
			return folders, fmt.Errorf("connecting to Google Drive: %w", err)
		}

		folders.Folders, err = srv.GetFolders(ctx)
		if err != nil {
			return folders, fmt.Errorf("listing folders: %w", err)
		}
		folders.Time = time.Now()

		_ = s.CacheSet(ctx, cacheKeyGDriveDirs, folders)
	}

	return folders, nil
}

func (s *Server) getGDriveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	var selectedDir string
	s.CacheGet(ctx, cacheKeyGDriveBaseDir, &selectedDir)

	var selectedTemplate GDriveFile
	s.CacheGet(ctx, cacheKeyGDriveTemplate, &selectedTemplate)

	folders, err := s.folders(w, r)
	if err != nil {
		s.renderError(w, r, commonData, err)
		return
	}

	GDrivePage(commonData, folders, selectedDir, nil, selectedTemplate).Render(ctx, w)
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

	var oldDir string
	s.CacheGet(ctx, cacheKeyGDriveBaseDir, &oldDir)

	if oldDir != newDir {
		if err := s.CacheSet(ctx, cacheKeyGDriveBaseDir, newDir); err != nil {
			s.renderError(w, r, commonData, fmt.Errorf("updating base dir: %w", err))
			return
		}

		// Have to reset the template since it's not in the same dir anymore
		s.Queries.CacheDelete(ctx, cacheKeyGDriveTemplate)
	}

	s.redirectToReferer(w, r)
}

func (s *Server) gdriveFindTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	var selectedDir string
	if !s.CacheGet(ctx, cacheKeyGDriveBaseDir, &selectedDir) {
		s.renderError(w, r, commonData, errors.New(commonData.User.Language.GDriveNoBaseDirSelected))
		return
	}
	var selectedTemplate GDriveFile
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
	srv, err := s.getDriveService(r)
	if err != nil {
		s.renderError(w, r, commonData, fmt.Errorf("connecting to Google Drive: %w", err))
		return
	}

	files, err := srv.GetFiles(ctx, selectedDir, filter)
	if err != nil {
		s.renderError(w, r, commonData, fmt.Errorf("listing documents: %w", err))
		return
	}

	GDrivePage(commonData, folders, selectedDir, files, selectedTemplate).Render(ctx, w)
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

	s.CacheSet(ctx, cacheKeyGDriveTemplate, GDriveFile{ID: id, Name: name})

	s.redirect(w, r, "/gdrive")
}
