package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"os"
	"time"

	"github.com/google/uuid"
)

// INTERFACE

type UploadResult struct {
	UniqueID       string
	Error          error
	HTTPStatusCode int
}

type DeleteResult struct {
	Error          error
	HTTPStatusCode int
}

type ReadResult struct {
	Data           []byte
	FileInfo       FileInfo
	Error          error
	HTTPStatusCode int
}

type CommitResult struct {
	Commited       []string
	Failed         []string
	Error          error
	HTTPStatusCode int
}

type FileInfo struct {
	FileName string
	MIMEInfo map[string][]string
	Size     int64
	Created  time.Time
	Creator  int32
}

func (fi FileInfo) MIMEContentType() string {
	return textproto.MIMEHeader(fi.MIMEInfo).Get("Content-Type")
}

type FileBackend interface {
	// Upload to temporary storage
	Upload(ctx context.Context, data io.Reader, fileInfo FileInfo) UploadResult
	// DeleteTemp from temporary storage
	DeleteTemp(ctx context.Context, ID string) DeleteResult
	// ReadTemp file from temporary storage
	ReadTemp(ctx context.Context, ID string) ReadResult
	// Commit files from temporary storage to real storage
	Commit(ctx context.Context, IDs []string) CommitResult
	// Read info
	ReadInfo(ctx context.Context, ID string) (FileInfo, error)
	// Open file
	Open(ctx context.Context, ID string, fileInfo FileInfo) (io.ReadCloser, error)
}

// LOCAL FILE API

type LocalFileStorage struct {
	MainDirectory string
	TmpDirectory  string
}

func NewLocalFileStorage(ctx context.Context, mainDir, tmpDir string) *LocalFileStorage {
	if err := os.MkdirAll(mainDir, os.ModePerm); err != nil {
		panic(fmt.Errorf("creating mainDir='%s': %w", mainDir, err))
	}
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		panic(fmt.Errorf("creating tmpDir='%s': %w", tmpDir, err))
	}
	return &LocalFileStorage{
		MainDirectory: mainDir,
		TmpDirectory:  tmpDir,
	}
}

func (lfs *LocalFileStorage) Upload(ctx context.Context, data io.Reader, fileInfo FileInfo) (out UploadResult) {
	id := uuid.New().String()

	// Open the file base directory
	dir, err := os.OpenRoot(lfs.TmpDirectory)
	if err != nil {
		return UploadResult{
			Error:          err,
			HTTPStatusCode: http.StatusInternalServerError,
		}
	}
	defer dir.Close()

	// Create UUID subdirectory
	if err := dir.Mkdir(id, os.ModePerm); err != nil {
		return UploadResult{
			Error:          fmt.Errorf("creating file directory: %w", err),
			HTTPStatusCode: http.StatusInternalServerError,
		}
	}

	// Create the metadata file
	metaFile, err := dir.Create(id + "/metadata.json")
	if err != nil {
		return UploadResult{
			Error:          fmt.Errorf("creating metadata.json: %w", err),
			HTTPStatusCode: http.StatusInternalServerError,
		}
	}
	defer func() {
		if err := metaFile.Close(); out.Error == nil && err != nil {
			out.Error = fmt.Errorf("closing metadata.json: %w", err)
			out.HTTPStatusCode = http.StatusInternalServerError
			out.UniqueID = ""
		}
	}()
	jsonWriter := json.NewEncoder(metaFile)
	if err := jsonWriter.Encode(fileInfo); err != nil {
		return UploadResult{
			Error:          fmt.Errorf("writing metadata.json: %w", err),
			HTTPStatusCode: http.StatusInternalServerError,
		}
	}

	// Create the file
	file, err := dir.Create(id + "/" + fileInfo.FileName)
	if err != nil {
		return UploadResult{
			Error:          fmt.Errorf("creating %s: %w", fileInfo.FileName, err),
			HTTPStatusCode: http.StatusInternalServerError,
		}
	}
	defer func() {
		if err := file.Close(); out.Error == nil && err != nil {
			out.Error = fmt.Errorf("closing file: %w", err)
			out.HTTPStatusCode = http.StatusInternalServerError
			out.UniqueID = ""
		}
	}()

	// Copy file data
	if n, err := io.Copy(file, data); err != nil {
		return UploadResult{
			Error:          fmt.Errorf("writing file contents: %w", err),
			HTTPStatusCode: http.StatusInternalServerError,
		}
	} else if n != fileInfo.Size {
		return UploadResult{
			Error:          fmt.Errorf("file size expected %d wrote %d", fileInfo.Size, n),
			HTTPStatusCode: http.StatusInternalServerError,
		}
	}

	return UploadResult{
		UniqueID:       id,
		Error:          nil,
		HTTPStatusCode: http.StatusOK,
	}
}

func (lfs *LocalFileStorage) DeleteTemp(ctx context.Context, id string) (out DeleteResult) {
	if err := uuid.Validate(id); err != nil {
		return DeleteResult{
			Error:          fmt.Errorf("'%s' is not a valid UUID: %w", id, err),
			HTTPStatusCode: http.StatusBadRequest,
		}
	}

	// Open the file base directory
	dir, err := os.OpenRoot(lfs.TmpDirectory)
	if err != nil {
		return DeleteResult{
			Error:          err,
			HTTPStatusCode: http.StatusInternalServerError,
		}
	}
	defer dir.Close()

	// Delete directory
	if err := dir.RemoveAll(id); err != nil {
		return DeleteResult{
			Error:          err,
			HTTPStatusCode: http.StatusInternalServerError,
		}
	}

	return DeleteResult{
		HTTPStatusCode: http.StatusOK,
	}
}

func (lfs *LocalFileStorage) ReadTemp(ctx context.Context, id string) (out ReadResult) {
	if err := uuid.Validate(id); err != nil {
		return ReadResult{
			Error:          fmt.Errorf("'%s' is not a valid UUID: %w", id, err),
			HTTPStatusCode: http.StatusBadRequest,
		}
	}

	dir, err := os.OpenRoot(lfs.TmpDirectory)
	if err != nil {
		return ReadResult{
			Error:          err,
			HTTPStatusCode: http.StatusInternalServerError,
		}
	}
	defer dir.Close()

	metaFile, err := dir.Open(id + "/metadata.json")
	if err != nil {
		return ReadResult{
			Error:          err,
			HTTPStatusCode: http.StatusNotFound,
		}
	}
	var info FileInfo
	if err := json.NewDecoder(metaFile).Decode(&info); err != nil {
		metaFile.Close()
		return ReadResult{
			Error:          err,
			HTTPStatusCode: http.StatusInternalServerError,
		}
	}
	metaFile.Close()

	file, err := dir.Open(id + "/" + info.FileName)
	if err != nil {
		return ReadResult{
			Error:          err,
			HTTPStatusCode: http.StatusNotFound,
		}
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return ReadResult{
			Error:          err,
			HTTPStatusCode: http.StatusInternalServerError,
		}
	}

	if stat.Size() > MaxImageSize {
		file.Close()
		return ReadResult{
			Error:          fmt.Errorf("file too large"),
			HTTPStatusCode: http.StatusRequestEntityTooLarge,
		}
	}

	data, err := io.ReadAll(file)
	file.Close()
	if err != nil {
		return ReadResult{
			Error:          err,
			HTTPStatusCode: http.StatusInternalServerError,
		}
	}

	return ReadResult{
		Data:           data,
		FileInfo:       info,
		HTTPStatusCode: http.StatusOK,
	}
}

func (lfs *LocalFileStorage) Commit(ctx context.Context, ids []string) CommitResult {
	var out CommitResult
	out.HTTPStatusCode = http.StatusOK
	for _, id := range ids {
		tmpDir := lfs.TmpDirectory + "/" + id
		mainDir := lfs.MainDirectory + "/" + id
		if err := os.Rename(tmpDir, mainDir); err != nil {
			out.Failed = append(out.Failed, id)
			out.Error = err
			out.HTTPStatusCode = http.StatusInternalServerError
		} else {
			out.Commited = append(out.Commited, id)
		}
	}
	return out
}

func (lfs *LocalFileStorage) ReadInfo(ctx context.Context, id string) (FileInfo, error) {
	dir, err := os.OpenRoot(lfs.MainDirectory)
	if err != nil {
		return FileInfo{}, err
	}
	defer dir.Close()

	metaFile, err := dir.Open(id + "/metadata.json")
	if err != nil {
		return FileInfo{}, err
	}
	var info FileInfo
	if err := json.NewDecoder(metaFile).Decode(&info); err != nil {
		metaFile.Close()
		return FileInfo{}, err
	}
	metaFile.Close()

	return info, nil
}

func (lfs *LocalFileStorage) Open(ctx context.Context, id string, info FileInfo) (io.ReadCloser, error) {
	dir, err := os.OpenRoot(lfs.MainDirectory)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	file, err := dir.Open(id + "/" + info.FileName)
	if err != nil {
		return nil, err
	}
	return file, nil
}
