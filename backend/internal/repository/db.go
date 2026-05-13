package repository

import (
	"embed"
	"fmt"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

func InitDB(dataDir string) (*sqlx.DB, error) {
	dbPath := filepath.Join(dataDir, "wxnote.db")

	db, err := sqlx.Connect("sqlite", dbPath+"?_pragma=WAL")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return nil, fmt.Errorf("read schema: %w", err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		return nil, fmt.Errorf("run schema: %w", err)
	}

	return db, nil
}
