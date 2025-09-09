package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/julienschmidt/httprouter"
)

// Flexible request: supports both {"name","password"} (legacy) and
// {"username","email","password","gender"} (OpenAPI).
type registerFlexibleReq struct {
	Name     string `json:"name,omitempty"`
	Password string `json:"password,omitempty"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Gender   string `json:"gender,omitempty"`
}

// Response format matching OpenAPI: {code,message,data:{user,token}}
type registerResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    registerDataObject `json:"data"`
}

type registerDataObject struct {
	User  models.User `json:"user"`
	Token string      `json:"token,omitempty"`
}

// doRegister handles POST /register.
// - Legacy: {"name","password"} → username=name, email=auto.
// - OpenAPI: {"username","email","password"} → used directly.
// - If name missing, fallback to username.
func (rt *_router) doRegister(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req registerFlexibleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	username := strings.TrimSpace(req.Username)
	email := strings.TrimSpace(req.Email)
	password := strings.TrimSpace(req.Password)
	displayName := strings.TrimSpace(req.Name)
	gender := strings.TrimSpace(req.Gender)

	// Legacy compatibility
	if username == "" && req.Name != "" {
		username = strings.TrimSpace(req.Name)
	}
	if password == "" && req.Password != "" {
		password = strings.TrimSpace(req.Password)
	}
	if email == "" && username != "" {
		email = username + "@example.com"
	}
	if displayName == "" && username != "" {
		displayName = username
	}
	if gender == "" {
		gender = "unspecified"
	}

	// Minimal validation
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

	// Pick random default avatar
	rand.Seed(time.Now().UnixNano())
	avatarIndex := rand.Intn(10) + 1
	avatarURL := fmt.Sprintf("/uploads/photos/avatar%d.jpg", avatarIndex)

	// Create user
	user := models.User{
		Username:  username,
		Email:     email,
		Name:      displayName,
		Gender:    gender,
		AvatarUrl: avatarURL,
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
			Token: created.ID, // assignment: use userID as bearer token
		},
	}

	if err := writeJSON(w, http.StatusCreated, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode register response")
	}
}
