package database

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// NOTE (English):
// This file consolidates "user" DB methods that are NOT already implemented
// in other files (e.g., database_connection.go). We intentionally AVOID
// redefining CheckUserExists and GetUserByID here if they exist elsewhere.
// We DO provide GetUser as a thin alias to GetUserByID, because the interface
// requires GetUser and some projects don't implement it elsewhere.

// CreateUser inserts a new user record and its credentials.
// The incoming user may not have an ID; we generate one when missing.
func (db *appdbimpl) CreateUser(user models.User, password string) (models.User, error) {
	user.Username = strings.TrimSpace(user.Username)
	user.Email = strings.TrimSpace(user.Email)
	user.Name = strings.TrimSpace(user.Name)
	user.Gender = strings.TrimSpace(user.Gender)

	// Generate ID if empty (TEXT PK).
	if strings.TrimSpace(user.ID) == "" {
		if err := db.c.QueryRow(`SELECT lower(hex(randomblob(16)))`).Scan(&user.ID); err != nil {
			return models.User{}, err
		}
	}

	// Insert into users.
	_, err := db.c.Exec(`
		INSERT INTO users (id, username, name, email, avatar_url, gender)
		VALUES (?, ?, ?, ?, ?, ?)
	`, user.ID, user.Username, user.Name, user.Email, user.AvatarUrl, user.Gender)
	if err != nil {
		return models.User{}, err
	}

	// Insert credentials if the table exists; ignore if schema doesn't have it.
	// We keep both "user_id + username" mapping to cover different schemas.
	if _, err := db.c.Exec(`
		INSERT OR REPLACE INTO user_credentials (user_id, username, password)
		VALUES (?, ?, ?)
	`, user.ID, user.Username, password); err != nil {
		// Fallback: legacy inline password column on users (if present).
		_, _ = db.c.Exec(`UPDATE users SET password = ? WHERE id = ?`, password, user.ID)
	}

	return user, nil
}

// AuthenticateUser checks credentials by email + password and returns the user.
// We accept both credential storage models (user_credentials or users.password).
func (db *appdbimpl) AuthenticateUser(email, password string) (models.User, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	// Resolve user by email.
	var u models.User
	if err := db.c.QueryRow(`
		SELECT id, username, name, email, avatar_url, gender
		FROM users WHERE email = ?
	`, email).Scan(&u.ID, &u.Username, &u.Name, &u.Email, &u.AvatarUrl, &u.Gender); err != nil {
		return models.User{}, err
	}

	// Check user_credentials first.
	var cnt int
	if err := db.c.QueryRow(`
		SELECT COUNT(1) FROM user_credentials
		WHERE user_id = ? AND password = ?
	`, u.ID, password).Scan(&cnt); err == nil && cnt > 0 {
		return u, nil
	}

	// Fallback: inline password column on users (legacy).
	if err := db.c.QueryRow(`
		SELECT COUNT(1) FROM users WHERE id = ? AND password = ?
	`, u.ID, password).Scan(&cnt); err == nil && cnt > 0 {
		return u, nil
	}

	return models.User{}, errors.New("invalid credentials")
}

// GetUserIDFromIdentifier resolves a username or an email to a user ID.
func (db *appdbimpl) GetUserIDFromIdentifier(identifier string) (string, error) {
	identifier = strings.TrimSpace(identifier)
	var id string
	err := db.c.QueryRow(`
		SELECT id FROM users
		WHERE lower(username) = lower(?)
		   OR lower(email)    = lower(?)
		LIMIT 1
	`, identifier, identifier).Scan(&id)
	return id, err
}

// GetUser is a thin alias to GetUserByID (required by AppDatabase).
// We keep it here so the concrete type *appdbimpl* satisfies the interface,
// even if another file doesn't provide GetUser.
func (db *appdbimpl) GetUser(userID string) (models.User, error) {
	return db.GetUserByID(userID)
}

// GetUserByCredentials fetches a user by (username,password).
// We check user_credentials first, then fall back to users.password (legacy).
func (db *appdbimpl) GetUserByCredentials(name, password string) (models.User, error) {
	name = strings.TrimSpace(name)
	password = strings.TrimSpace(password)

	// Try user_credentials (username + password).
	var u models.User
	err := db.c.QueryRow(`
		SELECT u.id, u.username, u.name, u.email, u.avatar_url, u.gender
		  FROM users u
		  JOIN user_credentials uc ON uc.user_id = u.id
		 WHERE lower(uc.username) = lower(?) AND uc.password = ?
		 LIMIT 1
	`, name, password).Scan(&u.ID, &u.Username, &u.Name, &u.Email, &u.AvatarUrl, &u.Gender)
	if err == nil {
		return u, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return models.User{}, err
	}

	// Fallback: users table has password column (legacy).
	err = db.c.QueryRow(`
		SELECT id, username, name, email, avatar_url, gender
		  FROM users
		 WHERE lower(username) = lower(?) AND password = ?
		 LIMIT 1
	`, name, password).Scan(&u.ID, &u.Username, &u.Name, &u.Email, &u.AvatarUrl, &u.Gender)
	if err != nil {
		return models.User{}, err
	}
	return u, nil
}

// UpdateUserName updates the user's username (handle).
// If you intended to change display name, adjust SQL accordingly.
func (db *appdbimpl) UpdateUserName(userID, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("empty username")
	}

	// Ensure uniqueness.
	var cnt int
	if err := db.c.QueryRow(`
		SELECT COUNT(1) FROM users
		WHERE lower(username) = lower(?) AND id <> ?
	`, name, userID).Scan(&cnt); err != nil {
		return err
	}
	if cnt > 0 {
		return errors.New("username already taken")
	}

	// Update users.username.
	if _, err := db.c.Exec(`UPDATE users SET username = ? WHERE id = ?`, name, userID); err != nil {
		return err
	}

	// Keep user_credentials in sync if table exists.
	_, _ = db.c.Exec(`UPDATE user_credentials SET username = ? WHERE user_id = ?`, name, userID)
	return nil
}

// UpdateUserPhoto updates the avatar_url for a user.
func (db *appdbimpl) UpdateUserPhoto(userID, photoPath string) error {
	photoPath = strings.TrimSpace(photoPath)
	_, err := db.c.Exec(`UPDATE users SET avatar_url = ? WHERE id = ?`, photoPath, userID)
	return err
}

// SearchUsers returns users matching the query (by username/name/email), excluding self.
func (db *appdbimpl) SearchUsers(ctx context.Context, userID string, query string) ([]models.User, error) {
	q := strings.TrimSpace(query)
	args := []interface{}{userID}
	where := `WHERE id <> ?`

	if q != "" {
		qLike := "%" + strings.ToLower(q) + "%"
		where += ` AND (lower(username) LIKE ? OR lower(name) LIKE ? OR lower(email) LIKE ?)`
		args = append(args, qLike, qLike, qLike)
	}

	rows, err := db.c.QueryContext(ctx, `
		SELECT id, username, name, email, avatar_url, gender
		  FROM users
		`+where+`
		  ORDER BY username COLLATE NOCASE ASC
		  LIMIT 50
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Name, &u.Email, &u.AvatarUrl, &u.Gender); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}
