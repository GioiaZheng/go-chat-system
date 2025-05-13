//go:build webui

package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/webui"
)

// registerWebUI serves embedded frontend assets if the webui tag is enabled.
func registerWebUI(hdl http.Handler) (http.Handler, error) {
	// Load the embedded frontend files from the "dist" directory
	distDirectory, err := fs.Sub(webui.Dist, "dist")
	if err != nil {
		return nil, fmt.Errorf("error embedding WebUI dist/ directory: %w", err)
	}

	// Serve the frontend files for requests starting with "/dashboard/"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.RequestURI, "/dashboard/") {
			http.StripPrefix("/dashboard/", http.FileServer(http.FS(distDirectory))).ServeHTTP(w, r)
			return
		}
		hdl.ServeHTTP(w, r)
	}), nil
}
