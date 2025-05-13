package database

// CommentMessage appends a comment to a message
func (db *appdbimpl) CommentMessage(messageID string, userID string, comment string) error {
	_, err := db.c.Exec(`
		UPDATE messages
		SET content = content || '\n[Comment from ' || ? || ']: ' || ?
		WHERE id = ?
	`, userID, comment, messageID)
	return err
}
