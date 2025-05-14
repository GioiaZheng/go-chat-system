//go:build !webui

package main

import (
	"net/http"
)

// registerWebUI is a stub for builds without the webui tag.
// This returns the given handler unchanged.
func registerWebUI(hdl http.Handler) (http.Handler, error) {
	return hdl, nil
}
