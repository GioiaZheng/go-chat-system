package api

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// NOTE ON RESPONSE SHAPE (matches FileUploadEnvelope in api.yaml):
// {
//   "code": 200,
//   "message": "User photo updated successfully",
//   "data": {
//     "file": {
//       "filename": "<original-or-derived>",
//       "url": "/uploads/photos/<stored-file>"
//       // "size": <bytes>   // optional in schema
//     }
//   }
// }
//
// Modes supported:
//   1) Preset mode:  PUT /users/set_photo?preset=avatar7
//      -> uses /uploads/photos/avatar7.jpg
//   2) Upload mode:  multipart/form-data with field "upload"
//
// IMPORTANT: Regardless of the mode, we ALWAYS return { data: { file: {...} } }.

const (
	maxUploadSizeBytes = 10 << 20 // 10 MiB (as in the OpenAPI schema)
)

// ---------- helpers ----------

func allowedExt(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png":
		return true
	default:
		return false
	}
}

func detectContentType(f multipart.File) (string, error) {
	const sniffLen = 512
	buf := make([]byte, sniffLen)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	return http.DetectContentType(buf[:n]), nil
}

func (rt *_router) publicURL(rel string) string {
	if !strings.HasPrefix(rel, "/") {
		return "/" + rel
	}
	return rel
}

func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

// ---------- handler ----------

// setMyPhoto handles: PUT /api/v1/users/set_photo
// - Preset: ?preset=avatar7 -> /uploads/photos/avatar7.jpg
// - Upload: multipart/form-data field "upload" (JPEG/PNG, <=10MB)
func (rt *_router) setMyPhoto(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	userID := strings.TrimSpace(ctx.UserID)
	if userID == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 1) PRESET MODE (uses /uploads/photos/avatarX.jpg)
	if preset := strings.TrimSpace(r.URL.Query().Get("preset")); preset != "" {
		// Defensive check: ensure the preset looks like "avatar<number>"
		if !strings.HasPrefix(strings.ToLower(preset), "avatar") {
			rt.sendError(w, http.StatusBadRequest, "invalid preset name")
			return
		}

		derivedFilename := preset + ".jpg"
		presetURL := rt.publicURL(filepath.ToSlash(filepath.Join("uploads", "photos", derivedFilename)))

		// Persist via DB: UpdateUserPhoto(userID, url)
		if err := rt.db.UpdateUserPhoto(userID, presetURL); err != nil {
			ctx.Logger.WithError(err).Error("failed to set preset user photo")
			rt.sendError(w, http.StatusInternalServerError, "Failed to update photo")
			return
		}

		resp := map[string]interface{}{
			"code":    http.StatusOK,
			"message": "User photo updated successfully",
			"data": map[string]interface{}{
				"file": map[string]interface{}{
					"filename": derivedFilename,
					"url":      presetURL,
				},
			},
		}
		if err := writeJSON(w, http.StatusOK, resp); err != nil {
			ctx.Logger.WithError(err).Error("failed to encode setMyPhoto (preset) response")
		}
		return
	}

	// 2) UPLOAD MODE (multipart/form-data, field name: "upload")
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

	origName := header.Filename
	if strings.TrimSpace(origName) == "" {
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

	baseDir := filepath.Join("uploads", "photos")
	if err := ensureDir(baseDir); err != nil {
		ctx.Logger.WithError(err).Error("failed to create uploads dir")
		rt.sendError(w, http.StatusInternalServerError, "Failed to store file")
		return
	}

	ext := strings.ToLower(filepath.Ext(origName))
	safeUser := strings.NewReplacer("/", "_", "\\", "_", ":", "_").Replace(userID)
	newName := safeUser + "_" + time.Now().UTC().Format("20060102T150405Z") + ext
	dstPath := filepath.Join(baseDir, newName)

	dst, err := os.Create(dstPath)
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to create destination file")
		rt.sendError(w, http.StatusInternalServerError, "Failed to store file")
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, io.LimitReader(file, maxUploadSizeBytes+1))
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to write uploaded file")
		rt.sendError(w, http.StatusInternalServerError, "Failed to store file")
		return
	}
	if written > maxUploadSizeBytes {
		_ = os.Remove(dstPath)
		rt.sendError(w, http.StatusBadRequest, "File too large (max 10MB)")
		return
	}

	publicPath := rt.publicURL(filepath.ToSlash(dstPath))

	// Persist the public URL via DB
	if err := rt.db.UpdateUserPhoto(userID, publicPath); err != nil {
		_ = os.Remove(dstPath) // cleanup on failure
		ctx.Logger.WithError(err).Error("failed to persist user photo url")
		rt.sendError(w, http.StatusInternalServerError, "Failed to update photo")
		return
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "User photo updated successfully",
		"data": map[string]interface{}{
			"file": map[string]interface{}{
				"filename": origName,   // original name returned to client
				"size":     written,    // optional
				"url":      publicPath, // final public URL
			},
		},
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode setMyPhoto response")
	}
}
