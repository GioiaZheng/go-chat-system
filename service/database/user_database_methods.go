package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/GioiaZheng/Wasa_proj/service/models"
)

// 生成简单唯一ID（如需 UUID 可替换）
func newID() string {
	return fmt.Sprintf("u_%d", time.Now().UnixNano())
}

// GetUserByCredentials：支持简化登录 password==""（仅匹配存储为 NULL/"" 的用户）
func (db *appdbimpl) GetUserByCredentials(username, password string) (models.User, error) {
	var u models.User

	if password == "" {
		// 简化登录：只匹配空密码用户
		err := db.c.QueryRow(`
			SELECT id, username, name, COALESCE(avatar_url, ''), COALESCE(photo, '')
			FROM users
			WHERE username = ?
			  AND COALESCE(password, '') = ''
			LIMIT 1
		`, username).Scan(&u.ID, &u.Username, &u.Name, &u.AvatarUrl, &u.Photo)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return models.User{}, sql.ErrNoRows
			}
			return models.User{}, err
		}
		return u, nil
	}

	// 正常登录：明文匹配（如有 hash，自行替换）
	err := db.c.QueryRow(`
		SELECT id, username, name, COALESCE(avatar_url, ''), COALESCE(photo, '')
		FROM users
		WHERE username = ?
		  AND password = ?
		LIMIT 1
	`, username, password).Scan(&u.ID, &u.Username, &u.Name, &u.AvatarUrl, &u.Photo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, sql.ErrNoRows
		}
		return models.User{}, err
	}
	return u, nil
}

// CreateUser：按当前 users 表写入（含 password、avatar_url、photo）
// password 可为 ""（简化登录），AvatarUrl/Photo 可空
func (db *appdbimpl) CreateUser(u models.User, password string) (models.User, error) {
	if u.Username == "" {
		return models.User{}, errors.New("empty username")
	}
	if u.Name == "" {
		u.Name = u.Username
	}
	if u.ID == "" {
		u.ID = newID()
	}

	_, err := db.c.Exec(`
		INSERT INTO users (id, username, password, name, avatar_url, photo)
		VALUES (?, ?, ?, ?, ?, ?)
	`, u.ID, u.Username, password, u.Name, u.AvatarUrl, u.Photo)
	if err != nil {
		return models.User{}, err
	}

	// 返回创建后的完整用户
	return db.GetUserByID(u.ID)
}

// AuthenticateUser：实现接口，委托到 GetUserByCredentials
func (db *appdbimpl) AuthenticateUser(username, password string) (models.User, error) {
	return db.GetUserByCredentials(username, password)
}

// GetUser：实现接口，委托到 GetUserByID（若接口其实带 context，会有编译提示，我再给你 context 版）
func (db *appdbimpl) GetUser(userID string) (models.User, error) {
	return db.GetUserByID(userID)
}

// GetUserIDFromIdentifier：把“标识符”解析成用户ID
// 规则：
// 1) 先按 id 直接查（精确匹配 users.id）；
// 2) 不存在则按用户名（不区分大小写）查其 id；
func (db *appdbimpl) GetUserIDFromIdentifier(identifier string) (string, error) {
	ident := strings.TrimSpace(identifier)
	if ident == "" {
		return "", errors.New("empty identifier")
	}

	// 尝试把 ident 当成 id
	var id string
	err := db.c.QueryRow(`
		SELECT id FROM users WHERE id = ? LIMIT 1
	`, ident).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	// 回退：按用户名（不区分大小写）查 id
	err = db.c.QueryRow(`
		SELECT id FROM users WHERE lower(username) = lower(?) LIMIT 1
	`, ident).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

// SearchUsers 在 users 表中按用户名或昵称模糊搜索（不区分大小写）。
// - ctx:      上下文取消/超时
// - userID:   调用者自身ID（用于从结果中过滤自己）
// - query:    搜索关键词
// 返回最多 50 条，按 username 排序。
func (db *appdbimpl) SearchUsers(ctx context.Context, userID string, query string) ([]models.User, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return []models.User{}, nil
	}

	rows, err := db.c.QueryContext(ctx, `
		SELECT id, username, name, COALESCE(avatar_url, ''), COALESCE(photo, '')
		FROM users
		WHERE (username LIKE ? ESCAPE '\' OR name LIKE ? ESCAPE '\')
		  AND id <> ?
		ORDER BY username COLLATE NOCASE ASC
		LIMIT 50
	`, "%"+q+"%", "%"+q+"%", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Name, &u.AvatarUrl, &u.Photo); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUserName 更新用户的显示名称（无 context 版，匹配 AppDatabase 接口）
func (db *appdbimpl) UpdateUserName(userID string, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("empty name")
	}

	res, err := db.c.Exec(`
		UPDATE users SET name = ?
		WHERE id = ?
	`, name, userID)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
// UpdateUserPhoto：同时写入 photo 与 avatar_url，便于前端直接使用
func (db *appdbimpl) UpdateUserPhoto(userID string, photoPath string) error {
	photoPath = strings.TrimSpace(photoPath)
	if photoPath == "" {
		return errors.New("empty photo path")
	}

	res, err := db.c.Exec(`
		UPDATE users SET photo = ?, avatar_url = ?
		WHERE id = ?
	`, photoPath, photoPath, userID)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
