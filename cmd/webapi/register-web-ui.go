//go:build webui

package main

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/webui"
)

// 当使用 -tags webui 构建时，将前端的嵌入静态资源挂到 "/"
// 其余路径都交给 API 处理器
func registerWebUI(api http.Handler) (http.Handler, error) {
	mux := http.NewServeMux()

	// 前端
	mux.Handle("/", webui.Handler())

	// API（把常见前缀路由给到原有 API handler）
	// 如果你的 API 全部在根路径，也可以把默认注册顺序换一下：
	// 先注册具体前缀到 api，再用 "/" 给前端。
	mux.Handle("/session", api)
	mux.Handle("/register", api)
	mux.Handle("/liveness", api)
	mux.Handle("/users/", api)
	mux.Handle("/groups/", api)
	mux.Handle("/messages/", api)
	mux.Handle("/conversations", api)
	mux.Handle("/conversations/", api)
	mux.Handle("/uploads/", api)

	return mux, nil
}
