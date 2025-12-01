// file: service/api/auth_register_user.go
package api

import (
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/julienschmidt/httprouter"
)

type registerRequest struct {
	Name string `json:"name"`
}

func (rt *_router) doRegister(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req registerRequest
	if err := readJSON(r, &req); err != nil {
		rt.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		rt.writeErrorResponse(w, http.StatusBadRequest, "Name is required")
		return
	}

	// If already exists, just return that user (idempotent behavior)
	exists, err := rt.db.CheckUserExists(name)
	if err != nil {
		rt.writeErrorResponse(w, http.StatusInternalServerError, "Database error")
		return
	}

	var user models.User
	if exists {
		id, err := rt.db.GetUserIDFromIdentifier(name)
		if err != nil {
			rt.writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch user")
			return
		}
		u, err := rt.db.GetUserByID(id)
		if err != nil {
			rt.writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch user")
			return
		}
		user = u
	} else {
		u, err := rt.db.CreateUser(models.User{Name: name})
		if err != nil {
			rt.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create user")
			return
		}
		user = u
	}

	// Return AuthEnvelope (201)
	resp := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			User  apiUser `json:"user"`
			Token string  `json:"token"`
		} `json:"data"`
	}{}
	resp.Code = http.StatusCreated
	resp.Message = "User registered successfully"
	resp.Data.User = toAPIUser(user)
	resp.Data.Token = user.ID

	_ = writeJSON(w, http.StatusCreated, resp)
}
