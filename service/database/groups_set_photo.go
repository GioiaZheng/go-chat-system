package database

func (db *appdbimpl) SetGroupPhoto(groupID string, photoUrl string) error {
	_, err := db.c.Exec(`UPDATE groups SET avatar_url = ? WHERE id = ?`, photoUrl, groupID)
	return err
}
