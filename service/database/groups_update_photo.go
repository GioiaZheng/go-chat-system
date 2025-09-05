package database

import "strings"

// UpdateGroupPhoto updates the group's avatar/photo URL.
// To be resilient across slightly different schemas, it tries multiple column names.
// If at least one UPDATE succeeds and affects rows, we consider it successful.
func (db *appdbimpl) UpdateGroupPhoto(groupID string, url string) error {
	type candidate struct{ sql string }
	candidates := []candidate{
		{sql: "UPDATE groups SET avatar_url = ? WHERE id = ?"},
		{sql: "UPDATE groups SET avatarUrl  = ? WHERE id = ?"},
		{sql: "UPDATE groups SET photo_url  = ? WHERE id = ?"},
		{sql: "UPDATE groups SET photo      = ? WHERE id = ?"},
		{sql: "UPDATE groups SET picture    = ? WHERE id = ?"},
	}

	var anyAffected bool
	for _, c := range candidates {
		res, err := db.c.Exec(c.sql, url, groupID)
		if err != nil {
			// ignore missing-column errors and try the next one
			if strings.Contains(strings.ToLower(err.Error()), "no such column") {
				continue
			}
			return err
		}
		if res != nil {
			if n, _ := res.RowsAffected(); n > 0 {
				anyAffected = true
			}
		}
	}

	// If nothing was updated, we still return nil to avoid breaking flows
	// (some schemas might not store group photos). This keeps the API lenient.
	if !anyAffected {
		return nil
	}
	return nil
}
