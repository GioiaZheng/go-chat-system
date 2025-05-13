package database

import (
	"context"
	"fmt"
)

func (db *appdbimpl) DeleteMessage(userID string, messageID string) error {
	res, err := db.c.ExecContext(context.Background(),
		`DELETE FROM messages WHERE id = ? AND sender_id = ?`,
		messageID, userID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("unauthorized or message not found")
	}
	return nil
}
