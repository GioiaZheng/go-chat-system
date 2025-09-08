package database

import (
	"context"
	"database/sql"
	"fmt" // FIX: needed for fmt.Errorf
	"log"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// SearchUsers finds users whose username, name, or email matches the query,
// excluding the requesting user (userID).
// Notes:
// - Uses a simple LIKE %query% match across (username, name, email).
// - Trims spaces in the incoming query to avoid pointless full-table scans on whitespace.
// - After iterating rows.Next(), we MUST check rows.Err() to catch driver-side errors.
func (db *appdbimpl) SearchUsers(ctx context.Context, userID string, query string) ([]models.User, error) {
	var users []models.User

	q := strings.TrimSpace(query)
	rows, err := db.c.QueryContext(ctx, `
		SELECT id, username, name, email, avatar_url, photo, gender
		FROM users
		WHERE (username LIKE ? OR name LIKE ? OR email LIKE ?)
		  AND id != ?
	`, "%"+q+"%", "%"+q+"%", "%"+q+"%", userID)
	if err != nil {
		log.Println("SearchUsers error:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		var name, email, avatarURL, photo, gender sql.NullString

		if err := rows.Scan(&user.ID, &user.Username, &name, &email, &avatarURL, &photo, &gender); err != nil {
			log.Println("SearchUsers row scan error:", err)
			return nil, err
		}

		user.Name = name.String
		user.Email = email.String
		user.AvatarUrl = avatarURL.String
		user.Photo = photo.String
		user.Gender = gender.String

		users = append(users, user)
	}

	// FIX: must check rows.Err() after iteration to catch driver-side errors.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// GetUserByCredentials retrieves a user by username and verifies the password.
// Security notes:
//   - Uses VerifyPassword (constant-time compare) to avoid obvious timing side-channels.
//   - Returns a generic "invalid credentials" error on mismatch (do not leak which field failed).
//   - The stored password field must be a plain text or pre-hashed value consistent with VerifyPassword.
//     For the assignment baseline we use plain text; replace VerifyPassword with a bcrypt/argon2 check in production.
func (db *appdbimpl) GetUserByCredentials(username string, password string) (models.User, error) {
	var user models.User
	var storedPassword string
	var name, email, avatarURL, photo, gender sql.NullString

	err := db.c.QueryRow(`
		SELECT id, username, name, email, password, avatar_url, photo, gender
		FROM users
		WHERE username = ?
	`, username).Scan(
		&user.ID,
		&user.Username,
		&name,
		&email,
		&storedPassword,
		&avatarURL,
		&photo,
		&gender,
	)
	if err != nil {
		// Propagate not-found or driver error as-is.
		log.Println("GetUserByCredentials query error:", err)
		return models.User{}, err
	}

	// Use constant-time comparison to avoid trivial timing differences.
	if !VerifyPassword(password, storedPassword) {
		// Keep a generic error to avoid leaking which check failed.
		log.Println("GetUserByCredentials password mismatch")
		return models.User{}, fmt.Errorf("invalid credentials")
	}

	user.Name = name.String
	user.Email = email.String
	user.AvatarUrl = avatarURL.String
	user.Photo = photo.String
	user.Gender = gender.String

	// NOTE: user.Password has json:"-" tag in models, so it won't leak in API responses.
	return user, nil
}
