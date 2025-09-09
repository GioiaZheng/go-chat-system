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

// setGroupPhoto handles: PUT /api/v1/groups/:id/photo
// Two modes:
//   1) Preset: ?preset=avatar7  -> /uploads/photos/avatar7.jpg
//   2) Upload: multipart/form-data field "upload"
// Response strictly follows FileUploadEnvelope.
func (rt *_router) setGroupPhoto(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	groupID := strings.TrimSpace(ps.ByName("id"))
	if groupID == "" {
		rt.sendError(w, http.StatusBadRequest, "Group ID is required")
		return
	}

	// ---------- 1) PRESET MODE ----------
	if preset := strings.TrimSpace(r.URL.Query().Get("preset")); preset != "" {
		if !strings.HasPrefix(strings.ToLower(preset), "avatar") {
			rt.sendError(w, http.StatusBadRequest, "invalid preset name")
			return
		}

		derivedFilename := preset + ".jpg"
		presetURL := rt.publicURL(filepath.ToSlash(filepath.Join("uploads", "photos", derivedFilename)))

		// DB call: UpdateGroupPhoto(groupID, url)
		if err := rt.db.UpdateGroupPhoto(groupID, presetURL); err != nil {
			ctx.Logger.WithError(err).Error("failed to set preset group photo")
			rt.sendError(w, http.StatusInternalServerError, "Failed to update group photo")
			return
		}

		resp := map[string]interface{}{
			"code":    http.StatusOK,
			"message": "Group photo updated successfully",
			"data": map[string]interface{}{
				"file": map[string]interface{}{
					"filename": derivedFilename,
					"url":      presetURL,
				},
			},
		}
		if err := writeJSON(w, http.StatusOK, resp); err != nil {
			ctx.Logger.WithError(err).Error("failed to encode setGroupPhoto (preset) response")
		}
		return
	}

	// ---------- 2) UPLOAD MODE ----------
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

	origName := header.Filename
	if strings.TrimSpace(origName) == "" {
		rt.sendError(w, http.StatusBadRequest, "Invalid filename")
		return
	}
	if !allowedExt(origName) {
		rt.sendError(w, http.StatusBadRequest, "Only JPEG/PNG are allowed")
		return
	}

	// Content-type sniff
	if ctype, err := detectContentType(file); err == nil {
		if !(strings.HasPrefix(ctype, "image/jpeg") || strings.HasPrefix(ctype, "image/png")) {
			rt.sendError(w, http.StatusBadRequest, "Invalid content type, only JPEG/PNG are allowed")
			return
		}
	}

	// Save under /uploads/groups
	baseDir := filepath.Join("uploads", "groups")
	if err := ensureDir(baseDir); err != nil {
		ctx.Logger.WithError(err).Error("failed to create uploads dir for groups")
		rt.sendError(w, http.StatusInternalServerError, "Failed to store file")
		return
	}

	ext := strings.ToLower(filepath.Ext(origName))
	safeGroup := strings.NewReplacer("/", "_", "\\", "_", ":", "_").Replace(groupID)
	newName := safeGroup + "_" + time.Now().UTC().Format("20060102T150405Z") + ext
	dstPath := filepath.Join(baseDir, newName)

	dst, err := os.Create(dstPath)
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to create destination file (group photo)")
		rt.sendError(w, http.StatusInternalServerError, "Failed to store file")
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, io.LimitReader(file, maxUploadSizeBytes+1))
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to write uploaded group photo")
		rt.sendError(w, http.StatusInternalServerError, "Failed to store file")
		return
	}
	if written > maxUploadSizeBytes {
		_ = os.Remove(dstPath)
		rt.sendError(w, http.StatusBadRequest, "File too large (max 10MB)")
		return
	}

	publicPath := rt.publicURL(filepath.ToSlash(dstPath))

	// Persist URL in DB
	if err := rt.db.UpdateGroupPhoto(groupID, publicPath); err != nil {
		_ = os.Remove(dstPath)
		ctx.Logger.WithError(err).Error("failed to persist group photo url")
		rt.sendError(w, http.StatusInternalServerError, "Failed to update group photo")
		return
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Group photo updated successfully",
		"data": map[string]interface{}{
			"file": map[string]interface{}{
				"filename": origName,
				"size":     written,
				"url":      publicPath,
			},
		},
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode setGroupPhoto response")
	}
}
