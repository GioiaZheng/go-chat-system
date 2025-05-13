package database

// UpdateUserPhoto updates a user's profile photo
func (db *appdbimpl) UpdateUserPhoto(userID string, photoPath string) error {
	_, err := db.c.Exec(
		`UPDATE users SET photo = ? WHERE id = ?`,
		photoPath, userID,
	)
	return err
}
