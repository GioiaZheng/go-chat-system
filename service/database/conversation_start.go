package database

import (
	"context"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/google/uuid"
)

func (db *appdbimpl) StartConversation(ctx context.Context, userID string, memberIDs []string, name string) (models.Conversation, error) {
	tx, err := db.c.BeginTx(ctx, nil)
	if err != nil {
		return models.Conversation{}, err
	}
	defer tx.Rollback()

	conversationID := uuid.NewString()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO conversations (id, name) VALUES (?, ?)`,
		conversationID, name,
	)
	if err != nil {
		return models.Conversation{}, err
	}

	for _, memberID := range append(memberIDs, userID) {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO conversation_members (conversation_id, user_id) VALUES (?, ?)`,
			conversationID, memberID,
		)
		if err != nil {
			return models.Conversation{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return models.Conversation{}, err
	}

	return models.Conversation{
		ID:   conversationID,
		Name: name,
	}, nil
}
