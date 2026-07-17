package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const (
	sqliteMaxOpenConnections = 8
	sqliteMaxIdleConnections = 4
	sqliteBusyTimeoutMillis  = 5000
)

// OpenSQLite creates the shared SQLite connection pool used by the service.
// Driver options are encoded in the DSN so they are applied to every physical
// connection opened by database/sql, not only to the first connection.
func OpenSQLite(filename string) (*sql.DB, error) {
	dsn, err := buildSQLiteDSN(filename)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open SQLite database: %w", err)
	}

	db.SetMaxOpenConns(sqliteMaxOpenConnections)
	db.SetMaxIdleConns(sqliteMaxIdleConnections)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connect to SQLite database: %w", err)
	}

	return db, nil
}

func buildSQLiteDSN(filename string) (string, error) {
	if strings.TrimSpace(filename) == "" {
		return "", fmt.Errorf("SQLite filename is empty")
	}

	base := filename
	rawQuery := ""
	if queryStart := strings.IndexRune(filename, '?'); queryStart >= 0 {
		base = filename[:queryStart]
		rawQuery = filename[queryStart+1:]
	}

	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return "", fmt.Errorf("parse SQLite options: %w", err)
	}

	query.Set("_busy_timeout", fmt.Sprintf("%d", sqliteBusyTimeoutMillis))
	query.Set("_foreign_keys", "on")
	query.Set("_journal_mode", "WAL")

	return base + "?" + query.Encode(), nil
}
