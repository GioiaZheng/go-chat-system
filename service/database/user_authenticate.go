package database

import (
	"database/sql"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// AuthenticateUser authenticates a user by email and password.
// It fetches the user by email, then verifies the provided password.
// Returns sql.ErrNoRows if the user does not exist, or false if password is invalid.
func (db *appdbimpl) AuthenticateUser(email, password string) (models.User, error) {
	var u models.User

	err := db.c.QueryRow(`
		SELECT id, username, name, email, password, avatar_url, gender
		FROM users
		WHERE email = ?
	`, email).Scan(&u.ID, &u.Username, &u.Name, &u.Email, &u.Password, &u.AvatarUrl, &u.Gender)
	if err != nil {
		// Propagate not found / driver errors as-is.
		return models.User{}, err
	}

	if !VerifyPassword(password, u.Password) {
		// Wrong credentials: return empty user and a standardized error.
		// Callers can translate this into 401 Unauthorized.
		return models.User{}, sql.ErrNoRows
	}

	// Do not leak the stored password beyond DB layer (already tagged json:"-").
	return u, nil
}
