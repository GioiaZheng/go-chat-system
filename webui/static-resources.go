package webui

import "embed"

// dist 包含打包后的前端资源。
//
//go:embed dist/*
var distFS embed.FS
