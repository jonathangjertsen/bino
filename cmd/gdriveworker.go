//go:generate go tool go-enum --no-iota --values
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/drive/v3"
)

const (
	maxNConcurrentGDriveTaskRequests = 100
	nWorkers                         = 1
	timeFormatDriveQ                 = "2006-01-02T15:04:05"
)

// ENUM(
//
//	GetFile,
//	InviteUser,
//	CreateJournal,
//	ListFiles,
//
// )
type GDriveTaskRequestID int

type GDriveWorker struct {
	cfg GDriveConfig
	g   *GDrive

	in chan GDriveTaskRequest

	cachedInfo   *GDriveConfigInfo
	cachedInfoMu *sync.Mutex
}

type GDriveTaskRequest struct {
	Response chan GDriveTaskResponse
	Type     GDriveTaskRequestID
	Payload  any
}

func (gdtr GDriveTaskRequest) String() string {
	return fmt.Sprintf("<GDriveTaskRequest of type %s>", gdtr.Type)
}

func newGDriveTaskRequest() GDriveTaskRequest {
	return GDriveTaskRequest{
		Response: make(chan GDriveTaskResponse, 1),
	}
}

func newGDriveTaskRequestGetFile(id string) GDriveTaskRequest {
	req := newGDriveTaskRequest()
	req.Type = GDriveTaskRequestIDGetFile
	req.Payload = id
	return req
}

type payloadInviteUser struct {
	ID    string
	Email string
	Role  string
}

func newGDriveTaskRequestInviteUser(id, email, role string) GDriveTaskRequest {
	req := newGDriveTaskRequest()
	req.Type = GDriveTaskRequestIDInviteUser
	req.Payload = payloadInviteUser{
		ID:    id,
		Email: email,
		Role:  role,
	}
	return req
}

type ListFilesParams struct {
	Parent         string
	ModifiedAfter  time.Time
	ModifiedBefore time.Time
	PageToken      string
}

type ListFilesResult struct {
	Folder        GDriveItem
	Files         []GDriveItem
	NextPageToken string
}

func newGDriveTaskRequestListFiles(params ListFilesParams) GDriveTaskRequest {
	req := newGDriveTaskRequest()
	req.Type = GDriveTaskRequestIDListFiles
	req.Payload = params
	return req
}

func newGDriveTaskRequestCreateJournal(vars GDriveTemplateVars) GDriveTaskRequest {
	req := newGDriveTaskRequest()
	req.Type = GDriveTaskRequestIDCreateJournal
	req.Payload = vars
	return req
}

func (req GDriveTaskRequest) decodeGetFile() (string, error) {
	id, ok := req.Payload.(string)
	if !ok {
		return "", fmt.Errorf("decodeGetFile called on request with payload of type %T", req.Payload)
	}
	return id, nil
}

func (req GDriveTaskRequest) decodeInviteUser() (payloadInviteUser, error) {
	inv, ok := req.Payload.(payloadInviteUser)
	if !ok {
		return payloadInviteUser{}, fmt.Errorf("decodeInviteUser called on request with payload of type %T", req.Payload)
	}
	return inv, nil
}

func (req GDriveTaskRequest) decodeCreateJournal() (GDriveTemplateVars, error) {
	vars, ok := req.Payload.(GDriveTemplateVars)
	if !ok {
		return GDriveTemplateVars{}, fmt.Errorf("decodeCreateJournal called on request with payload of type %T", req.Payload)
	}
	return vars, nil
}

func (req GDriveTaskRequest) decodeListFiles() (ListFilesParams, error) {
	payload, ok := req.Payload.(ListFilesParams)
	if !ok {
		return ListFilesParams{}, fmt.Errorf("decodeListFiles called on request with payload of type %T", req.Payload)
	}
	return payload, nil
}

func (resp GDriveTaskResponse) decodeError() error {
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func (resp GDriveTaskResponse) decodeGetFile() (GDriveItem, error) {
	if err := resp.decodeError(); err != nil {
		return GDriveItem{}, err
	}
	if resp.Type != GDriveTaskRequestIDGetFile {
		return GDriveItem{}, fmt.Errorf("decodeGetFile called on response of type %s", resp.Type.String())
	}
	item, ok := resp.Payload.(GDriveItem)
	if !ok {
		return GDriveItem{}, fmt.Errorf("decodeGetFile called with bad payload type %T", resp.Payload)
	}
	return item, nil
}

func (resp GDriveTaskResponse) decodeInviteUser() error {
	if err := resp.decodeError(); err != nil {
		return err
	}
	if resp.Type != GDriveTaskRequestIDInviteUser {
		return fmt.Errorf("decodeInviteUser called on response of type %s", resp.Type.String())
	}
	return nil
}

func (resp GDriveTaskResponse) decodeCreateJournal() (GDriveItem, error) {
	if err := resp.decodeError(); err != nil {
		return GDriveItem{}, err
	}
	if resp.Type != GDriveTaskRequestIDCreateJournal {
		return GDriveItem{}, fmt.Errorf("decodeCreateJournal called on response of type %s", resp.Type.String())
	}
	item, ok := resp.Payload.(GDriveItem)
	if !ok {
		return GDriveItem{}, fmt.Errorf("decodeCreateJournal called with bad payload type %T", resp.Payload)
	}
	return item, nil
}

func (resp GDriveTaskResponse) decodeListFiles() (ListFilesResult, error) {
	if err := resp.decodeError(); err != nil {
		return ListFilesResult{}, err
	}
	if resp.Type != GDriveTaskRequestIDListFiles {
		return ListFilesResult{}, fmt.Errorf("decodeListFiles called on response of type %s", resp.Type.String())
	}
	result, ok := resp.Payload.(ListFilesResult)
	if !ok {
		return ListFilesResult{}, fmt.Errorf("decodeListFiles called with bad payload type %T", resp.Payload)
	}
	return result, nil
}

type GDriveTaskResponse struct {
	Type    GDriveTaskRequestID
	Error   error
	Payload any
}

func (gdtr GDriveTaskResponse) String() string {
	return fmt.Sprintf("<GDriveTaskResponse of type %s>", gdtr.Type)
}

func NewGDriveWorker(ctx context.Context, cfg GDriveConfig, g *GDrive) *GDriveWorker {
	w := &GDriveWorker{
		cfg:          cfg,
		g:            g,
		in:           make(chan GDriveTaskRequest, maxNConcurrentGDriveTaskRequests),
		cachedInfoMu: &sync.Mutex{},
	}

	// Warm the cache on gdrive info, then start pollin
	go func() {
		_ = w.GetGDriveConfigInfo()

		w.searchIndexWorker(ctx)
	}()

	// Workers
	for i := range nWorkers {
		go w.worker(i)
	}

	return w
}

type GDriveConfigInfo struct {
	JournalFolder GDriveItem
	TemplateDoc   GDriveJournal
	ExtraFolders  []GDriveItem
}

func (w *GDriveWorker) GetGDriveConfigInfo() GDriveConfigInfo {
	w.cachedInfoMu.Lock()
	defer w.cachedInfoMu.Unlock()
	if w.cachedInfo == nil {
		w.cachedInfo = new(GDriveConfigInfo)
		item, err := w.g.GetFile(w.cfg.JournalFolder)
		if err != nil {
			panic(err)
		}
		w.cachedInfo.JournalFolder = item

		doc, err := w.g.ReadDocument(w.cfg.TemplateFile)
		if err != nil {
			panic(err)
		}
		w.cachedInfo.TemplateDoc = doc

		if err := doc.Validate(); err != nil {
			panic(err)
		}

		for _, id := range w.cfg.ExtraJournalFolders {
			folder, err := w.g.GetFile(id)
			if err != nil {
				log.Printf("error getting extra folder %s: %v", id, err)
				continue
			}
			w.cachedInfo.ExtraFolders = append(
				w.cachedInfo.ExtraFolders,
				folder,
			)
		}

		log.Printf("Fetched GDrive Config info")
	}
	return *w.cachedInfo
}

func (w *GDriveWorker) Exec(req GDriveTaskRequest) GDriveTaskResponse {
	w.in <- req
	return <-req.Response
}

func (w *GDriveWorker) GetFile(id string) (GDriveItem, error) {
	return w.Exec(newGDriveTaskRequestGetFile(id)).decodeGetFile()
}

func (w *GDriveWorker) InviteUser(id, email, role string) error {
	return w.Exec(newGDriveTaskRequestInviteUser(id, email, role)).decodeInviteUser()
}

func (w *GDriveWorker) CreateJournal(vars GDriveTemplateVars) (GDriveItem, error) {
	return w.Exec(newGDriveTaskRequestCreateJournal(vars)).decodeCreateJournal()
}

func (w *GDriveWorker) ListFiles(params ListFilesParams) (ListFilesResult, error) {
	return w.Exec(newGDriveTaskRequestListFiles(params)).decodeListFiles()
}

func (w *GDriveWorker) worker(workerID int) {
	for {
		req := <-w.in
		log.Printf("Worker %d received request: %s", workerID, req)
		resp := w.handleRequest(req)
		req.Response <- resp
		log.Printf("Worker %d sent back response: %s", workerID, resp)
	}
}

func (w *GDriveWorker) handleRequest(req GDriveTaskRequest) GDriveTaskResponse {
	switch req.Type {
	case GDriveTaskRequestIDGetFile:
		return w.handleRequestGetFile(req)
	case GDriveTaskRequestIDInviteUser:
		return w.handleRequestInviteUser(req)
	case GDriveTaskRequestIDCreateJournal:
		return w.handleRequestCreateJournal(req)
	case GDriveTaskRequestIDListFiles:
		return w.handleRequestListFiles(req)
	}
	return w.errorResponse(req, fmt.Errorf("unknown request type"))
}

func (w *GDriveWorker) handleRequestGetFile(req GDriveTaskRequest) GDriveTaskResponse {
	id, err := req.decodeGetFile()
	if err != nil {
		return w.errorResponse(req, err)
	}

	item, err := w.g.GetFile(id)
	if err != nil {
		return w.errorResponse(req, err)
	}

	return w.successResponse(req, item)
}

func (w *GDriveWorker) handleRequestInviteUser(req GDriveTaskRequest) GDriveTaskResponse {
	payload, err := req.decodeInviteUser()
	if err != nil {
		return w.errorResponse(req, err)
	}

	call := w.g.Drive.Permissions.Create(payload.ID, &drive.Permission{
		Type:         "user",
		EmailAddress: payload.Email,
		Role:         payload.Role,
	}).SendNotificationEmail(true)

	if w.g.DriveBase != "" {
		call = call.
			SupportsAllDrives(true)
	}

	_, err = call.Do()
	if err != nil {
		return w.errorResponse(req, err)
	}

	return w.successResponse(req, nil)
}

func (w *GDriveWorker) handleRequestCreateJournal(req GDriveTaskRequest) GDriveTaskResponse {
	vars, err := req.decodeCreateJournal()
	if err != nil {
		return w.errorResponse(req, err)
	}
	info := w.GetGDriveConfigInfo()
	item, err := w.g.CreateDocument(info, vars)
	if err != nil {
		return w.errorResponse(req, err)
	}
	return w.successResponse(req, item)
}

func (w *GDriveWorker) handleRequestListFiles(req GDriveTaskRequest) GDriveTaskResponse {
	params, err := req.decodeListFiles()
	if err != nil {
		return w.errorResponse(req, err)
	}

	folderItem, err := w.g.GetFile(params.Parent)
	if err != nil {
		return w.errorResponse(req, err)
	}

	call := w.g.Drive.Files.List()

	if w.g.DriveBase != "" {
		call = call.
			SupportsAllDrives(true)
		call = call.DriveId(w.cfg.DriveBase)
		call = call.IncludeItemsFromAllDrives(true)
		call = call.Corpora("drive")
	}

	rules := []string{
		"mimeType = 'application/vnd.google-apps.document'",
		fmt.Sprintf("'%s' in parents", params.Parent),
	}

	if !params.ModifiedAfter.IsZero() {
		rules = append(rules, fmt.Sprintf("modifiedTime > '%s'", params.ModifiedAfter.Format(timeFormatDriveQ)))
	}
	if !params.ModifiedBefore.IsZero() {
		rules = append(rules, fmt.Sprintf("modifiedTime < '%s'", params.ModifiedBefore.Format(timeFormatDriveQ)))
	}

	q := strings.Join(rules, " and ")
	call = call.Q(q)

	if params.PageToken != "" {
		call = call.PageToken(params.PageToken)
	}

	call = call.OrderBy("modifiedTime desc")
	call = call.Fields("files(id, name, modifiedTime, createdTime, trashed), nextPageToken")

	fileList, err := call.Do()
	if err != nil {
		return w.errorResponse(req, err)
	}

	return GDriveTaskResponse{
		Type: GDriveTaskRequestIDListFiles,
		Payload: ListFilesResult{
			Folder: folderItem,
			Files: SliceToSlice(fileList.Files, func(in *drive.File) GDriveItem {
				return GDriveItemFromFile(in, nil)
			}),
			NextPageToken: fileList.NextPageToken,
		},
		Error: nil,
	}
}

func (w *GDriveWorker) errorResponse(req GDriveTaskRequest, err error) GDriveTaskResponse {
	return GDriveTaskResponse{
		Type:    req.Type,
		Error:   err,
		Payload: nil,
	}
}

func (w *GDriveWorker) successResponse(req GDriveTaskRequest, obj any) GDriveTaskResponse {
	return GDriveTaskResponse{
		Type:    req.Type,
		Error:   nil,
		Payload: obj,
	}
}
