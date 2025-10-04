package main

import (
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()

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

	err = startServer(ctx, conn, queries, config, BuildKey)
	if err != nil {
		panic(err)
	}

	fmt.Printf("started...\n")
	select {}
}
