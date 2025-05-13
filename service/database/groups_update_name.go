package database

// UpdateGroupName updates the name of a group
func (db *appdbimpl) UpdateGroupName(groupID string, name string) error {
	_, err := db.c.Exec(`
		UPDATE groups SET name = ? WHERE id = ?
	`, name, groupID)
	return err
}
