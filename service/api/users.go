package api

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

//
// Helpers (request bodies)
//

// setUsernameBody accepts both the OpenAPI field {"name":"..."} and the legacy {"username":"..."}.
type setUsernameBody struct {
	Username string `json:"username,omitempty"`
	Name     string `json:"name,omitempty"`
}

//
// Get my info  (GET /users/me)
//

// getUserInfo returns the profile of the authenticated user (UserEnvelope).
func (rt *_router) getUserInfo(
	w http.ResponseWriter,
	_ *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	uid := strings.TrimSpace(ctx.UserID)
	if uid == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

        // Support either DB method name
	u, err := rt.db.GetUser(uid)
	if err != nil {
		u, err = rt.db.GetUserByID(uid)
	}
	if err != nil {
		rt.sendError(w, http.StatusNotFound, "User not found")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    http.StatusOK,
		"message": "User information retrieved",
		"data":    u,
	})
}

//
// Get user profile by id  (GET /users/profile/{userId})
//

// getUserProfile reads public profile data by :userId and keeps a minimal query-key fallback.
func (rt *_router) getUserProfile(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	_ reqcontext.RequestContext,
) {
	// Path param
	userID := strings.TrimSpace(ps.ByName("userId"))
	if userID == "" {
                // Minimal compatibility: also allow :id or ?id= style fallbacks (non-OpenAPI)
		userID = strings.TrimSpace(ps.ByName("id"))
		if userID == "" {
			userID = strings.TrimSpace(r.URL.Query().Get("id"))
		}
	}

	if userID == "" {
		rt.sendError(w, http.StatusBadRequest, "Missing user id")
		return
	}

	u, err := rt.db.GetUserByID(userID)
	if err != nil {
		rt.sendError(w, http.StatusNotFound, "User not found")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    http.StatusOK,
		"message": "User information retrieved",
		"data": map[string]interface{}{
                        "user": u, // Aligns with the OpenAPI UserEnvelope
		},
	})
}

//
// Update username  (PUT /users/set_username)
//

// setUserUsername updates the display name (OpenAPI request field: name).
func (rt *_router) setUserUsername(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	uid := strings.TrimSpace(ctx.UserID)
	if uid == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var body setUsernameBody
	if err := readJSON(r, &body); err != nil {
		rt.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
        // Prefer OpenAPI field name; fall back to legacy username
	newName := strings.TrimSpace(body.Name)
	if newName == "" {
		newName = strings.TrimSpace(body.Username)
	}
	if newName == "" {
		rt.sendError(w, http.StatusBadRequest, "name is required")
		return
	}

	if err := rt.db.UpdateUserName(uid, newName); err != nil {
                // Let the DB enforce uniqueness/regex constraints; surface them as 400
		rt.sendError(w, http.StatusBadRequest, "Failed to update user name")
		return
	}

        // OpenAPI: BaseSuccessResponse (code and message only)
	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    http.StatusOK,
		"message": "user name updated successfully",
	})
}

// setMyUserName alias used by the router (see api-handler.go)
func (rt *_router) setMyUserName(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	rt.setUserUsername(w, r, ps, ctx)
}

//
// Update avatar/photo  (PUT /users/set_photo)
// Two modes: ?preset=avatar7 or multipart/form-data field "upload"
// Returns FileUploadEnvelope: data.file { filename, uri, size? }
//

func (rt *_router) setUserPhoto(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	uid := strings.TrimSpace(ctx.UserID)
	if uid == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

        // 1) Preset avatar mode: ?preset=avatarX
	if preset := strings.TrimSpace(r.URL.Query().Get("preset")); preset != "" {
		if !strings.HasPrefix(strings.ToLower(preset), "avatar") {
			rt.sendError(w, http.StatusBadRequest, "Invalid preset name")
			return
		}
		derived := preset + ".jpg"
		publicURI := rt.publicURL(filepath.ToSlash(filepath.Join("uploads", "photos", derived)))

		if err := rt.db.UpdateUserPhoto(uid, publicURI); err != nil {
			rt.sendError(w, http.StatusInternalServerError, "Failed to update user photo")
			return
		}
		_ = writeJSON(w, http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"message": "user photo updated",
			"data": map[string]interface{}{
				"file": map[string]interface{}{
					"filename": derived,
                                        "uri":      publicURI, // OpenAPI field is named uri
				},
			},
		})
		return
	}

        // 2) Upload mode
	const maxUploadSizeBytes = 10 << 20 // 10 MiB
	if err := r.ParseMultipartForm(maxUploadSizeBytes); err != nil {
		rt.sendError(w, http.StatusBadRequest, "Invalid multipart form")
		return
	}
	file, header, err := r.FormFile("upload")
	if err != nil {
		rt.sendError(w, http.StatusBadRequest, "Missing file field 'upload'")
		return
	}
	defer file.Close()

	origName := strings.TrimSpace(header.Filename)
	if origName == "" {
		rt.sendError(w, http.StatusBadRequest, "Invalid filename")
		return
	}
	if !allowedExt(origName) {
		rt.sendError(w, http.StatusBadRequest, "Only JPEG/PNG are allowed")
		return
	}
	if ctype, err := detectContentType(file); err == nil {
		if !(strings.HasPrefix(ctype, "image/jpeg") || strings.HasPrefix(ctype, "image/png")) {
			rt.sendError(w, http.StatusBadRequest, "Invalid content type, only JPEG/PNG are allowed")
			return
		}
	}

        // Save the file to /uploads/users
	baseDir := filepath.Join("uploads", "users")
	if err := ensureDir(baseDir); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "Failed to prepare upload dir")
		return
	}

	ext := strings.ToLower(filepath.Ext(origName))
	safeUID := strings.NewReplacer("/", "_", "\\", "_", ":", "_").Replace(uid)
	newName := safeUID + "_" + time.Now().UTC().Format("20060102T150405Z") + ext
	dstPath := filepath.Join(baseDir, newName)

	dst, err := os.Create(dstPath)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "Failed to create destination file")
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, io.LimitReader(file, maxUploadSizeBytes+1))
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "Failed to store file")
		return
	}
	if written > maxUploadSizeBytes {
		_ = os.Remove(dstPath)
		rt.sendError(w, http.StatusBadRequest, "File too large (max 10MB)")
		return
	}

	publicURI := rt.publicURL(filepath.ToSlash(dstPath))
	if err := rt.db.UpdateUserPhoto(uid, publicURI); err != nil {
		_ = os.Remove(dstPath)
		rt.sendError(w, http.StatusInternalServerError, "Failed to update user photo")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    http.StatusOK,
		"message": "user photo updated",
		"data": map[string]interface{}{
			"file": map[string]interface{}{
				"filename": origName,
				"size":     written,
                                "uri":      publicURI, // OpenAPI field is named uri
			},
		},
	})
}

// setMyPhoto alias used by the router (see api-handler.go)
func (rt *_router) setMyPhoto(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	rt.setUserPhoto(w, r, ps, ctx)
}

// Search users  (GET /users/search?q=xxx)  — exclude myself
func (rt *_router) searchUsers(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	uid := strings.TrimSpace(ctx.UserID)
	if uid == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

        // q may be empty, which means "search all"
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
                // Also accept the older ?query= variant
		q = strings.TrimSpace(r.URL.Query().Get("query"))
	}

	users, err := rt.db.SearchUsers(r.Context(), uid, q)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "Failed to search users")
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Matching users found",
		"data": map[string]interface{}{
                        "items": users, // Matches the structure expected by api.js unwrap + searchUsers()
		},
	})
}
