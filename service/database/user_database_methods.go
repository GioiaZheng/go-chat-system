// user_database_methods.go implements user CRUD and search queries.
// Related files: service/api/users.go, service/models/models.go.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// newID generates a simple unique user identifier; swap with a UUID generator if needed.
func newID() string {
	return fmt.Sprintf("u_%d", time.Now().UnixNano())
}

// CreateUser inserts a new user row, ensuring the display name is unique.
func (db *appdbimpl) CreateUser(u models.User) (models.User, error) {
	u.Name = strings.TrimSpace(u.Name)
	if u.Name == "" {
		return models.User{}, errors.New("empty name")
	}
	if u.ID == "" {
		u.ID = newID()
	}
	_, err := db.c.Exec(`
		INSERT INTO users (id, name, avatar_url)
		VALUES (?, ?, COALESCE(?, NULL))
	`, u.ID, u.Name, strings.TrimSpace(u.AvatarUrl))
	if err != nil {
		return models.User{}, err
	}
	return db.GetUserByID(u.ID)
}

// GetUser resolves a user by ID and mirrors GetUserByID for compatibility.
func (db *appdbimpl) GetUser(userID string) (models.User, error) {
	return db.GetUserByID(userID)
}

// GetUserByID resolves a user by identifier and returns its full record.
func (db *appdbimpl) GetUserByID(userID string) (models.User, error) {
	var u models.User
	var name, avatarUrl sql.NullString
	err := db.c.QueryRow(`
		SELECT id, name, avatar_url
		  FROM users
		 WHERE id = ?
	`, userID).Scan(&u.ID, &name, &avatarUrl)
	if err != nil {
		return models.User{}, err
	}
	u.Name = name.String
	u.AvatarUrl = avatarUrl.String
	return u, nil
}

// GetUserIDFromIdentifier looks up a user ID directly or by case-insensitive name.
func (db *appdbimpl) GetUserIDFromIdentifier(identifier string) (string, error) {
	ident := strings.TrimSpace(identifier)
	if ident == "" {
		return "", errors.New("empty identifier")
	}

	// try id
	var id string
	err := db.c.QueryRow(`SELECT id FROM users WHERE id = ? LIMIT 1`, ident).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	// fallback: by name (case-insensitive)
	err = db.c.QueryRow(`SELECT id FROM users WHERE lower(name) = lower(?) LIMIT 1`, ident).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

// CheckUserExists checks existence by name (case-insensitive).
func (db *appdbimpl) CheckUserExists(name string) (bool, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return false, nil
	}
	var cnt int
	if err := db.c.QueryRow(`
		SELECT COUNT(1) FROM users WHERE lower(name) = lower(?)
	`, name).Scan(&cnt); err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// UpdateUserName updates the user's display name.
func (db *appdbimpl) UpdateUserName(userID string, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("empty name")
	}
	res, err := db.c.Exec(`
		UPDATE users SET name = ? WHERE id = ?
	`, name, userID)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// UpdateUserPhoto replaces only the avatar URL (OpenAPI: avatarUri).
func (db *appdbimpl) UpdateUserPhoto(userID string, publicURL string) error {
	publicURL = strings.TrimSpace(publicURL)
	if publicURL == "" {
		return errors.New("empty avatar url")
	}
	res, err := db.c.Exec(`
		UPDATE users SET avatar_url = ? WHERE id = ?
	`, publicURL, userID)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SearchUsers by name (like), exclude myself, order by name.
func (db *appdbimpl) SearchUsers(ctx context.Context, me string, query string) ([]models.User, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return []models.User{}, nil
	}
	rows, err := db.c.QueryContext(ctx, `
		SELECT id, name, COALESCE(avatar_url, '')
		  FROM users
		 WHERE name LIKE ? ESCAPE '\'
		   AND id <> ?
		 ORDER BY name COLLATE NOCASE ASC
		 LIMIT 50
	`, "%"+q+"%", me)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.AvatarUrl); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
