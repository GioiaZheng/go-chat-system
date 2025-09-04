package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/julienschmidt/httprouter"
)

type RegisterRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    []UserResponse `json:"data"`
}

// doRegister handles POST /register
func (rt *_router) doRegister(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req RegisterRequest
	if err := readJSON(r, &req); err != nil {
		rt.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" || req.Password == "" {
		rt.writeErrorResponse(w, http.StatusBadRequest, "Name and password are required")
		return
	}

	// If user exists, return 409 Conflict to avoid duplicates
	exists, err := rt.db.CheckUserExists(req.Name)
	if err != nil {
		rt.writeErrorResponse(w, http.StatusInternalServerError, "Database error")
		return
	}
	if exists {
		rt.writeErrorResponse(w, http.StatusConflict, "User already exists")
		return
	}

	// Create user
	user := models.User{
		Username: req.Name,
		Name:     req.Name,
	}
	user, err = rt.db.CreateUser(user, req.Password)
	if err != nil {
		rt.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	resp := RegisterResponse{
		Code:    http.StatusCreated,
		Message: "User created",
		Data: []UserResponse{
			{
				User:  user,
				Token: user.ID, // This token is used as Bearer token in this assignment
			},
		},
	}
	if err := writeJSON(w, http.StatusCreated, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode register response")
	}
}

// writeErrorResponse is a small helper for auth endpoints
func (rt *_router) writeErrorResponse(w http.ResponseWriter, status int, msg string) {
	_ = writeJSON(w, status, map[string]interface{}{
		"code":    status,
		"message": msg,
	})
}
