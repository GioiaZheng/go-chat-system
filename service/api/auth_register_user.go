package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/julienschmidt/httprouter"
)

// doRegister handles POST /users: 用户注册
func (rt *_router) doRegister(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.baseLogger.WithError(err).Error("failed to decode registration request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		rt.baseLogger.Error("missing required fields in registration")
		http.Error(w, "Username, email and password are required", http.StatusBadRequest)
		return
	}

	user, err := rt.db.CreateUser(models.User{
		Username: req.Username,
		Email:    req.Email,
	})
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to create user")
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	response := struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    models.User `json:"data"`
	}{
		Code:    201,
		Message: "User registered successfully",
		Data:    user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode registration response")
	}
}
