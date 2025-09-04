package database

import (
	"database/sql"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

func (db *appdbimpl) GetPrivateConversation(userID1, userID2 string) ([]models.Message, error) {
	rows, err := db.c.Query(`
		SELECT id, sender_id, receiver_id, content, created_at
		FROM messages
		WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)
		ORDER BY created_at ASC
	`, userID1, userID2, userID2, userID1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var (
			id, senderID, receiverID, content, createdAt sql.NullString
		)
		if err := rows.Scan(&id, &senderID, &receiverID, &content, &createdAt); err != nil {
			return nil, err
		}
		// 若出现“全 NULL”或 id 为空，跳过这行，避免 Scan NULL→string 报错
		if !id.Valid {
			continue
		}
		messages = append(messages, models.Message{
			ID:         id.String,
			Content:    content.String,
			SenderID:   senderID.String,
			ReceiverID: receiverID.String, // 为空时为 ""，符合你的 json omitempty
			CreatedAt:  createdAt.String,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}
