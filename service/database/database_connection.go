package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type appdbimpl struct {
	c *sql.DB
}

// New creates a new database handler from an existing connection
func New(db *sql.DB) (AppDatabase, error) {
	// Verify the connection is alive
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &appdbimpl{c: db}, nil
}

func (db *appdbimpl) Close() error {
	return db.c.Close()
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}

var _ AppDatabase = (*appdbimpl)(nil)
