package database

// UncommentMessage removes the comment from a message
func (db *appdbimpl) UncommentMessage(messageID string) error {
	_, err := db.c.Exec(`
		UPDATE messages
		SET comment = NULL
		WHERE id = ?
	`, messageID)
	return err
}
