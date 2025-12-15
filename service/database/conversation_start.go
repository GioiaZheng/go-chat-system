// file: service/database/conversation_database_methods.go
package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/google/uuid"
)

/*
StartConversation — create a conversation and insert members.
Handles private chat reuse automatically.
*/
func (db *appdbimpl) StartConversation(
	ctx context.Context,
	userID string,
	memberIDs []string,
	name string,
) (models.Conversation, error) {

	userID = strings.TrimSpace(userID)

	// Normalize + dedupe
	all := make([]string, 0, len(memberIDs)+1)
	others := make([]string, 0, len(memberIDs))
	seen := make(map[string]struct{})

	push := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		all = append(all, id)
		if id != userID {
			others = append(others, id)
		}
	}

	for _, id := range memberIDs {
		push(id)
	}
	push(userID)

	// Private chat detection: 1 other member
	if len(all) == 2 && len(others) == 1 {
		otherID := others[0]

		var existingID string
		err := db.c.QueryRowContext(ctx, `
			SELECT conversation_id 
			FROM conversation_members cm
			WHERE cm.user_id IN (?, ?)
			GROUP BY cm.conversation_id
			HAVING COUNT(DISTINCT cm.user_id) = 2
			AND COUNT(DISTINCT cm.user_id) = (
				SELECT COUNT(*)
					FROM conversation_members cm2
					WHERE cm2.conversation_id = cm.conversation_id
		)
		LIMIT 1
		`, userID, otherID).Scan(&existingID)

		if err == nil && existingID != "" {
			return db.buildConversationFromID(existingID, userID)
		}
	}

	// Create new conversation
	convID := uuid.NewString()

	tx, err := db.c.BeginTx(ctx, nil)
	if err != nil {
		return models.Conversation{}, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO conversations (id, name) VALUES (?, ?)`,
		convID, strings.TrimSpace(name),
	); err != nil {
		return models.Conversation{}, err
	}

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO conversation_members (conversation_id, user_id) VALUES (?, ?)`,
	)
	if err != nil {
		return models.Conversation{}, err
	}
	defer stmt.Close()

	for _, uid := range all {
		if _, err := stmt.ExecContext(ctx, convID, uid); err != nil {
			return models.Conversation{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return models.Conversation{}, err
	}

	return db.buildConversationFromID(convID, userID)
}

/*
buildConversationFromID — load full conversation info (name, participants, lastMessage, timestamps)
*/
func (db *appdbimpl) buildConversationFromID(cid string, requesterID string) (models.Conversation, error) {

	// Conversation base info: created_at is TEXT, not datetime
	var name string
	var avatarURL sql.NullString
	var createdAtStr string

	err := db.c.QueryRow(`
		SELECT name, avatar_url, created_at
		FROM conversations
		WHERE id = ?
	`, cid).Scan(&name, &avatarURL, &createdAtStr)
	if err != nil {
		return models.Conversation{}, err
	}

	// Parse created_at
	createdAt, _ := time.Parse(time.RFC3339, createdAtStr)

	// Load participants
	rows, err := db.c.Query(`
		SELECT u.id, u.name, COALESCE(u.avatar_url, '')
		FROM conversation_members cm
		JOIN users u ON u.id = cm.user_id
		WHERE cm.conversation_id = ?
		ORDER BY u.name ASC
	`, cid)
	if err != nil {
		return models.Conversation{}, err
	}
	defer rows.Close()

	var participants []models.User
	for rows.Next() {
		var u models.User
		var av string
		if err := rows.Scan(&u.ID, &u.Name, &av); err != nil {
			return models.Conversation{}, err
		}
		u.AvatarUrl = strings.TrimSpace(av)
		participants = append(participants, u)
	}

	if err := rows.Err(); err != nil {
		return models.Conversation{}, err
	}

	// Determine type
	convType := "group"
	if len(participants) == 2 {
		convType = "private"

                // Identify the peer for private conversations
                for _, p := range participants {
                        if p.ID != requesterID {
                                name = p.Name                  // Conversation name mirrors the peer
                                avatarURL.String = p.AvatarUrl // Conversation avatar mirrors the peer
                                break
                        }
                }
	}

	// LastMessage
	var msg models.Message
	var msgCreatedStr string

	err = db.c.QueryRow(`
		SELECT id, content, sender_id, conversation_id, created_at
		FROM messages
		WHERE conversation_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, cid).Scan(&msg.ID, &msg.Content, &msg.SenderID, &msg.ConversationID, &msgCreatedStr)

	var lastMessage *models.Message
	var updatedAt time.Time

	if err == nil {
		t, _ := time.Parse(time.RFC3339, msgCreatedStr)
		msg.CreatedAt = t.UTC().Format(time.RFC3339)
		lastMessage = &msg
		updatedAt = t
	} else {
		// no messages → updatedAt = created_at
		updatedAt = createdAt
		lastMessage = nil
	}

	return models.Conversation{
		ID:           cid,
		Type:         convType,
		Name:         strings.TrimSpace(name),
		AvatarUrl:    strings.TrimSpace(avatarURL.String),
		Participants: participants,
		LastMessage:  lastMessage,
		CreatedAt:    createdAt.UTC().Format(time.RFC3339),
		UpdatedAt:    updatedAt.UTC().Format(time.RFC3339),
	}, nil
}

/*
GetConversationMembers — ID only
*/
func (db *appdbimpl) GetConversationMembers(conversationID string) ([]string, error) {
	rows, err := db.c.Query(`
		SELECT user_id
		FROM conversation_members
		WHERE conversation_id = ?
	`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		members = append(members, uid)
	}
	return members, rows.Err()
}

/*
DeleteConversation
*/
func (db *appdbimpl) DeleteConversation(conversationID string) error {
	if conversationID == "" {
		return fmt.Errorf("empty conversation id")
	}

	var exists bool
	err := db.c.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM conversations WHERE id = ?)
	`, conversationID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("conversation not found")
	}

	_, err = db.c.Exec(`
		DELETE FROM conversations WHERE id = ?
	`, conversationID)
	return err
}

/*
GetMyConversations — correctly handle TEXT timestamps
*/
func (db *appdbimpl) GetMyConversations(userID string) ([]models.Conversation, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("empty userID")
	}

	rows, err := db.c.Query(`
		SELECT
			c.id,
			c.created_at,
			COALESCE(MAX(m.created_at), c.created_at) AS updated_at
		FROM conversations c
		JOIN conversation_members cm ON cm.conversation_id = c.id
		LEFT JOIN messages m ON m.conversation_id = c.id
		WHERE cm.user_id = ?
		GROUP BY c.id
		ORDER BY updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []models.Conversation{}

	for rows.Next() {
		var cid string
		var createdStr, updatedStr string

		if err := rows.Scan(&cid, &createdStr, &updatedStr); err != nil {
			return nil, err
		}

		conv, err := db.buildConversationFromID(cid, userID)
		if err != nil {
			return nil, err
		}

		result = append(result, conv)
	}
	return result, rows.Err()
}
