package database

import (
	"fmt"
	"log"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser creates a new user with a hashed password
func (db *appdbimpl) CreateUser(user models.User) (models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}

	res, err := db.c.Exec(`
		INSERT INTO users (username, name, email, password, gender)
		VALUES (?, ?, ?, ?, ?)`,
		user.Username, user.Name, user.Email, hashedPassword, user.Gender,
	)
	if err != nil {
		log.Println("CreateUser insert error:", err)
		return models.User{}, err
	}

	id, _ := res.LastInsertId()
	user.ID = fmt.Sprintf("%d", id)
	return user, nil
}
