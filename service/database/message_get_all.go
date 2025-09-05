package database

import "github.com/GioiaZheng/Wasa_proj/service/models"

// GetAllMessages returns all messages (used by admin/tests).
// FIX: check rows.Err() after iteration to capture driver-side errors.
func (db *appdbimpl) GetAllMessages() ([]models.Message, error) {
	rows, err := db.c.Query(`
		SELECT id, content, sender_id, receiver_id, group_id, conversation_id, created_at
		FROM messages
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.Content, &m.SenderID, &m.ReceiverID, &m.GroupID, &m.ConversationID, &m.CreatedAt); err != nil {
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
