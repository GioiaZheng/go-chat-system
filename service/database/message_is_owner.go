package database

import (
	"database/sql"
	"fmt"
)

// IsMessageOwner checks if the user is the owner of the message
func (db *appdbimpl) IsMessageOwner(userID string, messageID string) (bool, error) {
	var senderID string
	err := db.c.QueryRow(`
		SELECT sender_id
		FROM messages
		WHERE id = ?
	`, messageID).Scan(&senderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("error checking message ownership: %w", err)
	}
	return senderID == userID, nil
}
