package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Identifier string `json:"identifier"`
}

// doLogin 处理 POST /session: 用户登录或自动注册
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code": 400, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// 校验请求参数
	if req.Name == "" || req.Password == "" {
		http.Error(w, `{"code": 400, "message": "Name and Password are required"}`, http.StatusBadRequest)
		return
	}

	// 检查用户是否存在
	exists, err := rt.db.CheckUserExists(req.Name)
	if err != nil {
		rt.baseLogger.WithError(err).Error("database error checking user existence")
		http.Error(w, `{"code": 500, "message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	var identifier string
	if exists {
		// 用户存在，尝试登录
		identifier, err = rt.db.GetUserByCredentials(req.Name, req.Password)
		if err != nil {
			rt.baseLogger.WithError(err).Error("error getting user credentials")
			http.Error(w, `{"code": 401, "message": "Invalid credentials"}`, http.StatusUnauthorized)
			return
		}
	} else {
		// 用户不存在，尝试注册
		identifier, err = rt.db.CreateUser(req.Name, req.Password)
		if err != nil {
			rt.baseLogger.WithError(err).Error("error creating user")
			http.Error(w, `{"code": 500, "message": "Internal server error"}`, http.StatusInternalServerError)
			return
		}
	}

	// 返回成功响应
	resp := LoginResponse{
		Identifier: identifier,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode response")
		http.Error(w, `{"code": 500, "message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
}
