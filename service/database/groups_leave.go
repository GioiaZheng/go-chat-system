package database

// LeaveGroup removes a user from a group
func (db *appdbimpl) LeaveGroup(groupID string, userID string) error {
	_, err := db.c.Exec(`
		DELETE FROM group_members WHERE group_id = ? AND user_id = ?
	`, groupID, userID)
	return err
}
