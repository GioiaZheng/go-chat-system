//go:build !webui
// +build !webui

// static-resources_stub.go provides the empty embed used when the webui build
// tag is not enabled.
// Related files: webui/static-resources.go, webui/register_webui.go.
package webui

import "embed"

// distFS is empty when the embedded web UI is omitted at build time.
// Keeping this stub in place lets builds and tests run without bundling the frontend assets.
var distFS embed.FS
