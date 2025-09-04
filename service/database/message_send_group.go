package database

import (
	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// SendGroupMessage inserts a new message into a group chat
func (db *appdbimpl) SendGroupMessage(message models.Message) error {
	_, err := db.c.Exec(`
		INSERT INTO messages (sender_id, group_id, content, conversation_id, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, message.SenderID, message.GroupID, message.Content, message.ConversationID)
	return err
}
