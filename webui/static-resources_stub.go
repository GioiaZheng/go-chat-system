//go:build !webui
// +build !webui

package webui

import "embed"

// distFS is empty when the embedded web UI is not compiled in.
// This keeps builds and tests working without the frontend assets.
var distFS embed.FS