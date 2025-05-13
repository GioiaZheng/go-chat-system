package database

import (
	"log"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser creates a new user with a hashed password
func (db *appdbimpl) CreateUser(user models.User, password string) (models.User, error) {
	// 生成 UUID 作为用户 ID
	userID, err := uuid.NewV4()
	if err != nil {
		log.Println("CreateUser UUID error:", err)
		return models.User{}, err
	}
	user.ID = userID.String()

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("CreateUser password hash error:", err)
		return models.User{}, err
	}

	// 插入用户数据
	_, err = db.c.Exec(`
		INSERT INTO users (id, username, name, email, avatar_url, gender, password)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.Username, user.Name, user.Email, user.AvatarUrl, user.Gender, hashedPassword,
	)
	if err != nil {
		log.Println("CreateUser insert error:", err)
		return models.User{}, err
	}

	return user, nil
}
