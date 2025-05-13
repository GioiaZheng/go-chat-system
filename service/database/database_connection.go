package database

import (
	"database/sql"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	_ "github.com/mattn/go-sqlite3"
)

type appdbimpl struct {
	c *sql.DB
}

// New creates a new database handler from an existing connection
func New(db *sql.DB) (AppDatabase, error) {
	// Verify the connection is alive
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &appdbimpl{c: db}, nil
}

func (db *appdbimpl) Close() error {
	return db.c.Close()
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}

func (db *appdbimpl) GetUserByID(userID string) (models.User, error) {
	var user models.User
	err := db.c.QueryRow(`
		SELECT id, username, name, email, avatar_url, photo, gender
		FROM users WHERE id = ?
	`, userID).Scan(&user.ID, &user.Username, &user.Name, &user.Email, &user.AvatarUrl, &user.Photo, &user.Gender)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

var _ AppDatabase = (*appdbimpl)(nil)
