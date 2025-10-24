package main

import (
	"context"
	"fmt"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type GDriveWorkerConfig struct {
	ServiceAccountKeyLocation string
	DriveBase                 string
	JournalFolder             string
	TemplateFile              string
}

func NewGDriveWithServiceAccount(ctx context.Context, config GDriveWorkerConfig, queries *Queries) (*GDrive, error) {
	srv, err := drive.NewService(ctx, option.WithCredentialsFile(config.ServiceAccountKeyLocation))
	if err != nil {
		return nil, fmt.Errorf("creating Drive service: %w", err)
	}

	return &GDrive{
		Service:   srv,
		Queries:   queries,
		DriveBase: config.DriveBase,
	}, nil
}
