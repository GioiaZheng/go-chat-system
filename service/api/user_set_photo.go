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

// setMyPhoto handles PUT /users/set_photo
// Modes:
//
//  1. Preset shortcut (no upload):
//     PUT /users/set_photo?preset=avatar6
//     -> avatarUrl = "/uploads/photos/avatar6.jpg"
//
//  2. Multipart upload (field: "upload"):
//     Content-Type: multipart/form-data
//     Body: upload=<file>
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code":401,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// --- Prefer preset if provided (works for any Content-Type) ---
	if preset := strings.TrimSpace(r.URL.Query().Get("preset")); preset != "" {
		// Accept "avatarN" (e.g., avatar1, avatar10); extension is assumed to be .jpg
		if !strings.HasPrefix(preset, "avatar") {
			http.Error(w, `{"code":400,"message":"preset must be like avatar1, avatar2, ... "}`, http.StatusBadRequest)
			return
		}
		url := fmt.Sprintf("/uploads/photos/%s.jpg", preset)
		if err := rt.db.UpdateUserPhoto(userID, url); err != nil {
			http.Error(w, `{"code":500,"message":"Failed to update photo in database"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    200,
			"message": "Photo updated successfully (preset)",
			"data": map[string]interface{}{
				"url": url,
			},
		})
		return
	}

	// --- Only parse multipart if Content-Type indicates so ---
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

		// Save file
		filename := fmt.Sprintf("%s_%d%s", userID, time.Now().UnixNano(), ext)
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
		if err := rt.db.UpdateUserPhoto(userID, url); err != nil {
			http.Error(w, `{"code":500,"message":"Failed to update photo in database"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    200,
			"message": "Photo updated successfully",
			"data": map[string]interface{}{
				"file": map[string]interface{}{
					"filename": handler.Filename,
					"size":     handler.Size,
					"url":      url,
				},
			},
		}); err != nil {
			rt.baseLogger.WithError(err).Error("failed to encode set photo response")
		}
		return
	}

	// Neither preset nor multipart upload provided
	http.Error(w, `{"code":400,"message":"Provide ?preset=avatarN or multipart field 'upload'"}`, http.StatusBadRequest)
}
