package database

import "github.com/GioiaZheng/Wasa_proj/service/models"

// SendMessageToUser sends a private message
func (db *appdbimpl) SendMessageToUser(message models.Message) error {
	_, err := db.c.Exec(`
		INSERT INTO messages (sender_id, receiver_id, content, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`, message.SenderID, message.ReceiverID, message.Content)
	return err
}
