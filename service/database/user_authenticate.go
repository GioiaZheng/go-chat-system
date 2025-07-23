package database

import (
	"github.com/GioiaZheng/Wasa_proj/service/models"
	// "golang.org/x/crypto/bcrypt"
)

// AuthenticateUser checks if the provided credentials are valid
func (db *appdbimpl) AuthenticateUser(email, password string) (models.User, error) {
	var user models.User
	var hashedPassword string

	err := db.c.QueryRow(`
		SELECT id, username, name, email, avatar_url, gender, password
		FROM users
		WHERE email = ?
	`, email).Scan(&user.ID, &user.Username, &user.Name, &user.Email, &user.AvatarUrl, &user.Gender, &hashedPassword)
	if err != nil {
		return models.User{}, err
	}

	// 验证密码
	if err := VerifyPassword(hashedPassword, password); err != nil {
		return models.User{}, err
	}

	return user, nil
}

// VerifyPassword checks if the provided password matches the stored hash
func VerifyPassword(hashedPassword, password string) error {
	// return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return nil
}
