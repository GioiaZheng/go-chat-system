package database

// UpdateUserName updates a user's username
func (db *appdbimpl) UpdateUserName(userID string, name string) error {
	_, err := db.c.Exec(
		`UPDATE users SET username = ? WHERE id = ?`,
		name, userID,
	)
	return err
}
