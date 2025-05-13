package database

// AreFriends checks if two users are friends
func (db *appdbimpl) AreFriends(userID1, userID2 string) (bool, error) {
	var count int
	err := db.c.QueryRow(`
		SELECT COUNT(*)
		FROM friends
		WHERE (user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)
	`, userID1, userID2, userID2, userID1).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
