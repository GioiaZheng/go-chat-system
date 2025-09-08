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

// setMyPhoto handles PUT /users/set_photo
// Two modes supported:
// 1) Preset (?preset=avatar6 -> "/uploads/photos/avatar6.jpg")
// 2) Multipart upload (field name: "upload")
//
// English notes:
// - Do NOT use http.Error / fmt.* / json.Encoder here; always use rt.sendError / writeJSON.
// - Log internal errors with ctx.Logger.WithError(err).Error("...").
func (rt *_router) setMyPhoto(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if strings.TrimSpace(ctx.UserID) == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// --- Preset mode (?preset=avatar6) ---
	if raw := strings.TrimSpace(r.URL.Query().Get("preset")); raw != "" {
		if !(strings.HasPrefix(raw, "avatar") || strings.HasPrefix(raw, "user")) {
			rt.sendError(w, http.StatusBadRequest, "preset must be like avatar6 or user6")
			return
		}
		url := "/uploads/photos/" + raw + ".jpg"
		if err := rt.db.UpdateUserPhoto(ctx.UserID, url); err != nil {
			ctx.Logger.WithError(err).Error("failed to update user photo (preset)")
			rt.sendError(w, http.StatusInternalServerError, "Failed to update user photo")
			return
		}
		resp := map[string]interface{}{
			"code":    http.StatusOK,
			"message": "User photo updated successfully (preset)",
			"data": map[string]interface{}{
				"url": url,
			},
		}
		if err := writeJSON(w, http.StatusOK, resp); err != nil {
			ctx.Logger.WithError(err).Error("failed to encode set photo preset response")
		}
		return
	}

	// --- Multipart upload mode ---
	ct := strings.ToLower(r.Header.Get("Content-Type"))
	if !strings.HasPrefix(ct, "multipart/form-data") {
		rt.sendError(w, http.StatusBadRequest, "Provide ?preset=avatarN or multipart field 'upload'")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB
		rt.sendError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}
	file, handler, err := r.FormFile("upload")
	if err != nil {
		rt.sendError(w, http.StatusBadRequest, "Field 'upload' is required (or use ?preset=avatarN)")
		return
	}
	defer file.Close()

	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	ext := strings.ToLower(filepath.Ext(handler.Filename))
	if !allowed[ext] {
		rt.sendError(w, http.StatusBadRequest, "Unsupported file type. Allowed: jpg,jpeg,png,gif")
		return
	}
	if handler.Size > 10*1024*1024 {
		rt.sendError(w, http.StatusBadRequest, "File size exceeds 10MB")
		return
	}

	filename := "user_" + ctx.UserID + "_" + time.Now().Format("20060102150405") + ext
	path := filepath.Join("uploads", "photos", filename)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		ctx.Logger.WithError(err).Error("failed to create directory for user photo")
		rt.sendError(w, http.StatusInternalServerError, "Failed to create directory")
		return
	}

	out, err := os.Create(path)
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to create file for user photo")
		rt.sendError(w, http.StatusInternalServerError, "Failed to save photo")
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		ctx.Logger.WithError(err).Error("failed to copy uploaded user photo")
		rt.sendError(w, http.StatusInternalServerError, "Failed to save photo")
		return
	}

	url := "/" + path
	if err := rt.db.UpdateUserPhoto(ctx.UserID, url); err != nil {
		ctx.Logger.WithError(err).Error("failed to update DB for uploaded user photo")
		rt.sendError(w, http.StatusInternalServerError, "Failed to update user photo")
		return
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "User photo updated successfully",
		"data": map[string]interface{}{
			"file": map[string]interface{}{
				"filename": handler.Filename,
				"size":     handler.Size,
				"url":      url,
			},
		},
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode set photo response")
	}
}
