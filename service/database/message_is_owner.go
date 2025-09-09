package database

// IsMessageOwner returns true if the message is authored by the given user.
func (db *appdbimpl) IsMessageOwner(messageID, userID string) (bool, error) {
	var cnt int
	if err := db.c.QueryRow(`
		SELECT COUNT(1)
		  FROM messages
		 WHERE id = ?
		   AND sender_id = ?
	`, messageID, userID).Scan(&cnt); err != nil {
		return false, err
	}
	return cnt > 0, nil
}
