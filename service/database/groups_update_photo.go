package database

// UpdateGroupPhoto updates the photo path of a group
func (db *appdbimpl) UpdateGroupPhoto(groupID string, photoPath string) error {
	_, err := db.c.Exec(`
		UPDATE groups SET photo = ? WHERE id = ?
	`, photoPath, groupID)
	return err
}
