package database

import "github.com/GioiaZheng/Wasa_proj/service/models"

// SendPrivateMessage inserts a direct message to a single receiver (legacy private chat).
func (db *appdbimpl) SendPrivateMessage(message models.Message) error {
	if message.Type == "" {
		message.Type = "text"
	}
	if message.Status == "" {
		message.Status = "sent"
	}
	_, err := db.c.Exec(`
		INSERT INTO messages (id, content, sender_id, receiver_id, type, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, COALESCE(?, datetime('now')))
	`, message.ID, message.Content, message.SenderID, message.ReceiverID,
		message.Type, message.Status, nullIfEmpty(message.CreatedAt))
	return err
}
