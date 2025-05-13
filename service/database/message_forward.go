package database

// ForwardMessage forwards a message to another user or group
func (db *appdbimpl) ForwardMessage(userID string, messageID string, toUserID string, toGroupID string) error {
	var content string
	err := db.c.QueryRow(`
		SELECT content
		FROM messages
		WHERE id = ?
	`, messageID).Scan(&content)
	if err != nil {
		return err
	}

	if toUserID != "" {
		_, err = db.c.Exec(`
			INSERT INTO messages (sender_id, receiver_id, content, created_at)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		`, userID, toUserID, content)
	} else if toGroupID != "" {
		_, err = db.c.Exec(`
			INSERT INTO messages (sender_id, group_id, content, created_at)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		`, userID, toGroupID, content)
	}
	return err
}
