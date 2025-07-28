package database

// UpdateUserName updates a user's username
func (db *appdbimpl) UpdateUserName(userID string, username string) error {
	_, err := db.c.Exec(`UPDATE users SET username = ? WHERE id = ?`, username, userID)
	return err
}
