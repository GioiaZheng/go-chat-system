package database

import (
	"fmt"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"golang.org/x/crypto/bcrypt"
)

// AuthenticateUser authenticates a user by email and password
func (db *appdbimpl) AuthenticateUser(email string, password string) (models.User, string, error) {
	var user models.User
	var hashedPassword string

	err := db.c.QueryRow(`
		SELECT id, username, email, password FROM users WHERE email = ?
	`, email).Scan(&user.ID, &user.Username, &user.Email, &hashedPassword)
	if err != nil {
		return models.User{}, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return models.User{}, "", fmt.Errorf("invalid credentials")
	}

	token := fmt.Sprintf("token-for-%s", user.ID)
	return user, token, nil
}
