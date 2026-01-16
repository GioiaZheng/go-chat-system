// upload_utils.go centralizes helpers for sanitizing upload paths and
// validating basic file properties before persisting user-supplied content.
// Related files: service/api/messages.go, service/api/groups.go.
package api

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// publicURL builds a public-facing URL (or absolute path) from a relative file
// path.
//
// Behavior:
//   - If the input is already absolute (http/https or starts with "/"), it is
//     returned as-is.
//   - Otherwise the path is normalized, traversal is stripped, and a single
//     leading slash is added so a static server can serve it directly.
//
// NOTE: Adjust this if you deploy behind a CDN or have an external base URL.
func (rt *_router) publicURL(rel string) string {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return rel
	}

	// Protocol check (case-insensitive), keep as-is for absolute URLs.
	low := strings.ToLower(rel)
	if strings.HasPrefix(low, "http://") || strings.HasPrefix(low, "https://") {
		return rel
	}

	// Normalize path separators and clean to avoid ../ traversal.
	rel = filepath.ToSlash(filepath.Clean(rel))
	// Remove any leading "../" after cleaning (defense-in-depth).
	for strings.HasPrefix(rel, "../") {
		rel = strings.TrimPrefix(rel, "../")
	}
	// If it is already absolute (starts with "/"), keep it.
	if strings.HasPrefix(rel, "/") {
		return rel
	}
	return "/" + rel
}

// allowedExt performs a coarse extension whitelist for JPEG/PNG files. Callers
// should still verify Content-Type and size limits.
func allowedExt(filename string) bool {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(filename)))
	switch ext {
	case ".jpg", ".jpeg", ".png":
		return true
	default:
		return false
	}
}

// detectContentType inspects up to the first 512 bytes using http.DetectContentType.
// It then attempts a best-effort rewind so downstream readers can start from
// offset 0. If the file is not seekable, we avoid hard failures but callers
// should not assume the offset was restored.
func detectContentType(f multipart.File) (string, error) {
	buf := make([]byte, 512)
	n, readErr := f.Read(buf)
	if readErr != nil && !errors.Is(readErr, io.EOF) && n == 0 {
		return "", readErr
	}

	// Best effort rewind. multipart.File typically supports Seek, but guard anyway.
	type seeker interface {
		Seek(offset int64, whence int) (int64, error)
	}
	if sf, ok := f.(seeker); ok {
		_, _ = sf.Seek(0, io.SeekStart)
	}

	if n == 0 {
		// No bytes read → return generic type; caller may decide how strict to be.
		return "application/octet-stream", nil
	}
	return http.DetectContentType(buf[:n]), nil
}

// ensureDir creates the directory tree for uploads if it does not exist
// (idempotent).
func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}
