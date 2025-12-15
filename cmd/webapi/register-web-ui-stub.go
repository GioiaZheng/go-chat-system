//go:build !webui

// register-web-ui-stub.go provides a no-op placeholder when the binary is
// compiled without the `webui` build tag, leaving routing untouched.
package main

import (
	"net/http"
)

// registerWebUI is a stub that simply returns the provided handler when the
// webui build tag is not enabled, allowing the API server to run without
// embedded frontend assets.
func registerWebUI(hdl http.Handler) (http.Handler, error) {
	return hdl, nil
}
