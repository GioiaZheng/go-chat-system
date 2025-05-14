package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/api/reqcontext"
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

// doLogin 处理 POST /session: 用户登录或自动注册
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Decode the login request
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code": 400, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate request fields
	if req.Name == "" || req.Password == "" {
		http.Error(w, `{"code": 400, "message": "Name and Password are required"}`, http.StatusBadRequest)
		return
	}

	// Check if the user exists
	exists, err := rt.db.CheckUserExists(req.Name)
	if err != nil {
		ctx.Logger.WithError(err).Error("database error checking user existence")
		http.Error(w, `{"code": 500, "message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	var user models.User
	if exists {
		// User exists, try to login
		userID, err := rt.db.GetUserByCredentials(req.Name, req.Password)
		if err != nil {
			ctx.Logger.WithError(err).Error("error getting user credentials")
			http.Error(w, `{"code": 401, "message": "Invalid credentials"}`, http.StatusUnauthorized)
			return
		}
		user, err = rt.db.GetUserByID(userID)
		if err != nil {
			ctx.Logger.WithError(err).Error("error retrieving user details")
			http.Error(w, `{"code": 500, "message": "Internal server error"}`, http.StatusInternalServerError)
			return
		}
	} else {
		// User doesn't exist, create a new user
		user = models.User{
			Username: req.Name,
			Name:     req.Name, // 默认使用用户名作为昵称
		}

		user, err = rt.db.CreateUser(user, req.Password)
		if err != nil {
			ctx.Logger.WithError(err).Error("error creating user")
			http.Error(w, `{"code": 500, "message": "Internal server error"}`, http.StatusInternalServerError)
			return
		}
	}

	// Return success response
	resp := LoginResponse{
		Identifier: user.ID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode response")
		http.Error(w, `{"code": 500, "message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Log successful login
	ctx.Logger.Infof("User %s logged in successfully", user.ID)
}
