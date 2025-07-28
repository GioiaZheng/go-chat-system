package database

import (
	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// GetUser retrieves a user by their ID
func (db *appdbimpl) GetUser(userID string) (models.User, error) {
	var user models.User
	err := db.c.QueryRow(`
		SELECT id, username, name, email, avatar_url, gender, photo
		FROM users WHERE id = ?
	`, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Email,
		&user.AvatarUrl,
		&user.Gender,
		&user.Photo,
	)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

// GetUserIDFromIdentifier retrieves the user ID from a given identifier (token)
func (db *appdbimpl) GetUserIDFromIdentifier(identifier string) (string, error) {
	var userID string
	err := db.c.QueryRow(`
		SELECT id FROM users WHERE token = ?
	`, identifier).Scan(&userID)
	if err != nil {
		return "", err
	}
	return userID, nil
}
