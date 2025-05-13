package database

import (
	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// GetFriendsList retrieves the list of friends for a given user
func (db *appdbimpl) GetFriendsList(userID string) ([]models.Friend, error) {
	rows, err := db.c.Query(`
		SELECT u.id, u.username, u.avatar_url
		FROM friends f
		JOIN users u ON u.id = f.friend_id
		WHERE f.user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []models.Friend
	for rows.Next() {
		var friend models.Friend
		if err := rows.Scan(&friend.UserID, &friend.UserName, &friend.AvatarUrl); err != nil {
			return nil, err
		}
		friends = append(friends, friend)
	}
	return friends, nil
}

// GetUserFriends retrieves the list of a user's friends (alternative version)
func (db *appdbimpl) GetUserFriends(userID string) ([]models.Friend, error) {
	return db.GetFriendsList(userID)
}
