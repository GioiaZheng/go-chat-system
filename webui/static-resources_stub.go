//go:build !webui
// +build !webui

package webui

import "embed"

// distFS is empty when the embedded web UI is omitted at build time.
// Keeping this stub in place lets builds and tests run without bundling the frontend assets.
var distFS embed.FS
