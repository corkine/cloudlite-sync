package database

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func New(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := initTables(db); err != nil {
		return nil, fmt.Errorf("failed to init tables: %w", err)
	}

	return &DB{db}, nil
}

func initTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			website TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS credentials (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			token TEXT UNIQUE NOT NULL,
			is_active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (project_id) REFERENCES projects(id)
		)`,
		`CREATE TABLE IF NOT EXISTS database_versions (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			version TEXT NOT NULL,
			file_hash TEXT NOT NULL,
			file_name TEXT NOT NULL,
			file_size INTEGER NOT NULL,
			oss_key TEXT NOT NULL,
			description TEXT,
			is_latest BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (project_id) REFERENCES projects(id)
		)`,
		`CREATE TABLE IF NOT EXISTS jwt_projects (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			public_key TEXT NOT NULL,
			private_key TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS jwt_tokens (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			purpose TEXT NOT NULL,
			username TEXT NOT NULL,
			role TEXT NOT NULL,
			token TEXT UNIQUE NOT NULL,
			is_active BOOLEAN DEFAULT 1,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (project_id) REFERENCES jwt_projects(id)
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

// GenerateProjectID 生成唯一的项目ID
func GenerateProjectID() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	id := make([]byte, 8)
	for i := range id {
		id[i] = letters[r.Intn(len(letters))]
	}
	return string(id)
}

// GenerateToken 生成唯一的凭证token
func GenerateToken() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	token := make([]byte, 32)
	for i := range token {
		token[i] = letters[r.Intn(len(letters))]
	}
	return string(token)
}
