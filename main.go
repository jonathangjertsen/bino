package main

import (
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()
	fmt.Println("Starting...")

	config, err := loadConfig("config.json")
	if err != nil {
		panic(err)
	}

	if BuildKey == "" {
		panic("missing build key")
	}

	conn, err := dbSetup(ctx)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	queries := New(conn)

	go backgroundDeleteExpiredSessions(ctx, queries)

	err = startServer(ctx, conn, queries, config, BuildKey)
	if err != nil {
		panic(err)
	}

	fmt.Println("Ready")
	select {}
}
