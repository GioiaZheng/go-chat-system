package database

import "github.com/GioiaZheng/Wasa_proj/service/models"

// SendMessageToConversation inserts a message into a conversation (group/multi-party).
func (db *appdbimpl) SendMessageToConversation(message models.Message) error {
	if message.Type == "" {
		message.Type = "text"
	}
	if message.Status == "" {
		message.Status = "sent"
	}
	_, err := db.c.Exec(`
		INSERT INTO messages (id, content, sender_id, conversation_id, type, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, COALESCE(?, datetime('now')))
	`, message.ID, message.Content, message.SenderID, message.ConversationID,
		message.Type, message.Status, nullIfEmpty(message.CreatedAt))
	return err
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
