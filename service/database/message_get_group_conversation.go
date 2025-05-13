package database

import "github.com/GioiaZheng/Wasa_proj/service/models"

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
		var m models.Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.GroupID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}
