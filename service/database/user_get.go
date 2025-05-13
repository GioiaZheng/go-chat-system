package database

import (
	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// GetUserByID retrieves a user by their ID
func (db *appdbimpl) GetUserByID(userID string) (models.User, error) {
	var user models.User
	err := db.c.QueryRow(`
		SELECT id, username, name, email, photo, avatar_url, gender
		FROM users
		WHERE id = ?
	`, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Email,
		&user.Photo,
		&user.AvatarURL,
		&user.Gender,
	)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}
