package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// GetUserByCredentials finds a user by username and verifies the password.
// NOTE: For the assignment's simplified login, you may pass an empty password ("").
func (db *appdbimpl) GetUserByCredentials(username string, password string) (models.User, error) {
	u := models.User{}
	var storedPassword string
	var name, email, avatarURL, photo, gender sql.NullString

	err := db.c.QueryRow(`
		SELECT id, username, name, email, password, avatar_url, photo, gender
		  FROM users
		 WHERE username = ?
	`, username).Scan(
		&u.ID,
		&u.Username,
		&name,
		&email,
		&storedPassword,
		&avatarURL,
		&photo,
		&gender,
	)
	if err != nil {
		log.Println("GetUserByCredentials query error:", err)
		return models.User{}, err
	}

	// Constant-time compare (see password_verify.go).
	if !VerifyPassword(password, storedPassword) {
		return models.User{}, fmt.Errorf("invalid credentials")
	}

	u.Name = name.String
	u.Email = email.String
	u.AvatarUrl = avatarURL.String
	u.Photo = photo.String
	u.Gender = gender.String
	return u, nil
}
