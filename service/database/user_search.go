package database

import (
	"context"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

func (db *appdbimpl) SearchUsers(ctx context.Context, query string) ([]models.User, error) {
	rows, err := db.c.QueryContext(ctx, `
		SELECT id, username, name, email, avatar_url, gender
		FROM users
		WHERE username LIKE '%' || ? || '%'
		OR name LIKE '%' || ? || '%'
		LIMIT 20
	`, query, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Username, &u.Name, &u.Email, &u.AvatarURL, &u.Gender)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
