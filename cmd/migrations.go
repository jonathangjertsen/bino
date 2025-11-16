package main

import "embed"

//go:embed migrations/*
var DBMigrations embed.FS
