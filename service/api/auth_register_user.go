package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/api/reqcontext"
	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/julienschmidt/httprouter"
)

// doRegister handles POST /register: 用户注册
func (rt *_router) doRegister(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// 解析请求体
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code": 400, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// 校验请求参数
	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, `{"code": 400, "message": "Username, Email, and Password are required"}`, http.StatusBadRequest)
		return
	}

	// 创建用户
	newUser := models.User{
		Username: req.Username,
		Email:    req.Email,
	}

	// 调用 CreateUser
	createdUser, err := rt.db.CreateUser(newUser, req.Password)
	if err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to register user"}`, http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	resp := map[string]interface{}{
		"code":    201,
		"message": "User registered successfully",
		"data":    createdUser,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to encode response"}`, http.StatusInternalServerError)
	}
}
