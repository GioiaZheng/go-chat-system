// webui/static-resources.go
package webui

import (
	"embed"
	"io/fs"
)

// 将整个 dist 目录打包进二进制
//go:embed dist/* dist/assets/*
var distFS embed.FS

// FS 返回以 dist 为根的只读文件系统
func FS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
