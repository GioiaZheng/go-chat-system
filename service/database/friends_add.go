package database

// AddFriend adds a friendship between two users
func (db *appdbimpl) AddFriend(userID string, friendID string) error {
	tx, err := db.c.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if friend exists
	var exists bool
	err = tx.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)
	`, friendID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return err
	}

	// Insert friendship (ignore if already exists)
	_, err = tx.Exec(`
		INSERT OR IGNORE INTO friends (user_id, friend_id)
		VALUES (?, ?)
	`, userID, friendID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
