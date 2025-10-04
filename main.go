package main

import (
	"context"
	"fmt"

	"github.com/jonathangjertsen/bino/sql"
)

func main() {
	ctx := context.Background()

	conn, err := dbSetup(ctx)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	queries := sql.New(conn)

	err = startServer(ctx, conn, queries)
	if err != nil {
		panic(err)
	}

	fmt.Printf("started...\n")
	select {}
}
