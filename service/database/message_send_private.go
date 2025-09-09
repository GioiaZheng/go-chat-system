package database

import "github.com/GioiaZheng/Wasa_proj/service/models"

// SendPrivateMessage inserts a direct message to a single receiver (legacy private chat).
func (db *appdbimpl) SendPrivateMessage(message models.Message) error {
	_, err := db.c.Exec(`
		INSERT INTO messages (id, content, sender_id, receiver_id, created_at)
		VALUES (?,  ?,       ?,         ?,           COALESCE(?, datetime('now')))
	`, message.ID, message.Content, message.SenderID, message.ReceiverID, nullIfEmpty(message.CreatedAt))
	return err
}
