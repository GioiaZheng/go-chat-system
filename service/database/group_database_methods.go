// group_database_methods.go implements group persistence and member lookups.
// Related files: service/api/groups.go, service/models/models.go.
package database

import (
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// GetGroup fetches a group row by its ID and populates member information.
// Matches interface: GetGroup(groupID string) (models.Group, error).
func (db *appdbimpl) GetGroup(groupID string) (models.Group, error) {
	var g models.Group
	var createdAt sql.NullString

	err := db.c.QueryRow(`
		SELECT id, name, avatar_url, conversation_id, created_at
		FROM groups
		WHERE id = ?
	 `, groupID).Scan(&g.ID, &g.Name, &g.AvatarUrl, &g.ConversationID, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Group{}, err
		}
		return models.Group{}, err
	}

	if createdAt.Valid {
		g.CreatedAt = normalizeTimestamp(createdAt.String)
	}

	members, err := db.getGroupMembersWithRole(groupID)
	if err != nil {
		return models.Group{}, err
	}
	g.Members = members

	return g, nil
}

// GetGroupsList lists groups where the given user is a member.
// Matches interface: GetGroupsList(userID string) ([]models.Group, error).
func (db *appdbimpl) GetGroupsList(userID string) ([]models.Group, error) {
	rows, err := db.c.Query(`
		SELECT g.id, g.name, g.avatar_url, g.conversation_id, g.created_at
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = ?
		ORDER BY g.name COLLATE NOCASE ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]models.Group, 0)
	for rows.Next() {
		var g models.Group
		var createdAt sql.NullString

		if err := rows.Scan(&g.ID, &g.Name, &g.AvatarUrl, &g.ConversationID, &createdAt); err != nil {
			return nil, err
		}

		if createdAt.Valid {
			g.CreatedAt = normalizeTimestamp(createdAt.String)
		}

		result = append(result, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// getGroupMembersWithRole returns group members along with their stored role.
func (db *appdbimpl) getGroupMembersWithRole(groupID string) ([]models.GroupMember, error) {
	rows, err := db.c.Query(`
                SELECT gm.user_id, u.name, COALESCE(gm.role, ''), COALESCE(u.avatar_url, '')
                FROM group_members gm
                JOIN users u ON u.id = gm.user_id
                WHERE gm.group_id = ?
                ORDER BY u.name COLLATE NOCASE ASC
        `, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]models.GroupMember, 0, 8)
	for rows.Next() {
		var gm models.GroupMember
		var role, avatar string
		if err := rows.Scan(&gm.UserID, &gm.Name, &role, &avatar); err != nil {
			return nil, err
		}
		gm.ID = gm.UserID
		gm.Role = strings.TrimSpace(role)
		if gm.Role == "" {
			gm.Role = "member"
		}
		gm.AvatarUrl = strings.TrimSpace(avatar)
		members = append(members, gm)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}

// normalizeTimestamp converts supported timestamps to RFC3339 while preserving
// any unrecognized formats unchanged.
func normalizeTimestamp(ts string) string {
	ts = strings.TrimSpace(ts)
	if ts == "" {
		return ts
	}

	if t, err := time.Parse(time.RFC3339, ts); err == nil {
		return t.UTC().Format(time.RFC3339)
	}
	if t, err := time.Parse("2006-01-02 15:04:05", ts); err == nil {
		return t.UTC().Format(time.RFC3339)
	}
	return ts
}

// GetGroupMembers lists the users in a given group.
// Matches interface: GetGroupMembers(groupID string) ([]models.User, error).
// The users table only includes id, name, and avatar_url columns.
func (db *appdbimpl) GetGroupMembers(groupID string) ([]models.User, error) {
	rows, err := db.c.Query(`
		SELECT u.id, u.name, u.avatar_url
		FROM users u
		JOIN group_members gm ON u.id = gm.user_id
		WHERE gm.group_id = ?
		ORDER BY u.name COLLATE NOCASE ASC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.AvatarUrl); err != nil {
			return nil, err
		}
		members = append(members, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

// GetGroupByName fetches a group by its unique name (if unique in your schema).
// Matches interface: GetGroupByName(name string) (models.Group, error).
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

// CreateGroup creates a new group using the provided string ID (UUID).
// The "groups" table should define id TEXT PRIMARY KEY, so the ID is inserted explicitly.
// Members are not inserted here; the API will call AddGroupMembers afterwards.
func (db *appdbimpl) CreateGroup(group models.Group) error {
	_, err := db.c.Exec(`
		INSERT INTO groups (id, name, avatar_url, conversation_id)
		VALUES (?, ?, COALESCE(?, NULL), ?)
	`, group.ID, group.Name, group.AvatarUrl, group.ConversationID)
	if err != nil {
		log.Printf("insert group error: %v", err)
	}
	return err
}

// AddGroupMembers adds one or more users into a group.
// Matches interface: AddGroupMembers(groupID string, userIDs []string) error.
// Duplicates are ignored via schema constraints on (group_id, user_id).
func (db *appdbimpl) AddGroupMembers(groupID string, memberIDs []string) error {
	stmt, err := db.c.Prepare(`
		INSERT OR IGNORE INTO group_members (group_id, user_id, role)
		VALUES (?, ?, 'member')
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, mid := range memberIDs {
		if _, err := stmt.Exec(groupID, mid); err != nil {
			log.Printf("insert group_members error: %v", err)
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
	tx, err := db.c.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`
                UPDATE groups SET name = ? WHERE id = ?
        `, newName, groupID); err != nil {
		return err
	}

	if _, err := tx.Exec(`
                UPDATE conversations
                   SET name = ?
                 WHERE id = (
                         SELECT conversation_id FROM groups WHERE id = ?
                 )
        `, newName, groupID); err != nil {
		return err
	}

	return tx.Commit()
}

// UpdateGroupPhoto updates the avatar/photo URL of a group.
// Matches interface: UpdateGroupPhoto(groupID, photoPath string) error
func (db *appdbimpl) UpdateGroupPhoto(groupID, photoURL string) error {
	tx, err := db.c.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`
                UPDATE groups SET avatar_url = ? WHERE id = ?
        `, photoURL, groupID); err != nil {
		return err
	}

	if _, err := tx.Exec(`
                UPDATE conversations
                   SET avatar_url = ?
                 WHERE id = (
                         SELECT conversation_id FROM groups WHERE id = ?
                 )
        `, photoURL, groupID); err != nil {
		return err
	}

	return tx.Commit()
}

// SetGroupPhoto is a backward-compatible alias to UpdateGroupPhoto.
// Matches interface: SetGroupPhoto(groupID, photoUrl string) error
func (db *appdbimpl) SetGroupPhoto(groupID, photoUrl string) error {
	return db.UpdateGroupPhoto(groupID, photoUrl)
}

// LeaveGroup removes a user from a group.
// Matches interface: LeaveGroup(groupID, userID string) error
func (db *appdbimpl) LeaveGroup(groupID, userID string) error {
	res, err := db.c.Exec(`
                DELETE FROM group_members WHERE group_id = ? AND user_id = ?
        `, groupID, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
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
