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
	Folders []GDriveFolder
}

type GDriveFolder struct {
	ID   string
	Name string
}

func (gdf GDriveFolder) HTMLIDSelectBaseFolder(prefix string) string {
	return fmt.Sprintf("%sset-base-folder-%s", prefix, gdf.ID)
}

func (gdf GDriveFolder) URLSelectBaseFolder() string {
	return fmt.Sprintf("/gdrive/set-base-folder/%s", gdf.ID)
}

type GDriveFile struct {
	ID   string
	Name string
}

func (gdf GDriveFile) HTMLIDSelectTemplate(prefix string) string {
	return fmt.Sprintf("%sset-template-%s", prefix, gdf.ID)
}

func (gdf GDriveFile) URLSelectTemplate() string {
	return fmt.Sprintf("/gdrive/set-template/%s", gdf.ID)
}
func (g *GDrive) GetFolders(ctx context.Context) ([]GDriveFolder, error) {
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

	var out []GDriveFolder
	for _, f := range fl.Files {
		out = append(out, GDriveFolder{
			ID:   f.Id,
			Name: f.Name,
		})
	}

	return out, nil
}

func (g *GDrive) GetFiles(ctx context.Context, folder string, filter string) ([]GDriveFile, error) {
	call := g.Service.Files.List().
		Q(fmt.Sprintf("mimeType = 'application/vnd.google-apps.document' and '%s' in parents and name contains '%s'", folder, filter)).
		PageSize(100).
		OrderBy("name desc").
		Fields("files(id, name, parents)")

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

	var out []GDriveFile
	for _, f := range fl.Files {
		out = append(out, GDriveFile{
			ID:   f.Id,
			Name: f.Name,
		})
	}

	return out, nil
}
