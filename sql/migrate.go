package sql

import "embed"

//go:embed migrations/*
var DBMigrations embed.FS
