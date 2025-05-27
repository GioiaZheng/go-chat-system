// auth_do_login.go
package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/julienschmidt/httprouter"
)

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Identifier string `json:"identifier"`
}

func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 解析请求体
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.baseLogger.WithError(err).Error("failed to decode login request")
		http.Error(w, `{"code": 400, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// 校验参数
	if req.Name == "" || req.Password == "" {
		rt.baseLogger.Error("missing name or password")
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

	var user models.User
	if exists {
		// 用户存在，验证密码
		userID, err := rt.db.GetUserByCredentials(req.Name, req.Password)
		if err != nil {
			rt.baseLogger.WithError(err).Error("invalid credentials")
			http.Error(w, `{"code": 401, "message": "Invalid credentials"}`, http.StatusUnauthorized)
			return
		}
		user, err = rt.db.GetUserByID(userID)
		if err != nil {
			rt.baseLogger.WithError(err).Error("failed to get user details")
			http.Error(w, `{"code": 500, "message": "Internal server error"}`, http.StatusInternalServerError)
			return
		}
	} else {
		// 用户不存在，自动注册
		user = models.User{
			Username: req.Name,
			Name:     req.Name,
		}
		user, err = rt.db.CreateUser(user, req.Password)
		if err != nil {
			rt.baseLogger.WithError(err).Error("failed to create user")
			http.Error(w, `{"code": 500, "message": "Internal server error"}`, http.StatusInternalServerError)
			return
		}
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(LoginResponse{
		Identifier: user.ID,
	})

	rt.baseLogger.Infof("User %s logged in", user.ID)
}
