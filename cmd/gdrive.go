package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"time"

	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

var (
	reDeleteImages = regexp.MustCompile(`<data:image/[a-zA-Z]+;base64,[^>]+>`)
	reUnbold       = regexp.MustCompile(`\*\*(.*?)\*\*`)
)

type GDrive struct {
	Drive     *drive.Service
	Docs      *docs.Service
	Queries   *Queries
	DriveBase string
}

type GDriveConfig struct {
	ServiceAccountKeyLocation string
	DriveBase                 string
	JournalFolder             string
	TemplateFile              string
	ExtraJournalFolders       []string
}

func NewGDriveWithServiceAccount(ctx context.Context, config GDriveConfig, queries *Queries) (*GDrive, error) {
	drive, err := drive.NewService(ctx, option.WithCredentialsFile(config.ServiceAccountKeyLocation))
	if err != nil {
		return nil, fmt.Errorf("creating Drive service: %w", err)
	}

	docs, err := docs.NewService(ctx, option.WithCredentialsFile(config.ServiceAccountKeyLocation))
	if err != nil {
		return nil, fmt.Errorf("creating Docs service: %w", err)
	}

	return &GDrive{
		Drive:     drive,
		Docs:      docs,
		Queries:   queries,
		DriveBase: config.DriveBase,
	}, nil
}

type GDriveItem struct {
	ID    string
	Name  string
	Valid bool

	// Optional
	Permissions  []GDrivePermission
	ModifiedTime time.Time
	Trashed      bool
	CreatedTime  time.Time
}

func GDriveFolderURL(id string) string {
	return "https://drive.google.com/drive/folders/" + id
}

func (item *GDriveItem) FolderURL() string {
	return GDriveFolderURL(item.ID)
}

func (item *GDriveItem) DocumentURL() string {
	return "https://docs.google.com/document/d/" + item.ID
}

func (server *Server) LoggedInUserCanShare(ctx context.Context, item GDriveItem) bool {
	cd := MustLoadCommonData(ctx)

	for _, p := range item.Permissions {
		if p.Email == cd.User.Email {
			return p.CanWrite()
		}
	}

	return false
}

func GDriveItemFromFile(f *drive.File, p *drive.PermissionList) GDriveItem {
	if f == nil {
		return GDriveItem{}
	}
	var permissions []GDrivePermission
	if p != nil {
		permissions = SliceToSlice(p.Permissions, func(p *drive.Permission) GDrivePermission {
			return GDrivePermission{
				DisplayName: p.DisplayName,
				Email:       p.EmailAddress,
				Role:        p.Role,
			}
		})
	}

	modifiedTime, _ := time.Parse(time.RFC3339, f.ModifiedTime)
	createdTime, _ := time.Parse(time.RFC3339, f.CreatedTime)

	return GDriveItem{
		ID:           f.Id,
		Name:         f.Name,
		Valid:        true,
		Permissions:  permissions,
		ModifiedTime: modifiedTime,
		CreatedTime:  createdTime,
		Trashed:      f.Trashed,
	}
}

type GDrivePermission struct {
	DisplayName string
	Email       string
	Role        string
}

func (gdp GDrivePermission) CanWrite() bool {
	return slices.Contains([]string{"owner", "organizer", "fileOrganizer", "writer"}, gdp.Role)
}

func (g *GDrive) fileToItem(file *drive.File) (GDriveItem, error) {
	call := g.Drive.Permissions.List(file.Id).Fields("permissions(displayName, emailAddress, role)")

	if g.DriveBase != "" {
		call = call.
			SupportsAllDrives(true)
	}

	pl, err := call.Do()
	if err != nil {
		return GDriveItem{}, err
	}
	item := GDriveItemFromFile(file, pl)

	return item, nil
}

func (g *GDrive) GetFile(id string) (GDriveItem, error) {
	call := g.Drive.Files.Get(id).
		Fields("id, name, capabilities")

	if g.DriveBase != "" {
		call = call.
			SupportsAllDrives(true)
	}

	f, err := call.Do()
	if err != nil {
		return GDriveItem{}, err
	}

	return g.fileToItem(f)
}

func (g *GDrive) ReadDocument(id string) (GDriveJournal, error) {
	item, err := g.GetFile(id)
	if err != nil {
		return GDriveJournal{}, err
	}

	call := g.Drive.Files.Export(id, "text/markdown")
	f, err := call.Download()
	if err != nil {
		return GDriveJournal{}, err
	}
	defer f.Body.Close()
	content, err := io.ReadAll(f.Body)
	if err != nil {
		return GDriveJournal{}, err
	}

	content = reDeleteImages.ReplaceAll(content, []byte{})
	content = reUnbold.ReplaceAll(content, []byte("$1"))

	return GDriveJournal{
		Content: string(content),
		Item:    item,
	}, nil
}

func (g *GDrive) CreateDocument(conf GDriveConfigInfo, vars GDriveTemplateVars) (GDriveItem, error) {
	call := g.Drive.Files.Copy(conf.TemplateDoc.Item.ID, &drive.File{
		Name:    vars.ApplyToString(conf.TemplateDoc.Item.Name),
		Parents: []string{conf.JournalFolder.ID},
	})

	if g.DriveBase != "" {
		call = call.
			SupportsAllDrives(true)
	}

	f, err := call.Do()
	if err != nil {
		return GDriveItem{}, err
	}

	updateCall := g.Docs.Documents.BatchUpdate(f.Id, vars.ReplaceRequests())
	_, err = updateCall.Do()
	if err != nil {
		deleteCall := g.Drive.Files.Delete(f.Id)

		if g.DriveBase != "" {
			deleteCall = deleteCall.
				SupportsAllDrives(true)
		}

		deleteErr := deleteCall.Do()
		if deleteErr != nil {
			return GDriveItem{}, errors.Join(err, deleteErr)
		}

		return GDriveItem{}, err
	}

	return g.fileToItem(f)
}
