package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// setGroupPhoto handles PUT /groups/:id/set_photo
// Modes:
// 1) Preset shortcut (no upload):
//    PUT /groups/:id/set_photo?preset=avatar6    -> "/uploads/photos/avatar6.jpg"
//    也接受 "group6" 这种前缀，统一映射到 "/uploads/photos/group6.jpg"
// 2) Multipart upload (field: "upload")
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Auth is already checked by middleware that fills ctx.UserID (if your project does that).
	if ctx.UserID == "" {
		http.Error(w, `{"code":401,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Get group id from :id or :groupId (be tolerant)
	groupID := ps.ByName("id")
	if groupID == "" {
		groupID = ps.ByName("groupId")
	}
	if groupID == "" {
		// try query fallback
		groupID = strings.TrimSpace(r.URL.Query().Get("group_id"))
	}
	if groupID == "" {
		http.Error(w, `{"code":400,"message":"Missing group id"}`, http.StatusBadRequest)
		return
	}

	// --- Prefer preset if provided (works for any Content-Type) ---
	if raw := strings.TrimSpace(r.URL.Query().Get("preset")); raw != "" {
		preset := raw
		// Accept avatarN or groupN; normalize to the filename we actually have.
		if strings.HasPrefix(preset, "avatar") || strings.HasPrefix(preset, "group") {
			// Assume .jpg in uploads/photos
			url := fmt.Sprintf("/uploads/photos/%s.jpg", preset)
			if err := rt.db.UpdateGroupPhoto(groupID, url); err != nil {
				http.Error(w, `{"code":500,"message":"Failed to update group photo in database"}`, http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code":    200,
				"message": "Group photo updated successfully (preset)",
				"data": map[string]interface{}{
					"url": url,
				},
			})
			return
		}
		http.Error(w, `{"code":400,"message":"preset must be like avatar6 or group6"}`, http.StatusBadRequest)
		return
	}

	// --- Multipart upload only if Content-Type is multipart/form-data ---
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(strings.ToLower(ct), "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
			http.Error(w, `{"code":400,"message":"Failed to parse form data"}`, http.StatusBadRequest)
			return
		}
		file, handler, err := r.FormFile("upload")
		if err != nil {
			http.Error(w, `{"code":400,"message":"Field 'upload' is required (or use ?preset=avatarN)"}`, http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Basic validation
		allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
		ext := strings.ToLower(filepath.Ext(handler.Filename))
		if !allowed[ext] {
			http.Error(w, `{"code":400,"message":"Unsupported file type. Allowed: jpg,jpeg,png,gif"}`, http.StatusBadRequest)
			return
		}
		if handler.Size > 10*1024*1024 {
			http.Error(w, `{"code":400,"message":"Photo size exceeds 10MB"}`, http.StatusBadRequest)
			return
		}

		// Save to uploads/photos
		filename := fmt.Sprintf("group_%s_%d%s", groupID, time.Now().UnixNano(), ext)
		path := filepath.Join("uploads", "photos", filename)

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			http.Error(w, `{"code":500,"message":"Failed to create directory"}`, http.StatusInternalServerError)
			return
		}
		out, err := os.Create(path)
		if err != nil {
			http.Error(w, `{"code":500,"message":"Failed to save photo"}`, http.StatusInternalServerError)
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, file); err != nil {
			http.Error(w, `{"code":500,"message":"Failed to save photo"}`, http.StatusInternalServerError)
			return
		}

		url := "/" + path
		if err := rt.db.UpdateGroupPhoto(groupID, url); err != nil {
			http.Error(w, `{"code":500,"message":"Failed to update group photo in database"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    200,
			"message": "Group photo updated successfully",
			"data": map[string]interface{}{
				"file": map[string]interface{}{
					"filename": handler.Filename,
					"size":     handler.Size,
					"url":      url,
				},
			},
		})
		return
	}

	http.Error(w, `{"code":400,"message":"Provide ?preset=avatarN or multipart field 'upload'"}`, http.StatusBadRequest)
}
