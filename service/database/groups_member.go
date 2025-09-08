package database

// IsGroupMember 检查用户是否是群组成员
func (db *appdbimpl) IsGroupMember(userID, groupID string) (bool, error) {
	row := db.c.QueryRow(`
		SELECT COUNT(*)
		FROM group_members
		WHERE user_id = ? AND group_id = ?
	`, userID, groupID)

	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}
