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
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    []UserResponse `json:"data"`
}

type UserResponse struct {
	User  models.User `json:"user"`
	Token string      `json:"token,omitempty"` // user.ID
}

func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" || req.Password == "" {
		rt.writeErrorResponse(w, http.StatusBadRequest, "Name and password are required")
		return
	}

	var user models.User
	exists, err := rt.db.CheckUserExists(req.Name)
	if err != nil {
		rt.writeErrorResponse(w, http.StatusInternalServerError, "Database error")
		return
	}

	if exists {
		user, err = rt.db.GetUserByCredentials(req.Name, req.Password)
		if err != nil {
			rt.writeErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
	} else {
		user = models.User{
			Username: req.Name,
			Name:     req.Name,
		}
		user, err = rt.db.CreateUser(user, req.Password)
		if err != nil {
			rt.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create user")
			return
		}
	}

	resp := LoginResponse{
		Code:    http.StatusOK,
		Message: "Login successful",
		Data: []UserResponse{
			{
				User:  user,
				Token: user.ID, // curl 可以放入 Authorization 头中
			},
		},
	}

	rt.writeJSONResponse(w, http.StatusOK, resp)
}
