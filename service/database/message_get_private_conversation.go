package database

import "github.com/GioiaZheng/Wasa_proj/service/models"

// GetPrivateConversation returns ordered messages between two users (either direction).
// FIX: check rows.Err() after iteration.
func (db *appdbimpl) GetPrivateConversation(userA, userB string) ([]models.Message, error) {
	rows, err := db.c.Query(`
		SELECT id, content, sender_id, receiver_id, conversation_id, created_at
		FROM messages
		WHERE (sender_id = ? AND receiver_id = ?)
		   OR (sender_id = ? AND receiver_id = ?)
		ORDER BY created_at ASC
	`, userA, userB, userB, userA)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.Content, &m.SenderID, &m.ReceiverID, &m.ConversationID, &m.CreatedAt); err != nil {
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
