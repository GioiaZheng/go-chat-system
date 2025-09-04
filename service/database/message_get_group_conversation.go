package database

import (
	"database/sql"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

func (db *appdbimpl) GetGroupConversation(groupID string) ([]models.Message, error) {
	rows, err := db.c.Query(`
		SELECT id, sender_id, group_id, content, created_at
		FROM messages
		WHERE group_id = ?
		ORDER BY created_at ASC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var (
			id, senderID, gid, content, createdAt sql.NullString
		)
		if err := rows.Scan(&id, &senderID, &gid, &content, &createdAt); err != nil {
			return nil, err
		}
		if !id.Valid {
			continue
		}
		messages = append(messages, models.Message{
			ID:        id.String,
			Content:   content.String,
			SenderID:  senderID.String,
			GroupID:   gid.String,    // 为空时为 ""
			CreatedAt: createdAt.String,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}
