package database

import (
	"database/sql"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// GetMessageByID fetches a message by its id, regardless of chat type.
func (db *appdbimpl) GetMessageByID(messageID string) (models.Message, error) {
	var m models.Message
	var convID sql.NullString
	err := db.c.QueryRow(`
		SELECT id, content, sender_id, receiver_id, conversation_id,
		       COALESCE(created_at, '') AS created_at,
		       type, status
		  FROM messages
		 WHERE id = ?
	`, messageID).Scan(&m.ID, &m.Content, &m.SenderID, &m.ReceiverID, &convID, &m.CreatedAt, &m.Type, &m.Status)
	if err != nil {
		return models.Message{}, err
	}
	m.ConversationID = convID.String
	return m, nil
}
