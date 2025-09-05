package database

import "github.com/GioiaZheng/Wasa_proj/service/models"

// GetMessageComments returns comments for a message if stored separately.
// FIX: check rows.Err() after iteration. If your schema doesn't use a separate
// table for comments, consider returning an empty slice or adapt the query.
func (db *appdbimpl) GetMessageComments(messageID string) ([]models.Message, error) {
	rows, err := db.c.Query(`
		SELECT id, content, sender_id, conversation_id, created_at
		FROM message_comments
		WHERE message_id = ?
		ORDER BY created_at ASC
	`, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.Content, &m.SenderID, &m.ConversationID, &m.CreatedAt); err != nil {
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
