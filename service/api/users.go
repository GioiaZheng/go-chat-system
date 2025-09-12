package api

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

//
// ────────────────────────────────────────────────────────────────────────────────
//  Helpers (request bodies)
// ────────────────────────────────────────────────────────────────────────────────
//

// setUsernameBody supports both { "username": "..." } and legacy { "name": "..." }.
type setUsernameBody struct {
	Username string `json:"username,omitempty"`
	Name     string `json:"name,omitempty"`
}

//
// ────────────────────────────────────────────────────────────────────────────────
//  Get my info  (GET /me or /users/me)
// ────────────────────────────────────────────────────────────────────────────────
//

// getUserInfo returns the current authenticated user profile (from ctx.UserID).
func (rt *_router) getUserInfo(
	w http.ResponseWriter,
	_ *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	uid := strings.TrimSpace(ctx.UserID)
	if uid == "" {
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Prefer GetUser; fallback to GetUserByID.
	var u models.User
	var err error
	if u, err = rt.db.GetUser(uid); err != nil {
		u, err = rt.db.GetUserByID(uid)
	}
	if err != nil {
		rt.sendError(w, http.StatusNotFound, "user not found")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"data": u,
	})
}

//
// ────────────────────────────────────────────────────────────────────────────────
//  Get user profile by id/name  (GET /users/:id 或 /users/profile?username=xxx)
// ────────────────────────────────────────────────────────────────────────────────
//

// getUserProfile loads another user's public profile.
func (rt *_router) getUserProfile(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	_ reqcontext.RequestContext,
) {
	// Path params first
	userID := strings.TrimSpace(ps.ByName("id"))
	if userID == "" {
		userID = strings.TrimSpace(ps.ByName("userId"))
	}
	// Query fallbacks
	if userID == "" {
		userID = strings.TrimSpace(r.URL.Query().Get("id"))
	}
	username := strings.TrimSpace(r.URL.Query().Get("username"))
	if username == "" {
		username = strings.TrimSpace(r.URL.Query().Get("name"))
	}

	// If only username provided, resolve to ID.
	if userID == "" && username != "" {
		id, err := rt.db.GetUserIDFromIdentifier(username)
		if err == nil && id != "" {
			userID = id
		}
	}

	if userID == "" {
		rt.sendError(w, http.StatusBadRequest, "missing user id or username")
		return
	}

	u, err := rt.db.GetUserByID(userID)
	if err != nil {
		rt.sendError(w, http.StatusNotFound, "user not found")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"data": u,
	})
}

// getUser is a compatibility alias that delegates to getUserProfile.
func (rt *_router) getUser(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	rt.getUserProfile(w, r, ps, ctx)
}

//
// ────────────────────────────────────────────────────────────────────────────────
//  Update username  (PUT /me/username 或 /users/me/set-username)
// ────────────────────────────────────────────────────────────────────────────────
//

// setUserUsername changes the current user's username (handle).
func (rt *_router) setUserUsername(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	uid := strings.TrimSpace(ctx.UserID)
	if uid == "" {
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body setUsernameBody
	if err := readJSON(r, &body); err != nil {
		rt.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	username := strings.TrimSpace(body.Username)
	if username == "" {
		username = strings.TrimSpace(body.Name) // legacy key
	}
	if username == "" {
		rt.sendError(w, http.StatusBadRequest, "username is required")
		return
	}

	if err := rt.db.UpdateUserName(uid, username); err != nil {
		rt.sendError(w, http.StatusBadRequest, "failed to update username")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    http.StatusOK,
		"message": "username updated",
		"data":    map[string]string{"username": username},
	})
}

// updateUserName is a compatibility alias to setUserUsername.
func (rt *_router) updateUserName(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	rt.setUserUsername(w, r, ps, ctx)
}

// setMyUserName is another compatibility alias used by api-handler.go.
func (rt *_router) setMyUserName(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	rt.setUserUsername(w, r, ps, ctx)
}

//
// ────────────────────────────────────────────────────────────────────────────────
//  Update avatar/photo  (PUT /me/photo 或 /users/me/set-photo)
//  支持两种模式：?preset=avatar7   或  multipart/form-data 字段 "upload"
// ────────────────────────────────────────────────────────────────────────────────
//

func (rt *_router) setUserPhoto(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	uid := strings.TrimSpace(ctx.UserID)
	if uid == "" {
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// --- 1) PRESET MODE via query param ---
	if preset := strings.TrimSpace(r.URL.Query().Get("preset")); preset != "" {
		// We only accept names like avatar1..avatarN.jpg kept under /uploads/photos/
		if !strings.HasPrefix(strings.ToLower(preset), "avatar") {
			rt.sendError(w, http.StatusBadRequest, "invalid preset name")
			return
		}
		derived := preset + ".jpg"
		publicURL := rt.publicURL(filepath.ToSlash(filepath.Join("uploads", "photos", derived)))

		if err := rt.db.UpdateUserPhoto(uid, publicURL); err != nil {
			rt.sendError(w, http.StatusInternalServerError, "failed to update user photo")
			return
		}
		_ = writeJSON(w, http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"message": "user photo updated",
			"data": map[string]interface{}{
				"file": map[string]interface{}{
					"filename": derived,
					"url":      publicURL,
				},
			},
		})
		return
	}

	// --- 2) UPLOAD MODE ---
	const maxUploadSizeBytes = 10 << 20 // 10 MiB
	if err := r.ParseMultipartForm(maxUploadSizeBytes); err != nil {
		rt.sendError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, header, err := r.FormFile("upload")
	if err != nil {
		rt.sendError(w, http.StatusBadRequest, "missing file field 'upload'")
		return
	}
	defer file.Close()

	origName := strings.TrimSpace(header.Filename)
	if origName == "" {
		rt.sendError(w, http.StatusBadRequest, "invalid filename")
		return
	}
	if !allowedExt(origName) {
		rt.sendError(w, http.StatusBadRequest, "only JPEG/PNG are allowed")
		return
	}
	if ctype, err := detectContentType(file); err == nil {
		if !(strings.HasPrefix(ctype, "image/jpeg") || strings.HasPrefix(ctype, "image/png")) {
			rt.sendError(w, http.StatusBadRequest, "invalid content type, only JPEG/PNG are allowed")
			return
		}
	}

	// Prepare destination under /uploads/users
	baseDir := filepath.Join("uploads", "users")
	if err := ensureDir(baseDir); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to prepare upload dir")
		return
	}

	ext := strings.ToLower(filepath.Ext(origName))
	safeUID := strings.NewReplacer("/", "_", "\\", "_", ":", "_").Replace(uid)
	newName := safeUID + "_" + time.Now().UTC().Format("20060102T150405Z") + ext
	dstPath := filepath.Join(baseDir, newName)

	dst, err := os.Create(dstPath)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to create destination file")
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, io.LimitReader(file, maxUploadSizeBytes+1))
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to store file")
		return
	}
	if written > maxUploadSizeBytes {
		_ = os.Remove(dstPath)
		rt.sendError(w, http.StatusBadRequest, "file too large (max 10MB)")
		return
	}

	publicPath := rt.publicURL(filepath.ToSlash(dstPath))
	if err := rt.db.UpdateUserPhoto(uid, publicPath); err != nil {
		_ = os.Remove(dstPath)
		rt.sendError(w, http.StatusInternalServerError, "failed to update user photo")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    http.StatusOK,
		"message": "user photo updated",
		"data": map[string]interface{}{
			"file": map[string]interface{}{
				"filename": origName,
				"size":     written,
				"url":      publicPath,
			},
		},
	})
}

// updateUserPhoto is a compatibility alias to setUserPhoto.
func (rt *_router) updateUserPhoto(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	rt.setUserPhoto(w, r, ps, ctx)
}

// setMyPhoto is another compatibility alias used by api-handler.go.
func (rt *_router) setMyPhoto(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	rt.setUserPhoto(w, r, ps, ctx)
}

//
// ────────────────────────────────────────────────────────────────────────────────
//  Search users  (GET /users/search?q=xxx)  —— 排除自己
// ────────────────────────────────────────────────────────────────────────────────
//

func (rt *_router) searchUsers(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	uid := strings.TrimSpace(ctx.UserID)
	if uid == "" {
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	// Also accept legacy key "query"
	if q == "" {
		q = strings.TrimSpace(r.URL.Query().Get("query"))
	}

	items, err := rt.db.SearchUsers(r.Context(), uid, q)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to search users")
		return
	}
	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"data": items,
	})
}
