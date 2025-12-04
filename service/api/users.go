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

// setUsernameBody 支持 { "name": "..." }（OpenAPI）以及兼容 { "username": "..." }。
type setUsernameBody struct {
	Username string `json:"username,omitempty"`
	Name     string `json:"name,omitempty"`
}

//
// Get my info  (GET /users/me)
//

// getUserInfo 返回当前登录用户资料（UserEnvelope）。
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

	// 兼容两种 DB 方法名
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

// getUserProfile 根据 path 的 :userId 读取公共资料；保留极简兼容查询键的降级方案。
func (rt *_router) getUserProfile(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	_ reqcontext.RequestContext,
) {
	// Path param
	userID := strings.TrimSpace(ps.ByName("userId"))
	if userID == "" {
		// 极简兼容：允许 :id 或 ?id= 之类（不在 OpenAPI 中，仅兜底）
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
			"user": u, // 与 OpenAPI: UserEnvelope 对齐
		},
	})
}

//
// Update username  (PUT /users/set_username)
//

// setUserUsername 修改当前用户的显示名（OpenAPI 请求字段是 name）。
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
	// OpenAPI 字段是 name；同时兼容 legacy 的 username
	newName := strings.TrimSpace(body.Name)
	if newName == "" {
		newName = strings.TrimSpace(body.Username)
	}
	if newName == "" {
		rt.sendError(w, http.StatusBadRequest, "name is required")
		return
	}

	if err := rt.db.UpdateUserName(uid, newName); err != nil {
		// 让 DB 层决定唯一性/正则等约束的错误；这里统一为 400
		rt.sendError(w, http.StatusBadRequest, "Failed to update user name")
		return
	}

	// OpenAPI: BaseSuccessResponse（只需要 code, message）
	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    http.StatusOK,
		"message": "user name updated successfully",
	})
}

// setMyUserName 供路由使用的名字（见 api-handler.go）
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
// 两种模式：?preset=avatar7  或 multipart/form-data 字段 "upload"
// 返回 FileUploadEnvelope：data.file { filename, uri, size? }
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

	// 1) 预设头像模式：?preset=avatarX
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
					"uri":      publicURI, // OpenAPI 字段叫 uri
				},
			},
		})
		return
	}

	// 2) 上传模式
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

	// 保存到 /uploads/users
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
				"uri":      publicURI, // OpenAPI 字段叫 uri
			},
		},
	})
}

// setMyPhoto 供路由使用的名字（见 api-handler.go）
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

	// q 可以为空：空就表示“搜索全部”
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		// 兼容一下 ?query= 这种旧写法
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
			"items": users, // 前端 api.js 里 unwrap + searchUsers() 正好兼容这个结构
		},
	})
}
