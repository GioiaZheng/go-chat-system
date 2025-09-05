package database

import (
	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// GetGroup retrieves a group by its ID and includes its members.
// FIX: after rows iteration, always check rows.Err() to catch latent driver errors.
func (db *appdbimpl) GetGroup(groupID string) (models.Group, error) {
	var group models.Group

	// Fetch group basics
	err := db.c.QueryRow(`
		SELECT id, name, avatar_url, created_at
		FROM groups
		WHERE id = ?
	`, groupID).Scan(&group.ID, &group.Name, &group.AvatarUrl, &group.CreatedAt)
	if err != nil {
		return models.Group{}, err
	}

	// Fetch members
	rows, err := db.c.Query(`
		SELECT u.id, u.username, u.avatar_url, gm.role
		FROM group_members gm
		JOIN users u ON u.id = gm.user_id
		WHERE gm.group_id = ?
		ORDER BY u.username ASC
	`, groupID)
	if err != nil {
		return models.Group{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var member models.GroupMember
		if err := rows.Scan(&member.UserID, &member.UserName, &member.AvatarUrl, &member.Role); err != nil {
			return models.Group{}, err
		}
		group.Members = append(group.Members, member)
	}

	// FIX: must check rows.Err() after iteration to catch any deferred scan/driver errors.
	if err := rows.Err(); err != nil {
		return models.Group{}, err
	}

	return group, nil
}
