package database

// UncommentMessage removes the most recent comment on a message (regardless of author).
// This matches the AppDatabase interface: UncommentMessage(messageID string) error
// and the API endpoint /messages/{id}/uncomment which has no request body.
func (db *appdbimpl) UncommentMessage(messageID string) error {
	_, err := db.c.Exec(`
		DELETE FROM message_comments
		      WHERE id IN (
					SELECT id
					  FROM message_comments
					 WHERE message_id = ?
					 ORDER BY datetime(created_at) DESC, id DESC
					 LIMIT 1
			  )
	`, messageID)
	return err
}
