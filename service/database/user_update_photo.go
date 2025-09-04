package database

// UpdateUserPhoto updates both avatar_url and photo columns for a user.
// /users/me reads avatar_url into models.User.AvatarUrl, so we must update it.
// Keeping photo in sync avoids divergence with other parts that may read photo.
func (db *appdbimpl) UpdateUserPhoto(userID string, photoPath string) error {
	_, err := db.c.Exec(
		`UPDATE users
		   SET avatar_url = ?, photo = ?
		 WHERE id = ?`,
		photoPath, photoPath, userID,
	)
	return err
}
