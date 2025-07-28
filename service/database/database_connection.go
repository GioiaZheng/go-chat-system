package database

import (
	"database/sql"
	"fmt"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	_ "github.com/mattn/go-sqlite3"
)

type appdbimpl struct {
	c *sql.DB
}

func New(db *sql.DB) (*appdbimpl, error) {
	// Verify the connection is alive
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	// Enable foreign key constraints
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Initialize tables if they don't exist
	if err := initializeTables(db); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return &appdbimpl{c: db}, nil
}

func initializeTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			name TEXT,
			avatar_url TEXT,
			photo TEXT,
			gender TEXT DEFAULT 'unspecified'
		);
		
		CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			content TEXT NOT NULL,
			sender_id TEXT NOT NULL,
			receiver_id TEXT,
			group_id TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(sender_id) REFERENCES users(id),
			FOREIGN KEY(receiver_id) REFERENCES users(id),
			FOREIGN KEY(group_id) REFERENCES groups(id)
		);

		CREATE TABLE IF NOT EXISTS groups (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			avatar_url TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS group_members (
			group_id TEXT,
			user_id TEXT,
			role TEXT,
			PRIMARY KEY (group_id, user_id),
			FOREIGN KEY(group_id) REFERENCES groups(id),
			FOREIGN KEY(user_id) REFERENCES users(id)
		);
	`)
	return err
}

func (db *appdbimpl) Close() error {
	return db.c.Close()
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}

func (db *appdbimpl) GetUserByID(userID string) (models.User, error) {
	var user models.User
	var name, email, avatarUrl, photo, gender sql.NullString

	err := db.c.QueryRow(`
		SELECT id, username, name, email, avatar_url, photo, gender
		FROM users WHERE id = ?
	`, userID).Scan(
		&user.ID,
		&user.Username,
		&name,
		&email,
		&avatarUrl,
		&photo,
		&gender,
	)

	if err != nil {
		return models.User{}, err
	}

	user.Name = name.String
	user.Email = email.String
	user.AvatarUrl = avatarUrl.String
	user.Photo = photo.String
	user.Gender = gender.String

	return user, nil
}

// Check if user exists by username
func (db *appdbimpl) CheckUserExists(username string) (bool, error) {
	var count int
	err := db.c.QueryRow(`
		SELECT COUNT(*) FROM users WHERE username = ?
	`, username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

