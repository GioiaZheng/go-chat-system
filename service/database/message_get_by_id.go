package database

import (
	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// GetMessageByID fetches a single message by its ID
func (db *appdbimpl) GetMessageByID(messageID string) (models.Message, error) {
	var m models.Message
	err := db.c.QueryRow(`
		SELECT id, sender_id, receiver_id, group_id, content, created_at
		FROM messages
		WHERE id = ?
	`, messageID).Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.GroupID, &m.Content, &m.CreatedAt)
	if err != nil {
		return models.Message{}, err
	}
	return m, nil
}
