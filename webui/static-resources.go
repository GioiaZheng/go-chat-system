//go:build webui
// +build webui

package webui

import "embed"

// distFS exposes the compiled frontend assets when the web UI is embedded at build time.
//
//go:embed dist/*
var distFS embed.FS
