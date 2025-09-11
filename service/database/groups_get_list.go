package database

import (
	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// GetGroupsList returns all groups the user belongs to, including their members.
func (db *appdbimpl) GetGroupsList(userID string) ([]models.Group, error) {
	rows, err := db.c.Query(`
		SELECT g.id, g.name
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(&group.ID, &group.Name); err != nil {
			return nil, err
		}

		// Enrich group with its members
		members, err := db.GetGroup(group.ID)
		if err != nil {
			return nil, err
		}
		group.Members = members.Members

		groups = append(groups, group)
	}

	// FIX: must check rows.Err() after iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}
