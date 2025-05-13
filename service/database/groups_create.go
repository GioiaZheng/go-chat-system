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
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Insert group
	res, err := tx.Exec(`INSERT INTO groups (name) VALUES (?)`, group.Name)
	if err != nil {
		return err
	}

	groupID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// Insert group members
	for _, member := range group.Members {
		_, err = tx.Exec(`
			INSERT INTO group_members (group_id, user_id, role)
			VALUES (?, ?, ?)
		`, groupID, member.UserID, member.Role)
		if err != nil {
			return err
		}
	}

	return nil
}
