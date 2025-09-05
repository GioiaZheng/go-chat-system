package database

import "github.com/GioiaZheng/Wasa_proj/service/models"

// GetGroupConversation returns the ordered messages of a group.
// FIX: check rows.Err() after iteration.
func (db *appdbimpl) GetGroupConversation(groupID string) ([]models.Message, error) {
	rows, err := db.c.Query(`
		SELECT id, content, sender_id, group_id, conversation_id, created_at
		FROM messages
		WHERE group_id = ?
		ORDER BY created_at ASC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.Content, &m.SenderID, &m.GroupID, &m.ConversationID, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}

	// FIX: must check rows.Err() after iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
