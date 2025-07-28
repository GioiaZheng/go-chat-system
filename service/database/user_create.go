package database

import (
	"log"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/gofrs/uuid"
)
func (db *appdbimpl) CreateUser(user models.User, password string) (models.User, error) {
	// 生成 UUID
	userID, err := uuid.NewV4()
	if err != nil {
		log.Println("CreateUser UUID error:", err)
		return models.User{}, err
	}
	user.ID = userID.String()

	// 设置默认字段，防止为 NULL
	if user.Name == "" {
		user.Name = user.Username
	}
	if user.Email == "" {
		user.Email = user.Username + "@example.com"
	}
	if user.AvatarUrl == "" {
		user.AvatarUrl = "https://example.com/default-avatar.png"
	}
	if user.Photo == "" {
		user.Photo = ""
	}
	if user.Gender == "" {
		user.Gender = "unspecified"
	}

	// 插入全部字段
	_, err = db.c.Exec(`
		INSERT INTO users (id, username, email, password, name, avatar_url, photo, gender)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.Username, user.Email, password,
		user.Name, user.AvatarUrl, user.Photo, user.Gender,
	)
	if err != nil {
		log.Println("CreateUser insert error:", err.Error())
		return models.User{}, err
	}

	log.Println("Created user with ID:", user.ID)
	return user, nil
}
