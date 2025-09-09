package database

import (
	"context"
	"database/sql"
	"log"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// SearchUsers finds users matched by username, name, or email (case-insensitive),
// excluding the requesting user. Escapes wildcard characters to avoid unintended
// broad matches and caps the result size for safety.
func (db *appdbimpl) SearchUsers(ctx context.Context, userID string, query string) ([]models.User, error) {
	const safetyLimit = 50

	var users []models.User
	q := strings.TrimSpace(query)
	if q == "" {
		// Short-circuit empty input; API layer should validate already.
		return users, nil
	}

	// Escape `%` and `_` in LIKE pattern; escape backslash first.
	escapeLike := func(s string) string {
		s = strings.ReplaceAll(s, `\`, `\\`)
		s = strings.ReplaceAll(s, `%`, `\%`)
		s = strings.ReplaceAll(s, `_`, `\_`)
		return s
	}
	escaped := escapeLike(q)
	pattern := "%" + escaped + "%"

	rows, err := db.c.QueryContext(ctx, `
		SELECT id, username, name, email, avatar_url, photo, gender
		  FROM users
		 WHERE (username LIKE ? ESCAPE '\' COLLATE NOCASE
		     OR  name     LIKE ? ESCAPE '\' COLLATE NOCASE
		     OR  email    LIKE ? ESCAPE '\' COLLATE NOCASE)
		   AND id != ?
		 LIMIT ?
	`, pattern, pattern, pattern, userID, safetyLimit)
	if err != nil {
		log.Println("SearchUsers error:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u models.User
		var name, email, avatarURL, photo, gender sql.NullString
		if err := rows.Scan(&u.ID, &u.Username, &name, &email, &avatarURL, &photo, &gender); err != nil {
			log.Println("SearchUsers row scan error:", err)
			return nil, err
		}
		u.Name = name.String
		u.Email = email.String
		u.AvatarUrl = avatarURL.String
		u.Photo = photo.String
		u.Gender = gender.String
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
