package database

import (
	"context"
	"database/sql"
	"log"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"golang.org/x/crypto/bcrypt"
)

// SearchUsers searches for users based on a query and user ID
func (db *appdbimpl) SearchUsers(ctx context.Context, userID string, query string) ([]models.User, error) {
	var users []models.User

	// 排除自己，防止自己出现在搜索结果中
	rows, err := db.c.QueryContext(ctx, `
		SELECT id, username, name, email, avatar_url, gender
		FROM users
		WHERE (username LIKE ? OR name LIKE ? OR email LIKE ?)
		AND id != ?
	`, "%"+query+"%", "%"+query+"%", "%"+query+"%", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Name, &user.Email, &user.AvatarUrl, &user.Gender); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// CheckUserExists 检查用户是否存在
func (db *appdbimpl) CheckUserExists(username string) (bool, error) {
	var exists bool
	err := db.c.QueryRow(`
		SELECT EXISTS (SELECT 1 FROM users WHERE username = ?)
	`, username).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Println("CheckUserExists error:", err)
		return false, err
	}
	return exists, nil
}

// GetUserByCredentials 根据用户名和密码获取用户 ID
func (db *appdbimpl) GetUserByCredentials(username, password string) (string, error) {
	var storedPassword string
	var userID string

	err := db.c.QueryRow(`
		SELECT id, password FROM users WHERE username = ?
	`, username).Scan(&userID, &storedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", sql.ErrNoRows
		}
		log.Println("GetUserByCredentials error:", err)
		return "", err
	}

	// 校验密码
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		log.Println("GetUserByCredentials password mismatch:", err)
		return "", sql.ErrNoRows
	}

	return userID, nil
}
