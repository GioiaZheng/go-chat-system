package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// setMyPhoto handles PUT /users/set_photo (multipart/form-data)
// Expected field name: "upload"
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code":401,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
		http.Error(w, `{"code":400,"message":"Invalid multipart form"}`, http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("upload")
	if err != nil {
		http.Error(w, `{"code":400,"message":"Missing file field 'upload'"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save to local uploads folder
	if err := os.MkdirAll("uploads", 0o755); err != nil {
		http.Error(w, `{"code":500,"message":"Failed to prepare upload directory"}`, http.StatusInternalServerError)
		return
	}
	filename := fmt.Sprintf("%s_%s", userID, header.Filename)
	dstPath := filepath.Join("uploads", filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, `{"code":500,"message":"Failed to save file"}`, http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, `{"code":500,"message":"Failed to write file"}`, http.StatusInternalServerError)
		return
	}

	// Persist relative path in DB
	urlPath := "/" + dstPath
	if err := rt.db.UpdateUserPhoto(userID, urlPath); err != nil {
		http.Error(w, `{"code":500,"message":"Failed to update user photo"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Photo uploaded",
		"data": map[string]interface{}{
			"file": map[string]interface{}{
				"url": urlPath,
			},
		},
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode set photo response")
	}
}
