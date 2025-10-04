package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	"golang.org/x/term"
)

func dbSetup(ctx context.Context) (*pgxpool.Pool, error) {
	url := os.Getenv("BINO_DB_URL")
	if url == "" {
		url = "localhost"
	}

	pass := os.Getenv("BINO_DB_PASSWORD")
	if pass == "" {
		fmt.Print("Password for user bino: ")
		b, _ := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		pass = string(b)
	}

	connStr := fmt.Sprintf("postgres://bino:%s@%s:5432/bino", pass, url)

	conn, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, err
	}

	migrations := migrate.EmbedFileSystemMigrationSource{
		FileSystem: DBMigrations,
		Root:       "migrations",
	}

	sqlDB := stdlib.OpenDBFromPool(conn)
	defer sqlDB.Close()

	n, err := migrate.ExecContext(ctx, sqlDB, "postgres", migrations, migrate.Up)
	if err != nil {
		return nil, fmt.Errorf("migrating: %w", err)
	}
	fmt.Printf("Did %d migrations\n", n)

	return conn, nil
}
