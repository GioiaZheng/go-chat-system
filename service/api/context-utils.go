package api

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError is a helper to write an error response as JSON
func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// 获取用户ID（从中间件注入的 context 中）
func getUserIDFromContext(r *http.Request) string {
	userID, ok := r.Context().Value("userId").(string)
	if !ok {
		return ""
	}
	return userID
}

// 通用 JSON 解析方法
func readJSON(r *http.Request, dst interface{}) error {
	return json.NewDecoder(r.Body).Decode(dst)
}
