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

// publicURL builds a public URL (or path) from a relative file path.
// If the input is already absolute (http/https or starts with "/"), return as-is.
// Otherwise, prefix with "/" so that a static file server can serve it.
//
// NOTE: Adjust this if you have a configured external base URL.
func (rt *_router) publicURL(rel string) string {
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	if rel == "" {
		return rel
	}
	if strings.HasPrefix(rel, "http://") || strings.HasPrefix(rel, "https://") {
		return rel
	}
	if strings.HasPrefix(rel, "/") {
		return rel
	}
	return "/" + rel
}

// allowedExt returns true if the filename has a JPEG/PNG extension.
func allowedExt(filename string) bool {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(filename)))
	switch ext {
	case ".jpg", ".jpeg", ".png":
		return true
	default:
		return false
	}
}

// detectContentType reads up to the first 512 bytes and uses http.DetectContentType.
// It restores the file offset afterwards so callers can read from the beginning again.
//
// multipart.File in Go provides Read and Seek; we rely on Seek to reset position.
// If reading fails with a non-EOF error, it returns the error.
func detectContentType(f multipart.File) (string, error) {
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) && n == 0 {
		return "", err
	}
	// Rewind to the beginning for subsequent readers.
	_, _ = f.Seek(0, io.SeekStart)

	if n == 0 {
		// No bytes read; return generic type (we won't strictly validate in callers if empty).
		return "application/octet-stream", nil
	}
	return http.DetectContentType(buf[:n]), nil
}

// ensureDir creates the directory path if it does not exist.
func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}
