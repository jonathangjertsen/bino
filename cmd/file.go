//go:generate go tool go-enum --no-iota --values
package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	MaxImageSize = 20 * 1024
)

// ENUM(Personal=0, Internal=1, Public=2)
type FileAccessibility int32

func (server *Server) fileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	name, err := server.getPathValue(r, "name")
	if err != nil {
		ajaxError(w, r, err, http.StatusNotFound)
		return
	}
	ext := filepath.Ext(name)
	idStr := strings.TrimSuffix(name, ext)
	id64, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		ajaxError(w, r, err, http.StatusNotFound)
		return
	}
	id := int32(id64)

	file, err := server.Queries.GetFileByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ajaxError(w, r, err, http.StatusNotFound)
		} else {
			ajaxError(w, r, err, http.StatusInternalServerError)
		}
		return
	}

	fileView := file.ToFileView()

	switch fileView.Accessibility {
	case FileAccessibilityPublic:
	case FileAccessibilityInternal:
		_, err := LoadCommonData(ctx)
		if err != nil {
			ajaxError(w, r, err, http.StatusUnauthorized)
			return
		}
	case FileAccessibilityPersonal:
		data, err := LoadCommonData(ctx)
		if err != nil || data.User.AppuserID != fileView.Creator {
			ajaxError(w, r, err, http.StatusUnauthorized)
			return
		}
	}

	rc, err := server.FileBackend.Open(ctx, fileView.UUID, fileView.FileInfo())
	if err != nil {
		ajaxError(w, r, err, http.StatusInternalServerError)
		return
	}
	defer rc.Close()
	w.Header().Set("Content-Type", fileView.MIMEType)
	w.Header().Set("Content-Length", strconv.Itoa(int(fileView.Size)))
	if _, err := io.Copy(w, rc); err != nil {
		LogCtx(ctx, "failed to write out file: %w", err)
	}
}

func (server *Server) filePage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := MustLoadCommonData(ctx)

	files, err := server.Queries.GetFilesForUser(ctx, GetFilesForUserParams{
		Creator:       data.User.AppuserID,
		Accessibility: int32(FileAccessibilityPersonal),
	})
	if err != nil {
		data.Error(data.User.Language.TODO("Failed to load files"), err)
		files = nil
	}

	_ = FileUploadPage(data, SliceToSlice(files, func(in File) FileView { return in.ToFileView() })).Render(ctx, w)
}

func (server *Server) filepondSubmit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := MustLoadCommonData(ctx)

	uuids, err := server.getFormMultiValue(r, "filepond")
	if err != nil {
		data.Error(data.User.Language.GenericFailed, err)
		server.redirect(w, r, "/file")
		return
	}

	if result := server.FileBackend.Commit(ctx, uuids); result.Error != nil {
		data.Error(data.User.Language.GenericFailed, err)
		server.redirect(w, r, "/file")
		return
	}

	ids, err := server.Queries.RegisterFiles(ctx, RegisterFilesParams{
		Uuids:         uuids,
		Creator:       data.User.AppuserID,
		Created:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
		Accessibility: int32(FileAccessibilityInternal),
	})
	if err != nil {
		data.Error(data.User.Language.GenericFailed, err)
		server.redirect(w, r, "/file")
		return
	}

	data.Success(data.User.Language.TODO(fmt.Sprintf("uploaded %d images", len(ids))))

	server.redirect(w, r, "/file")
}

// https://pqina.nl/filepond/docs/api/server/#process
func (server *Server) filepondProcess(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := MustLoadCommonData(ctx)

	// Parse multipart form with reasonable max memory
	err := r.ParseMultipartForm(MaxImageSize)
	if err != nil {
		ajaxError(w, r, err, http.StatusRequestEntityTooLarge)
		return
	}

	file, header, err := r.FormFile("filepond")
	if err != nil {
		ajaxError(w, r, err, http.StatusBadRequest)
		return
	}
	defer file.Close()

	result := server.FileBackend.Upload(ctx, file, FileInfo{
		FileName: header.Filename,
		MIMEType: header.Header.Get("Content-Type"),
		Size:     header.Size,
		Creator:  data.User.AppuserID,
		Created:  time.Now(),
	})
	if result.Error != nil {
		ajaxError(w, r, err, result.HTTPStatusCode)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(result.UniqueID))
}

// https://pqina.nl/filepond/docs/api/server/#revert
func (server *Server) imageFilepondRevert(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	in, err := io.ReadAll(r.Body)
	if err != nil {
		ajaxError(w, r, err, http.StatusBadRequest)
		return
	}

	if result := server.FileBackend.DeleteTemp(ctx, string(in)); result.Error != nil {
		ajaxError(w, r, err, result.HTTPStatusCode)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
}

// https://pqina.nl/filepond/docs/api/server/#restore
func (server *Server) imageFilepondRestore(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := server.getPathValue(r, "id")
	if err != nil {
		ajaxError(w, r, err, http.StatusInternalServerError)
	}

	res := server.FileBackend.ReadTemp(ctx, id)
	if res.Error != nil {
		ajaxError(w, r, res.Error, res.HTTPStatusCode)
		return
	}
	if res.Error != nil {
		ajaxError(w, r, res.Error, res.HTTPStatusCode)
		return
	}

	w.Header().Set("Content-Type", res.FileInfo.MIMEType)
	w.Header().Set("Content-Length", strconv.Itoa(len(res.Data)))
	w.Write(res.Data)
}

func (server *Server) fileDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := MustLoadCommonData(ctx)

	id, err := server.getPathID(r, "id")
	if err != nil {
		data.Error(data.User.Language.GenericFailed, err)
		server.redirectToReferer(w, r)
		return
	}

	file, err := server.Queries.GetFileByID(ctx, id)
	if err != nil {
		data.Error(data.User.Language.GenericNotFound, err)
		server.redirectToReferer(w, r)
		return
	}

	if file.Creator != data.User.AppuserID {
		data.Error(data.User.Language.GenericUnauthorized, err)
		server.redirectToReferer(w, r)
		return
	}

	if err := server.Queries.DeregisterFile(ctx, id); err != nil {
		data.Error(data.User.Language.GenericFailed, err)
		server.redirectToReferer(w, r)
		return
	}

	if result := server.FileBackend.Delete(ctx, file.Uuid); result.Error != nil {
		ajaxError(w, r, err, result.HTTPStatusCode)
		return
	}

	data.Success(data.User.Language.GenericSuccess)
	server.redirectToReferer(w, r)
}
