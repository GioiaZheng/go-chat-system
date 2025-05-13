package database

// AreFriends checks if two users are friends
func (db *appdbimpl) AreFriends(userID1 string, userID2 string) bool {
	var count int
	err := db.c.QueryRow(`
		SELECT COUNT(*) FROM friends
		WHERE (user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)
	`, userID1, userID2, userID2, userID1).Scan(&count)
	return err == nil && count > 0
}
