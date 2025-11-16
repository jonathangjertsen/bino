package main

import (
	"context"
	"fmt"

	"github.com/joho/godotenv"
)

func main() {
	if BuildKey == "" {
		panic("missing build key")
	}

	ctx := context.Background()
	fmt.Println("Starting...")

	godotenv.Load(".env")

	config, err := loadConfig("config.json")
	if err != nil {
		panic(err)
	}

	conn, err := dbSetup(ctx)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	queries := New(conn)

	gdriveSA, err := NewGDriveWithServiceAccount(ctx, config.GoogleDrive, queries)
	if err != nil {
		panic(err)
	}
	worker := NewGDriveWorker(ctx, config.GoogleDrive, gdriveSA)

	go backgroundDeleteExpiredItems(ctx, queries)

	err = startServer(ctx, conn, queries, worker, config, BuildKey)
	if err != nil {
		panic(err)
	}

	fmt.Println("Ready")
	select {}
}
