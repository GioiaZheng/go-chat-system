package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// setMyPhoto 处理 PUT /users/set_photo
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// 解析 multipart/form-data
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, `{"code": 400, "message": "Failed to parse form data"}`, http.StatusBadRequest)
		return
	}

	// 获取文件
	file, handler, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, `{"code": 400, "message": "Photo is required"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 校验文件类型和大小
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}
	fileExt := strings.ToLower(filepath.Ext(handler.Filename))
	if !allowedExtensions[fileExt] {
		http.Error(w, `{"code": 400, "message": "Unsupported file type. Allowed types: jpg, jpeg, png, gif"}`, http.StatusBadRequest)
		return
	}

	if handler.Size > 10485760 {
		http.Error(w, `{"code": 400, "message": "Photo size exceeds 10MB"}`, http.StatusBadRequest)
		return
	}

	// 生成唯一文件名
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("%s_%d%s", userID, timestamp, fileExt)
	filePath := filepath.Join("uploads", "photos", filename)

	// 创建目录
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to create directory"}`, http.StatusInternalServerError)
		return
	}

	// 保存文件
	outFile, err := os.Create(filePath)
	if err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to save photo"}`, http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, file); err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to save photo"}`, http.StatusInternalServerError)
		return
	}

	// 更新用户头像路径
	if err := rt.db.UpdateUserPhoto(userID, "/"+filePath); err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to update photo in database"}`, http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	resp := map[string]interface{}{
		"code":    200,
		"message": "Photo updated successfully",
		"data": map[string]string{
			"photoUrl": "/" + filePath,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
