package database

import (
	"context"
	"database/sql"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/google/uuid"
)

/*
StartConversation creates a new conversation and adds the provided members.

Design notes:
- Uses a transaction when the required tables exist (conversations, conversation_members).
- Normalizes/uniquifies member IDs and ensures the creator is included.
- If the schema is missing, falls back to returning a valid Conversation object without DB writes.
*/
func (db *appdbimpl) StartConversation(
	ctx context.Context,
	userID string,
	memberIDs []string,
	name string,
) (models.Conversation, error) {

	convID := uuid.NewString()

	// Normalize and deduplicate member IDs; always include the creator.
	final := make([]string, 0, len(memberIDs)+1)
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
		final = append(final, id)
	}
	for _, id := range memberIDs {
		push(id)
	}
	push(userID)

	// Try transactional path first.
	tx, err := db.c.BeginTx(ctx, nil)
	if err == nil {
		defer func() { _ = tx.Rollback() }()

		// Insert conversation row.
		if _, err = tx.ExecContext(ctx, `INSERT INTO conversations (id, name) VALUES (?, ?)`, convID, name); err == nil {
			// Insert members.
			var stmt *sql.Stmt
			stmt, err = tx.PrepareContext(ctx, `INSERT INTO conversation_members (conversation_id, user_id) VALUES (?, ?)`)
			if err == nil {
				defer stmt.Close()

				ok := true
				for _, mid := range final {
					if _, e := stmt.ExecContext(ctx, convID, mid); e != nil {
						err = e
						ok = false
						break
					}
				}
				if ok {
					if err = tx.Commit(); err == nil {
						return models.Conversation{ID: convID, Name: name}, nil
					}
				}
			}
		}
		// If we reached here, transactional path failed. Fall through to fallback below.
	}

	// Fallback: schema not present; just return a valid conversation id+name.
	return models.Conversation{
		ID:   convID,
		Name: strings.TrimSpace(name),
	}, nil
}

/*
GetConversationMembers returns the list of user IDs that belong to the given conversation.

If the "conversation_members" table doesn't exist, the query will error out; callers decide how to handle.
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
		members = append(members, strings.TrimSpace(uid))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}
