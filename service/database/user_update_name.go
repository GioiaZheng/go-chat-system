package database

import (
	"fmt"
	"strings"
)

// UpdateUserName updates a user's username, ensuring non-empty and uniqueness.
func (db *appdbimpl) UpdateUserName(userID string, username string) error {
	// Basic validation
	if strings.TrimSpace(username) == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Attempt update; rely on DB UNIQUE(username) constraint
	_, err := db.c.Exec(`UPDATE users SET username = ? WHERE id = ?`, username, userID)
	if err != nil {
		// Normalize UNIQUE violation into a readable error
		// (SQLite error text contains "UNIQUE constraint failed")
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "unique") && strings.Contains(msg, "username") {
			return fmt.Errorf("username already taken")
		}
		return err
	}
	return nil
}
