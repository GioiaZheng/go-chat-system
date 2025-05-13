package database

import (
	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// CreateGroup creates a new group in the database
func (db *appdbimpl) CreateGroup(group models.Group) error {
	tx, err := db.c.Begin()
	if err != nil {
		return err
	}

	// Insert group
	res, err := tx.Exec(`INSERT INTO groups (name) VALUES (?)`, group.Name)
	if err != nil {
		tx.Rollback()
		return err
	}

	groupID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	// Insert group members
	for _, member := range group.Members {
		_, err := tx.Exec(`
			INSERT INTO group_members (group_id, user_id, role)
			VALUES (?, ?, 'member')
		`, groupID, member.ID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
