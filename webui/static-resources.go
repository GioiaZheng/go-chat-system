//go:build webui
// +build webui

package webui

import "embed"

// dist contains the compiled frontend assets.
//
//go:embed dist/*
var distFS embed.FS
