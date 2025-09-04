package database

import (
	"context"
	"database/sql"
	"log"
	"fmt"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// SearchUsers searches for users based on a query and user ID
func (db *appdbimpl) SearchUsers(ctx context.Context, userID string, query string) ([]models.User, error) {
	var users []models.User

	rows, err := db.c.QueryContext(ctx, `
		SELECT id, username, name, email, avatar_url, photo, gender
		FROM users
		WHERE (username LIKE ? OR name LIKE ? OR email LIKE ?)
		AND id != ?
	`, "%"+query+"%", "%"+query+"%", "%"+query+"%", userID)
	if err != nil {
		log.Println("SearchUsers error:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		var name, email, avatarUrl, photo, gender sql.NullString

		err := rows.Scan(&user.ID, &user.Username, &name, &email, &avatarUrl, &photo, &gender)
		if err != nil {
			log.Println("SearchUsers row scan error:", err)
			return nil, err
		}

		user.Name = name.String
		user.Email = email.String
		user.AvatarUrl = avatarUrl.String
		user.Photo = photo.String
		user.Gender = gender.String

		users = append(users, user)
	}

	// ✅ 必须检查 rows.Err()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (db *appdbimpl) GetUserByCredentials(username string, password string) (models.User, error) {
	var user models.User
	var storedPassword string
	var name, email, avatarUrl, photo, gender sql.NullString

	err := db.c.QueryRow(`
		SELECT id, username, name, email, password, avatar_url, photo, gender
		FROM users WHERE username = ?
	`, username).Scan(
		&user.ID,
		&user.Username,
		&name,
		&email,
		&storedPassword,
		&avatarUrl,
		&photo,
		&gender,
	)

	if err != nil {
		log.Println("GetUserByCredentials query error:", err)
		return models.User{}, err
	}

	if password != storedPassword {
		log.Println("GetUserByCredentials password mismatch")
		return models.User{}, fmt.Errorf("invalid credentials")
	}

	user.Name = name.String
	user.Email = email.String
	user.AvatarUrl = avatarUrl.String
	user.Photo = photo.String
	user.Gender = gender.String

	return user, nil
}
