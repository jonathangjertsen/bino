package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	if BuildKey == "" {
		panic("missing build key")
	}

	ctx := context.Background()
	fmt.Println("Starting...")

	config, err := loadConfig("config.json")
	if err != nil {
		panic(err)
	}

	cache, err := NewCache(config.DB.CacheFile, func(action, key string, err error) {
		fmt.Fprintf(os.Stderr, "Cache%s(%s): %v\n", action, key, err)
	})
	if err != nil {
		panic(err)
	}

	conn, err := dbSetup(ctx, config.DB)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	queries := New(conn)

	gdriveSA, err := NewGDriveWithServiceAccount(ctx, config.GoogleDrive, queries)
	if err != nil {
		panic(err)
	}
	worker := NewGDriveWorker(config.GoogleDrive, gdriveSA, cache)

	go backgroundDeleteExpiredItems(ctx, queries)

	err = startServer(ctx, conn, queries, cache, worker, config, BuildKey)
	if err != nil {
		panic(err)
	}

	fmt.Println("Ready")
	select {}
}
