package database

// AddMemberToGroup adds a single user to a group
func (db *appdbimpl) AddMemberToGroup(groupID string, userID string, role string) error {
	_, err := db.c.Exec(`
		INSERT INTO group_members (group_id, user_id, role)
		VALUES (?, ?, ?)
	`, groupID, userID, role)
	return err
}

// AddGroupMembers adds multiple users to a group
func (db *appdbimpl) AddGroupMembers(groupID string, userIDs []string) error {
	tx, err := db.c.Begin()
	if err != nil {
		return err
	}

	for _, userID := range userIDs {
		_, err := tx.Exec(`
			INSERT INTO group_members (group_id, user_id, role)
			VALUES (?, ?, 'member')
		`, groupID, userID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
