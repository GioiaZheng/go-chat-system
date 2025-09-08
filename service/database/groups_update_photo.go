package database

import (
	"database/sql"
	"fmt"
)

// UpdateGroupPhoto updates the group's photo URL.
// English notes:
// - Updates exactly one column (e.g., photo_url). This avoids ambiguity with other columns.
// - Returns nil on success (even if the URL is the same).
// - Returns error on DB failures or if the group does not exist.
func (db *appdbimpl) UpdateGroupPhoto(groupID string, url string) error {
	if groupID == "" {
		return fmt.Errorf("UpdateGroupPhoto: empty groupID")
	}
	// NOTE: adjust the column name if your schema uses a different one
	// e.g., "photo", "avatar_url", or "photo_url".
	const stmt = `
		UPDATE groups
		   SET photo_url = ?
		 WHERE id = ?
	`
	res, err := db.c.Exec(stmt, url, groupID)
	if err != nil {
		return fmt.Errorf("UpdateGroupPhoto: exec failed: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("UpdateGroupPhoto: rows affected: %w", err)
	}
	if n == 0 {
		// Not found: it's better to return a typed error upstream could map to 404/NotFound.
		return sql.ErrNoRows
	}
	return nil
}

// (Optional) Context-aware variant, if you prefer passing context from handlers.
// Keep or remove based on your project style.
// English notes:
// - If you standardize on context everywhere, uncomment this and deprecate the non-ctx variant.
/*
func (db *appdbimpl) UpdateGroupPhotoCtx(ctx context.Context, groupID string, url string) error {
	if groupID == "" {
		return fmt.Errorf("UpdateGroupPhotoCtx: empty groupID")
	}
	const stmt = `
		UPDATE groups
		   SET photo_url = ?
		 WHERE id = ?
	`
	res, err := db.c.ExecContext(ctx, stmt, url, groupID)
	if err != nil {
		return fmt.Errorf("UpdateGroupPhotoCtx: exec failed: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("UpdateGroupPhotoCtx: rows affected: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
*/
