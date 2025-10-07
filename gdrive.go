package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const cacheKeyGDriveDirs = "gdrive-dirs"
const cacheKeyGDriveTemplate = "gdrive-template"
const cacheKeyGDriveBaseDir = "gdrive-base-dir"

var GoogleDriveScopes = []string{
	"https://www.googleapis.com/auth/drive",
}

func (server *Server) getDriveService(r *http.Request) (*GDrive, error) {
	token, err := server.getTokenFromSession(r)
	if err != nil {
		return nil, err
	}

	client := server.OAuthConfig.Client(r.Context(), token)
	srv, err := drive.NewService(r.Context(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create drive service: %v", err)
	}

	return &GDrive{
		Service:   srv,
		Queries:   server.Queries,
		DriveBase: server.Config.Auth.DriveBase,
	}, nil
}

type GDrive struct {
	Service   *drive.Service
	Queries   *Queries
	DriveBase string
}

type GDriveFoldersResult struct {
	Time    time.Time
	Folders []GDriveItem
	Valid   bool
}

type GDriveItem struct {
	ID          string
	Name        string
	Permissions []GDrivePermission
	Valid       bool
}

func GDriveItemFromFile(f *drive.File, p *drive.PermissionList, users map[string]UserView) GDriveItem {
	if f == nil {
		return GDriveItem{}
	}
	return GDriveItem{
		ID:          f.Id,
		Name:        f.Name,
		Permissions: GDrivePermissionFromPermissionList(p, users),
		Valid:       true,
	}
}

type GDrivePermission struct {
	DisplayName string
	Email       string
	Role        string
	BinoUser    UserView
}

func GDrivePermissionFromPermissionList(p *drive.PermissionList, users map[string]UserView) []GDrivePermission {
	return SliceToSlice(p.Permissions, func(p *drive.Permission) GDrivePermission {
		var binoUser UserView
		if users != nil {
			binoUser = users[p.EmailAddress]
		}
		return GDrivePermission{
			DisplayName: p.DisplayName,
			Email:       p.EmailAddress,
			Role:        p.Role,
			BinoUser:    binoUser,
		}
	})
}

func (gdf GDriveItem) HTMLIDSelectBaseFolder(prefix string) string {
	return fmt.Sprintf("%sset-base-folder-%s", prefix, gdf.ID)
}

func (gdf GDriveItem) URLSelectBaseFolder() string {
	return fmt.Sprintf("/gdrive/set-base-folder/%s", gdf.ID)
}

func (gdf GDriveItem) HTMLIDSelectTemplate(prefix string) string {
	return fmt.Sprintf("%sset-template-%s", prefix, gdf.ID)
}

func (gdf GDriveItem) URLSelectTemplate() string {
	return fmt.Sprintf("/gdrive/set-template/%s", gdf.ID)
}

func (server *Server) GetFolders(ctx context.Context, g *GDrive) ([]GDriveItem, error) {
	call := g.Service.Files.List().
		Q("mimeType = 'application/vnd.google-apps.folder'").
		PageSize(100).
		OrderBy("createdTime desc").
		Fields("files(id, name)")

	if g.DriveBase != "" {
		call = call.
			Corpora("drive").
			IncludeItemsFromAllDrives(true).
			DriveId(g.DriveBase).
			SupportsAllDrives(true)
	}

	fl, err := call.Do()
	if err != nil {
		return nil, err
	}

	return SliceToSliceErr(fl.Files, func(file *drive.File) (GDriveItem, error) {
		return server.fileToItem(ctx, g, file)
	})
}

func (server *Server) fileToItem(ctx context.Context, g *GDrive, file *drive.File) (GDriveItem, error) {
	call := g.Service.Permissions.List(file.Id).Fields("permissions(displayName, emailAddress, role)")

	if g.DriveBase != "" {
		call = call.
			SupportsAllDrives(true)
	}

	pl, err := call.Do()
	if err != nil {
		return GDriveItem{}, err
	}
	return GDriveItemFromFile(file, pl, server.getUserViews(ctx)), nil
}

func (server *Server) GetFile(ctx context.Context, g *GDrive, id string) (GDriveItem, error) {
	call := g.Service.Files.Get(id).
		Fields("id, name")

	if g.DriveBase != "" {
		call = call.
			SupportsAllDrives(true)
	}

	f, err := call.Do()
	if err != nil {
		return GDriveItem{}, err
	}

	return server.fileToItem(ctx, g, f)
}

func (server *Server) GetFiles(ctx context.Context, g *GDrive, folder string, filter string) ([]GDriveItem, error) {
	call := g.Service.Files.List().
		Q(fmt.Sprintf("mimeType = 'application/vnd.google-apps.document' and '%s' in parents and name contains '%s'", folder, filter)).
		PageSize(100).
		OrderBy("name desc").
		Fields("files(id, name)")

	if g.DriveBase != "" {
		call = call.
			Corpora("drive").
			IncludeItemsFromAllDrives(true).
			DriveId(g.DriveBase).
			SupportsAllDrives(true)
	}

	fl, err := call.Do()
	if err != nil {
		return nil, err
	}

	return SliceToSliceErr(fl.Files, func(file *drive.File) (GDriveItem, error) {
		return server.fileToItem(ctx, g, file)
	})
}
