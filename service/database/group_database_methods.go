package database

import (
	"database/sql"
	"errors"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// NOTE (English):
// This file consolidates ALL group-related DB methods on *appdbimpl* so we don't
// scatter implementations across multiple files (which caused redeclare errors).
// It matches the AppDatabase interface method names used by the API layer.

// ----------------------------------------------------------------------------
// Basic fetchers
// ----------------------------------------------------------------------------

// GetGroup fetches a group row by its ID.
// Matches interface: GetGroup(groupID string) (models.Group, error)
func (db *appdbimpl) GetGroup(groupID string) (models.Group, error) {
	var g models.Group
	err := db.c.QueryRow(`
		SELECT id, name, avatar_url
		FROM groups
		WHERE id = ?
	`, groupID).Scan(&g.ID, &g.Name, &g.AvatarUrl)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Group{}, err
		}
		return models.Group{}, err
	}
	return g, nil
}

// GetGroupsList lists groups where the given user is a member.
// Matches interface: GetGroupsList(userID string) ([]models.Group, error)
func (db *appdbimpl) GetGroupsList(userID string) ([]models.Group, error) {
	rows, err := db.c.Query(`
		SELECT g.id, g.name, g.avatar_url
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = ?
		ORDER BY g.name COLLATE NOCASE ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Group
	for rows.Next() {
		var g models.Group
		if err := rows.Scan(&g.ID, &g.Name, &g.AvatarUrl); err != nil {
			return nil, err
		}
		result = append(result, g)
	}
	return result, rows.Err()
}

// GetGroupMembers lists the users in a given group.
// Matches interface: GetGroupMembers(groupID string) ([]models.User, error)
func (db *appdbimpl) GetGroupMembers(groupID string) ([]models.User, error) {
	rows, err := db.c.Query(`
		SELECT u.id, u.username, u.name, u.avatar_url
		FROM users u
		JOIN group_members gm ON u.id = gm.user_id
		WHERE gm.group_id = ?
		ORDER BY u.username COLLATE NOCASE ASC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Name, &u.AvatarUrl); err != nil {
			return nil, err
		}
		members = append(members, u)
	}
	return members, rows.Err()
}

// GetGroupByName fetches a group by its unique name (if unique in your schema).
// Matches interface: GetGroupByName(name string) (models.Group, error)
func (db *appdbimpl) GetGroupByName(name string) (models.Group, error) {
	var g models.Group
	err := db.c.QueryRow(`
		SELECT id, name, avatar_url
		FROM groups
		WHERE name = ?
	`, name).Scan(&g.ID, &g.Name, &g.AvatarUrl)
	if err != nil {
		return models.Group{}, err
	}
	return g, nil
}

// ----------------------------------------------------------------------------
// Mutations (create / add / update / leave)
// ----------------------------------------------------------------------------

// CreateGroup creates a new group using the provided string ID (UUID).
// NOTE: The 'groups' table should define `id TEXT PRIMARY KEY`, so we insert the ID explicitly.
// Members are NOT inserted here; the API will call AddGroupMembers afterwards.
func (db *appdbimpl) CreateGroup(group models.Group) error {
	_, err := db.c.Exec(`
		INSERT INTO groups (id, name, avatar_url)
		VALUES (?, ?, COALESCE(?, NULL))
	`, group.ID, group.Name, group.AvatarUrl)
	return err
}

// AddGroupMembers adds one or more users into a group.
// Matches interface: AddGroupMembers(groupID string, userIDs []string) error
// NOTE: Ensure duplicates are handled by schema constraints (PRIMARY KEY or UNIQUE on (group_id,user_id)).
func (db *appdbimpl) AddGroupMembers(groupID string, memberIDs []string) error {
	stmt, err := db.c.Prepare(`
		INSERT OR IGNORE INTO group_members (group_id, user_id)
		VALUES (?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, mid := range memberIDs {
		if _, err := stmt.Exec(groupID, mid); err != nil {
			return err
		}
	}
	return nil
}

// AddMemberToGroup is a thin wrapper over AddGroupMembers for single user insert.
// The 'role' parameter is accepted for compatibility but unused by the current schema.
// Matches interface: AddMemberToGroup(groupID, userID, role string) error
func (db *appdbimpl) AddMemberToGroup(groupID, userID, role string) error {
	return db.AddGroupMembers(groupID, []string{userID})
}

// UpdateGroupName updates the name of a group.
// Matches interface: UpdateGroupName(groupID, name string) error
func (db *appdbimpl) UpdateGroupName(groupID, newName string) error {
	_, err := db.c.Exec(`
		UPDATE groups SET name = ? WHERE id = ?
	`, newName, groupID)
	return err
}

// UpdateGroupPhoto updates the avatar/photo URL of a group.
// Matches interface: UpdateGroupPhoto(groupID, photoPath string) error
func (db *appdbimpl) UpdateGroupPhoto(groupID, photoURL string) error {
	_, err := db.c.Exec(`
		UPDATE groups SET avatar_url = ? WHERE id = ?
	`, photoURL, groupID)
	return err
}

// SetGroupPhoto is a backward-compatible alias to UpdateGroupPhoto.
// Matches interface: SetGroupPhoto(groupID, photoUrl string) error
func (db *appdbimpl) SetGroupPhoto(groupID, photoUrl string) error {
	return db.UpdateGroupPhoto(groupID, photoUrl)
}

// LeaveGroup removes a user from a group.
// Matches interface: LeaveGroup(groupID, userID string) error
func (db *appdbimpl) LeaveGroup(groupID, userID string) error {
	_, err := db.c.Exec(`
		DELETE FROM group_members WHERE group_id = ? AND user_id = ?
	`, groupID, userID)
	return err
}

// IsGroupMember checks if a user is a member of a given group.
// Matches interface: IsGroupMember(userID, groupID string) (bool, error)
func (db *appdbimpl) IsGroupMember(userID, groupID string) (bool, error) {
	row := db.c.QueryRow(`
		SELECT COUNT(1)
		FROM group_members
		WHERE user_id = ? AND group_id = ?
	`, userID, groupID)

	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}
