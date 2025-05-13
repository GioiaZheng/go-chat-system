package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/julienschmidt/httprouter"
)

// doLogin handles POST /session: 用户登录
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.baseLogger.WithError(err).Error("failed to decode login request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		rt.baseLogger.Error("email or password missing in login request")
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	user, token, err := rt.db.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		rt.baseLogger.WithError(err).Error("authentication failed")
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	response := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string      `json:"token"`
			User  models.User `json:"user"`
		} `json:"data"`
	}{
		Code:    200,
		Message: "Login successful",
		Data: struct {
			Token string      `json:"token"`
			User  models.User `json:"user"`
		}{
			Token: token,
			User:  user,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode login response")
		w.WriteHeader(http.StatusInternalServerError)
	}
}
