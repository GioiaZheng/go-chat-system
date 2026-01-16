//go:build webui
// +build webui

// static-resources.go embeds the built frontend assets for webui builds.
// Related files: webui/register_webui.go, webui/static-resources_stub.go.
package webui

import "embed"

// distFS exposes the compiled frontend assets when the web UI is embedded at build time.
//
//go:embed dist/*
var distFS embed.FS
