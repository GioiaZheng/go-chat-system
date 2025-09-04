package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/julienschmidt/httprouter"
)

// Flexible request that accepts both legacy {"name","password"} and OpenAPI {"username","email","password",...}.
type registerFlexibleReq struct {
	// Legacy fields
	Name     string `json:"name,omitempty"`
	Password string `json:"password,omitempty"`
	// OpenAPI fields
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Gender   string `json:"gender,omitempty"`
}

// OpenAPI-shaped response: data is a single object (not an array)
type registerResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    registerDataObject `json:"data"`
}

type registerDataObject struct {
	User  models.User `json:"user"`
	Token string      `json:"token,omitempty"`
}

// doRegister handles POST /register
// Compatibility rules:
// - If the request contains {name,password}, we map username=name and email=username+"@example.com".
// - If the request contains {username,email,password}, we use them directly.
// - "name" in OpenAPI is optional; if missing we mirror username into user.Name.
func (rt *_router) doRegister(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req registerFlexibleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Normalize & compatibility layer
	username := strings.TrimSpace(req.Username)
	email := strings.TrimSpace(req.Email)
	password := strings.TrimSpace(req.Password)
	displayName := strings.TrimSpace(req.Name)
	gender := strings.TrimSpace(req.Gender)

	// Legacy payload fallback: {"name","password"}
	if username == "" && req.Name != "" {
		username = strings.TrimSpace(req.Name)
	}
	if password == "" && req.Password != "" {
		password = strings.TrimSpace(req.Password)
	}
	if email == "" && username != "" {
		// Auto-generate email for legacy path to satisfy schema
		email = username + "@example.com"
	}
	if displayName == "" && username != "" {
		displayName = username
	}
	if gender == "" {
		gender = "unspecified"
	}

	// Validate minimal requirements
	if username == "" || email == "" || password == "" {
		rt.writeErrorResponse(w, http.StatusBadRequest, "Username, email and password are required")
		return
	}

	// Check duplicates
	exists, err := rt.db.CheckUserExists(username)
	if err != nil {
		rt.writeErrorResponse(w, http.StatusInternalServerError, "Database error")
		return
	}
	if exists {
		rt.writeErrorResponse(w, http.StatusConflict, "User already exists")
		return
	}

	// Build model and create
	user := models.User{
		Username:  username,
		Email:     email,
		Name:      displayName,
		Gender:    gender,
		AvatarUrl: "https://example.com/default-avatar.png",
	}
	created, err := rt.db.CreateUser(user, password)
	if err != nil {
		rt.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	resp := registerResponse{
		Code:    http.StatusCreated,
		Message: "User registered successfully",
		Data: registerDataObject{
			User:  created,
			Token: created.ID, // In this assignment, we use userID as the bearer token
		},
	}

	if err := writeJSON(w, http.StatusCreated, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode register response")
	}
}

// writeErrorResponse is a small helper for auth endpoints (Go 1.17 friendly: interface{} instead of any)
func (rt *_router) writeErrorResponse(w http.ResponseWriter, status int, msg string) {
	_ = writeJSON(w, status, map[string]interface{}{
		"code":    status,
		"message": msg,
	})
}
