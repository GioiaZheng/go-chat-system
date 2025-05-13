package database

import (
	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// GetGroupByName retrieves a group by its name
func (db *appdbimpl) GetGroupByName(name string) (models.Group, error) {
	var group models.Group
	err := db.c.QueryRow(`
		SELECT id, name FROM groups WHERE name = ?
	`, name).Scan(&group.ID, &group.Name)
	if err != nil {
		return models.Group{}, err
	}
	return group, nil
}
