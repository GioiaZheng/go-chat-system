package database

import "github.com/google/uuid"

// CommentMessage inserts a comment for the message id.
// Assumes message existence was already validated by handler.
func (db *appdbimpl) CommentMessage(messageID, authorID, content string) error {
	id := uuid.NewString()
	_, err := db.c.Exec(`
		INSERT INTO message_comments (id, message_id, author_id, content, created_at)
		VALUES (?, ?, ?, ?, datetime('now'))
	`, id, messageID, authorID, content)
	return err
}
