package database

import (
	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// CreateGroup creates a new group using the provided string ID (UUID).
// NOTE:
// - The 'groups' table defines `id TEXT PRIMARY KEY`, so we must insert the ID explicitly.
// - Members are NOT inserted here; API layer will call AddGroupMembers after this.
func (db *appdbimpl) CreateGroup(group models.Group) error {
	_, err := db.c.Exec(`
		INSERT INTO groups (id, name, avatar_url)
		VALUES (?, ?, COALESCE(?, NULL))
	`, group.ID, group.Name, group.AvatarUrl)
	return err
}
