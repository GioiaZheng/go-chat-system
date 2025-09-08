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

// setGroupPhoto handles PUT /groups/:id/photo
// Modes supported:
// 1) Preset shortcut (?preset=avatar6 -> "/uploads/photos/avatar6.jpg")
// 2) Multipart upload (field: "upload")
//
// English notes:
// - All error/success responses now use rt.sendError / writeJSON for consistency.
// - No fmt.Print; logger used for errors.
// - Directory and file operations safely handled.
func (rt *_router) setGroupPhoto(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// --- Auth check ---
	if ctx.UserID == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// --- Group ID from path ---
	groupID := ps.ByName("id")
	if strings.TrimSpace(groupID) == "" {
		rt.sendError(w, http.StatusBadRequest, "Missing group id")
		return
	}

	// --- Preset mode (?preset=avatar6 / group6) ---
	if raw := strings.TrimSpace(r.URL.Query().Get("preset")); raw != "" {
		if strings.HasPrefix(raw, "avatar") || strings.HasPrefix(raw, "group") {
			url := "/uploads/photos/" + raw + ".jpg"
			if err := rt.db.UpdateGroupPhoto(groupID, url); err != nil {
				ctx.Logger.WithError(err).Error("failed to update group photo (preset)")
				rt.sendError(w, http.StatusInternalServerError, "Failed to update group photo in database")
				return
			}
			resp := map[string]interface{}{
				"code":    http.StatusOK,
				"message": "Group photo updated successfully (preset)",
				"data": map[string]interface{}{
					"url": url,
				},
			}
			if err := writeJSON(w, http.StatusOK, resp); err != nil {
				ctx.Logger.WithError(err).Error("failed to encode preset group photo response")
			}
			return
		}
		rt.sendError(w, http.StatusBadRequest, "preset must be like avatar6 or group6")
		return
	}

	// --- Multipart upload mode ---
	ct := strings.ToLower(r.Header.Get("Content-Type"))
	if strings.HasPrefix(ct, "multipart/form-data") {
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

		// Validate extension and size
		allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
		ext := strings.ToLower(filepath.Ext(handler.Filename))
		if !allowed[ext] {
			rt.sendError(w, http.StatusBadRequest, "Unsupported file type. Allowed: jpg,jpeg,png,gif")
			return
		}
		if handler.Size > 10*1024*1024 {
			rt.sendError(w, http.StatusBadRequest, "Photo size exceeds 10MB")
			return
		}

		// Save file under uploads/photos
		filename := "group_" + groupID + "_" + time.Now().Format("20060102150405") + ext
		path := filepath.Join("uploads", "photos", filename)

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			ctx.Logger.WithError(err).Error("failed to create directory for group photo")
			rt.sendError(w, http.StatusInternalServerError, "Failed to create directory")
			return
		}

		out, err := os.Create(path)
		if err != nil {
			ctx.Logger.WithError(err).Error("failed to create file for group photo")
			rt.sendError(w, http.StatusInternalServerError, "Failed to save photo")
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, file); err != nil {
			ctx.Logger.WithError(err).Error("failed to copy uploaded group photo")
			rt.sendError(w, http.StatusInternalServerError, "Failed to save photo")
			return
		}

		url := "/" + path
		if err := rt.db.UpdateGroupPhoto(groupID, url); err != nil {
			ctx.Logger.WithError(err).Error("failed to update DB for uploaded group photo")
			rt.sendError(w, http.StatusInternalServerError, "Failed to update group photo in database")
			return
		}

		resp := map[string]interface{}{
			"code":    http.StatusOK,
			"message": "Group photo updated successfully",
			"data": map[string]interface{}{
				"file": map[string]interface{}{
					"filename": handler.Filename,
					"size":     handler.Size,
					"url":      url,
				},
			},
		}
		if err := writeJSON(w, http.StatusOK, resp); err != nil {
			ctx.Logger.WithError(err).Error("failed to encode upload group photo response")
		}
		return
	}

	// --- If neither preset nor multipart upload ---
	rt.sendError(w, http.StatusBadRequest, "Provide ?preset=avatarN or multipart field 'upload'")
}
