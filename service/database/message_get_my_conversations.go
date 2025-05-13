package database

import "github.com/GioiaZheng/Wasa_proj/service/models"

func (db *appdbimpl) GetMyConversations(userID string) ([]models.Conversation, error) {
	var conversations []models.Conversation

	rows, err := db.c.Query(`
		SELECT u.id, u.username, u.avatar_url, m.content
		FROM messages m
		JOIN users u ON (m.sender_id = u.id OR m.receiver_id = u.id)
		WHERE (m.sender_id = ? OR m.receiver_id = ?) AND u.id != ?
		GROUP BY u.id
		ORDER BY MAX(m.created_at) DESC
	`, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var conv models.Conversation
		if err := rows.Scan(&conv.ID, &conv.Name, &conv.AvatarURL, &conv.LastMsg); err != nil {
			return nil, err
		}
		conversations = append(conversations, conv)
	}

	groupRows, err := db.c.Query(`
		SELECT g.id, g.name, g.photo, m.content
		FROM messages m
		JOIN groups g ON m.group_id = g.id
		JOIN group_members gm ON gm.group_id = g.id
		WHERE gm.user_id = ?
		GROUP BY g.id
		ORDER BY MAX(m.created_at) DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer groupRows.Close()

	for groupRows.Next() {
		var conv models.Conversation
		if err := groupRows.Scan(&conv.ID, &conv.Name, &conv.AvatarURL, &conv.LastMsg); err != nil {
			return nil, err
		}
		conversations = append(conversations, conv)
	}

	return conversations, nil
}
