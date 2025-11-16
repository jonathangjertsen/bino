package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
)

func dbSetup(ctx context.Context) (*pgxpool.Pool, error) {
	pass := os.Getenv("BINO_DB_PASSWORD")
	if pass == "" {
		panic("missing env variable: BINO_DB_PASSWORD")
	}

	host := os.Getenv("BINO_DB_HOST")
	if host == "" {
		panic("missing env variable: BINO_DB_HOST")
	}

	port := os.Getenv("BINO_DB_PORT")
	if port == "" {
		panic("missing env variable: BINO_DB_PORT")
	}

	connStr := fmt.Sprintf("postgres://bino:%s@%s:%s/bino?sslmode=disable", pass, host, port)
	fmt.Printf("conn=" + connStr + "\n")

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
