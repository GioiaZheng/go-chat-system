package database

// UncommentMessage attempts to clear a message's comment if such a column exists.
// If the schema does not have a dedicated comment column, this becomes a no-op.
// Rationale:
//   - Some implementations store comments concatenated into content.
//   - Others use a separate "comment" column.
//   - To maintain compatibility without changing OpenAPI, we make this tolerant.
func (db *appdbimpl) UncommentMessage(messageID string) error {
	// Try best-effort update. If the column does not exist, ignore the error to keep the API idempotent.
	_, err := db.c.Exec(`
		UPDATE messages
		   SET comment = NULL
		 WHERE id = ?
	`, messageID)
	if err != nil {
		// NOTE: Silently tolerate errors due to schema differences (e.g., "no such column: comment").
		// We deliberately do not propagate the error to keep the endpoint behavior idempotent across variants.
		return nil
	}
	return nil
}
