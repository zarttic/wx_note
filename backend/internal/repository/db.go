package repository

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"

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

	// 数据库迁移：为已有表添加新列（列已存在则忽略错误）
	migrations := []string{
		"ALTER TABLE user_configs ADD COLUMN last_author TEXT NOT NULL DEFAULT ''",
	}
	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			// SQLite ALTER TABLE ADD COLUMN 已存在时会报错，忽略即可
			if !strings.Contains(err.Error(), "duplicate column name") {
				return nil, fmt.Errorf("run migration %q: %w", m, err)
			}
		}
	}

	return db, nil
}
