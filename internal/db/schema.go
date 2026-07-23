package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
)

// ResetSchema drops all public tables and applies schema.sql
// on the database opened by Init.
func ResetSchema() error {
	if databaseURL == "" {
		return fmt.Errorf("database not initialized")
	}
	if err := dropAllTables(); err != nil {
		return err
	}
	return applySchema(databaseURL)
}

func dropAllTables() error {
	var dropSQL sql.NullString
	err := QueryRow(`
		SELECT COALESCE(
			'DROP TABLE IF EXISTS ' ||
			string_agg('"' || tablename || '"', ', ') ||
			' CASCADE',
			''
		)
		FROM pg_tables
		WHERE schemaname = 'public'
	`).Scan(&dropSQL)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("generate drop statement: %w", err)
	}
	if dropSQL.Valid && dropSQL.String != "" {
		if _, err := Exec(dropSQL.String); err != nil {
			return fmt.Errorf("drop tables: %w", err)
		}
	}
	return nil
}

func applySchema(url string) error {
	schema, err := os.ReadFile(findSchemaPath())
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}

	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		return fmt.Errorf("pgx connect: %w", err)
	}
	defer conn.Close(context.Background())

	if _, err := conn.Exec(context.Background(), string(schema)); err != nil {
		return fmt.Errorf("exec schema: %w", err)
	}
	return nil
}

func findSchemaPath() string {
	candidates := []string{
		"internal/db/schema.sql",
		filepath.Join("..", "db", "schema.sql"),
		filepath.Join("..", "..", "internal", "db", "schema.sql"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "internal/db/schema.sql"
}
