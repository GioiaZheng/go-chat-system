// cmd/webapi/register_webui.go
package main

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/webui"
)

func registerWebUI(router http.Handler) (http.Handler, error) {
	dist, err := webui.FS()
	if err != nil {
		return nil, err
	}

	// 静态文件 server（/assets/* /favicon.* /index.html 等）
	fileServer := http.FileServer(http.FS(dist))

	// SPA fallback：所有非静态命中路径，返回 dist/index.html
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 规范化路径
		up := path.Clean("/" + strings.TrimPrefix(r.URL.Path, "/"))

		// 明显的静态资源直接走 FileServer
		if up == "/index.html" || strings.HasPrefix(up, "/assets/") ||
			strings.HasSuffix(up, ".js") || strings.HasSuffix(up, ".css") ||
			strings.HasSuffix(up, ".ico") || strings.HasSuffix(up, ".png") ||
			strings.HasSuffix(up, ".svg") || strings.HasSuffix(up, ".txt") {
			r.URL.Path = up
			fileServer.ServeHTTP(w, r)
			return
		}

		// 尝试打开该文件（如果真实存在就直接返回）
		if f, err := dist.Open(strings.TrimPrefix(up, "/")); err == nil {
			_ = f.Close()
			r.URL.Path = up
			fileServer.ServeHTTP(w, r)
			return
		}

		// 回退到 index.html（单页应用前端路由）
		f, err := dist.Open("index.html")
		if err != nil {
			http.Error(w, "index.html not found", http.StatusInternalServerError)
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := io.Copy(w, f); err != nil && !errors.Is(err, io.EOF) {
			// 静默处理
		}
	}), nil
}
