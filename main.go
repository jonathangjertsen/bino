package main

import (
	"context"
	"fmt"

	"github.com/jonathangjertsen/bino/sql"
)

var BuildKey string

func main() {
	ctx := context.Background()

	conn, err := dbSetup(ctx)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	queries := sql.New(conn)

	if BuildKey == "" {
		panic("missing build key")
	}

	err = startServer(ctx, conn, queries, BuildKey)
	if err != nil {
		panic(err)
	}

	fmt.Printf("started...\n")
	select {}
}
