package database

// SetGroupPhoto is kept for backward compatibility with the AppDatabase interface.
// It simply delegates to UpdateGroupPhoto.
//
// English notes:
// - The interface AppDatabase still requires SetGroupPhoto.
// - We forward to UpdateGroupPhoto to avoid duplicate logic.
func (db *appdbimpl) SetGroupPhoto(groupID string, url string) error {
	return db.UpdateGroupPhoto(groupID, url)
}
