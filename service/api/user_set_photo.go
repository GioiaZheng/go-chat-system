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

// setMyPhoto handles PUT /users/set_photo (multipart/form-data)
// Expected field name: "upload".
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code":401,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
		http.Error(w, `{"code":400,"message":"Failed to parse form data"}`, http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("upload")
	if err != nil {
		http.Error(w, `{"code":400,"message":"Field 'upload' is required"}`, http.StatusBadRequest)
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

	url := "/" + path // store the relative URL
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
}
