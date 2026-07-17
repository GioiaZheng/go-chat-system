package database

import (
	"context"
	"database/sql"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenSQLiteConfiguresPoolAndEveryConnection(t *testing.T) {
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "pool.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	defer db.Close()

	if got := db.Stats().MaxOpenConnections; got != sqliteMaxOpenConnections {
		t.Fatalf("MaxOpenConnections = %d, want %d", got, sqliteMaxOpenConnections)
	}

	ctx := context.Background()
	connections := make([]*sql.Conn, 0, 2)
	for index := 0; index < 2; index++ {
		conn, err := db.Conn(ctx)
		if err != nil {
			t.Fatalf("acquire connection: %v", err)
		}
		connections = append(connections, conn)
	}
	defer func() {
		for _, conn := range connections {
			_ = conn.Close()
		}
	}()

	for index, conn := range connections {
		var foreignKeys int
		if err := conn.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
			t.Fatalf("connection %d foreign_keys: %v", index, err)
		}
		if foreignKeys != 1 {
			t.Errorf("connection %d foreign_keys = %d, want 1", index, foreignKeys)
		}

		var busyTimeout int
		if err := conn.QueryRowContext(ctx, "PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
			t.Fatalf("connection %d busy_timeout: %v", index, err)
		}
		if busyTimeout != sqliteBusyTimeoutMillis {
			t.Errorf("connection %d busy_timeout = %d, want %d", index, busyTimeout, sqliteBusyTimeoutMillis)
		}
	}
}

func TestOpenSQLiteEnforcesForeignKeys(t *testing.T) {
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "foreign-keys.db"))
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE parents (id INTEGER PRIMARY KEY);
		CREATE TABLE children (
			id INTEGER PRIMARY KEY,
			parent_id INTEGER NOT NULL,
			FOREIGN KEY(parent_id) REFERENCES parents(id)
		);
	`); err != nil {
		t.Fatalf("create foreign-key fixture: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO children (id, parent_id) VALUES (1, 999)`); err == nil {
		t.Fatal("insert with missing parent succeeded; foreign keys are not enforced")
	}
}

func TestBuildSQLiteDSNOverridesConnectionOptions(t *testing.T) {
	dsn, err := buildSQLiteDSN("test.db?_foreign_keys=off&_busy_timeout=1")
	if err != nil {
		t.Fatalf("buildSQLiteDSN: %v", err)
	}

	queryStart := strings.IndexRune(dsn, '?')
	if queryStart < 0 {
		t.Fatalf("DSN %q has no query options", dsn)
	}
	query, err := url.ParseQuery(dsn[queryStart+1:])
	if err != nil {
		t.Fatalf("parse DSN query: %v", err)
	}

	if got := query.Get("_foreign_keys"); got != "on" {
		t.Errorf("_foreign_keys = %q, want on", got)
	}
	if got := query.Get("_busy_timeout"); got != "5000" {
		t.Errorf("_busy_timeout = %q, want 5000", got)
	}
	if got := query.Get("_journal_mode"); got != "WAL" {
		t.Errorf("_journal_mode = %q, want WAL", got)
	}
}
