package database

import "github.com/GioiaZheng/Wasa_proj/service/models"

// GetMyConversations retrieves all conversations (direct and group) for a given user.
// FIX: after each rows iteration, check rows.Err() to capture driver-side errors.
func (db *appdbimpl) GetMyConversations(userID string) ([]models.Conversation, error) {
	var conversations []models.Conversation

	// --- Direct (private) conversations ---
	rows, err := db.c.Query(`
		SELECT u.id, u.username, u.avatar_url, m.content
		FROM messages m
		JOIN users u ON (m.sender_id = u.id OR m.receiver_id = u.id)
		WHERE (m.sender_id = ? OR m.receiver_id = ?) AND u.id != ?
		ORDER BY m.created_at DESC
	`, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seenUsers := make(map[string]bool)

	for rows.Next() {
		var conv models.Conversation
		if err := rows.Scan(&conv.ID, &conv.Name, &conv.AvatarUrl, &conv.LastMessage); err != nil {
			return nil, err
		}
		// Deduplicate by user id
		if seenUsers[conv.ID] {
			continue
		}
		seenUsers[conv.ID] = true
		conversations = append(conversations, conv)
	}

	// FIX: must check rows.Err() after iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// --- Group conversations ---
	groupRows, err := db.c.Query(`
		SELECT g.id, g.name, g.avatar_url, m.content
		FROM messages m
		JOIN groups g ON m.group_id = g.id
		JOIN group_members gm ON gm.group_id = g.id
		WHERE gm.user_id = ?
		ORDER BY m.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer groupRows.Close()

	seenGroups := make(map[string]bool)

	for groupRows.Next() {
		var conv models.Conversation
		if err := groupRows.Scan(&conv.ID, &conv.Name, &conv.AvatarUrl, &conv.LastMessage); err != nil {
			return nil, err
		}
		// Deduplicate by group id
		if seenGroups[conv.ID] {
			continue
		}
		seenGroups[conv.ID] = true
		conversations = append(conversations, conv)
	}

	// FIX: must check groupRows.Err() after iteration
	if err := groupRows.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}
