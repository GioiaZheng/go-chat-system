package database

import "github.com/GioiaZheng/Wasa_proj/service/models"

// SendPrivateMessage sends a private message (same as SendMessageToUser, slightly different)
func (db *appdbimpl) SendPrivateMessage(message models.Message) error {
	_, err := db.c.Exec(`
		INSERT INTO messages (sender_id, receiver_id, content, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`, message.SenderID, message.ReceiverID, message.Content)
	return err
}
