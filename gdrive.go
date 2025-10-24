package main

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

var GoogleDriveScopes = []string{
	"https://www.googleapis.com/auth/drive",
}

type GDrive struct {
	Service   *drive.Service
	Queries   *Queries
	DriveBase string
}

func (server *Server) getDriveService(r *http.Request) (*GDrive, error) {
	token, err := server.getTokenFromRequest(r)
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
		DriveBase: server.Config.GoogleDrive.DriveBase,
	}, nil
}

type GDriveItem struct {
	ID          string
	Name        string
	Valid       bool
	Permissions []GDrivePermission

	file *drive.File
}

func (gdi GDriveItemView) LoggedInUserCanShare(ctx context.Context) bool {
	cd := MustLoadCommonData(ctx)
	return (gdi.Item.Valid &&
		gdi.RequestingUser == cd.User.AppuserID &&
		gdi.RequestingUserCapabilities.CanShare)
}

func GDriveItemFromFile(f *drive.File, p *drive.PermissionList) GDriveItem {
	if f == nil {
		return GDriveItem{}
	}
	return GDriveItem{
		ID:    f.Id,
		Name:  f.Name,
		Valid: true,
		Permissions: SliceToSlice(p.Permissions, func(p *drive.Permission) GDrivePermission {
			return GDrivePermission{
				DisplayName: p.DisplayName,
				Email:       p.EmailAddress,
				Role:        p.Role,
			}
		}),
		file: f,
	}

}

type GDrivePermission struct {
	DisplayName string
	Email       string
	Role        string
}

func (g *GDrive) fileToItem(file *drive.File) (GDriveItem, error) {
	call := g.Service.Permissions.List(file.Id).Fields("permissions(displayName, emailAddress, role)")

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
	call := g.Service.Files.Get(id).
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
