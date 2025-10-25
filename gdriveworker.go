//go:generate go tool go-enum --no-iota --values
package main

import (
	"fmt"
	"log"

	"google.golang.org/api/drive/v3"
)

const (
	maxNConcurrentGDriveTaskRequests = 100
	nWorkers                         = 1
)

// ENUM(
//
//	GetFile,
//	InviteUser,
//	CreateJournal,
//
// )
type GDriveTaskRequestID int

type GDriveWorker struct {
	cfg GDriveConfig
	g   *GDrive
	c   *Cache

	in chan GDriveTaskRequest
}

type GDriveTaskRequest struct {
	Response chan GDriveTaskResponse
	Type     GDriveTaskRequestID
	Payload  any
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

type GDriveTaskResponse struct {
	Type    GDriveTaskRequestID
	Error   error
	Payload any
}

func NewGDriveWorker(cfg GDriveConfig, g *GDrive, c *Cache) *GDriveWorker {
	w := &GDriveWorker{
		cfg: cfg,
		g:   g,
		c:   c,
		in:  make(chan GDriveTaskRequest, maxNConcurrentGDriveTaskRequests),
	}

	// Warm the cache on gdrive info
	go func() {
		c.Delete("gdrive-config-info")
		_ = w.GetGDriveConfigInfo()
	}()

	for i := range nWorkers {
		go w.worker(i)
	}

	return w
}

type GDriveConfigInfo struct {
	JournalFolder GDriveItem
	TemplateDoc   GDriveJournal
}

func (w *GDriveWorker) GetGDriveConfigInfo() GDriveConfigInfo {
	var configInfo GDriveConfigInfo
	w.c.Cached("gdrive-config-info", &configInfo, func() error {
		item, err := w.g.GetFile(w.cfg.JournalFolder)
		if err != nil {
			return err
		}
		configInfo.JournalFolder = item

		doc, err := w.g.ReadDocument(w.cfg.TemplateFile)
		if err != nil {
			return err
		}
		configInfo.TemplateDoc = doc

		if err := doc.Validate(); err != nil {
			return err
		}
		log.Printf("Fetched GDrive Config info")

		return nil
	})
	return configInfo
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

func (w *GDriveWorker) worker(workerID int) {
	for {
		req := <-w.in
		log.Printf("Worker %d received request: %+v", workerID, req)
		resp := w.handleRequest(req)
		req.Response <- resp
		log.Printf("Worker %d sent back response: %+v", workerID, resp)
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
