package database

import (
	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// GetGroup retrieves a group by its ID
func (db *appdbimpl) GetGroup(groupID string) (models.Group, error) {
	var group models.Group
	err := db.c.QueryRow(`
		SELECT id, name FROM groups WHERE id = ?
	`, groupID).Scan(&group.ID, &group.Name)
	if err != nil {
		return models.Group{}, err
	}

	rows, err := db.c.Query(`
		SELECT u.id, u.username, u.avatar_url, gm.role
		FROM group_members gm
		JOIN users u ON u.id = gm.user_id
		WHERE gm.group_id = ?
	`, groupID)
	if err != nil {
		return models.Group{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var member models.User
		var role string
		if err := rows.Scan(&member.ID, &member.Username, &member.AvatarURL, &role); err != nil {
			return models.Group{}, err
		}
		group.Members = append(group.Members, member)
	}

	return group, nil
}
