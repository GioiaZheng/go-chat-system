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

// registerFlexibleReq accepts both the OpenAPI shape { "username","password","name?" }
// and the legacy shape { "name","password" } where username := name.
type registerFlexibleReq struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Name     string `json:"name,omitempty"`
}

// registerResponse matches the OpenAPI AuthEnvelope:
// { code, message, data: { user, token } }.
// NOTE: "user" uses the outward-facing apiUser (avatarUri) via toAPIUser().
type registerResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		User  apiUser `json:"user"`
		Token string  `json:"token,omitempty"` // For this assignment, token == userID
	} `json:"data"`
}

// doRegister handles POST /register.
// OpenAPI path:
//   - Input:  { "username": string, "password": string, "name"?: string }
//   - Output: 201 Created with { code, message, data: { user, token } }.
//
// Compatibility:
//   - If "username" is empty but "name" is provided (legacy), we set username := name.
//   - Token is an opaque string equal to user.ID (same as /session and auth middleware).
func (rt *_router) doRegister(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req registerFlexibleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	displayName := strings.TrimSpace(req.Name)

	// Legacy compatibility: if username missing but name provided, use it as username.
	if username == "" && displayName != "" {
		username = displayName
	}
	// If display name still empty, default to username.
	if displayName == "" {
		displayName = username
	}

	// Minimal validation. You can tighten this to the OpenAPI constraints if desired:
	//   username: min 3, max 50, pattern ^[a-zA-Z0-9_-]+$
	//   password: min 6, max 100
	if username == "" || password == "" {
		rt.writeErrorResponse(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	// Check if the username is already taken.
	exists, err := rt.db.CheckUserExists(username)
	if err != nil {
		rt.writeErrorResponse(w, http.StatusInternalServerError, "Database error")
		return
	}
	if exists {
		// 409 aligns with usual semantics for "already exists".
		rt.writeErrorResponse(w, http.StatusConflict, "User already exists")
		return
	}

	// Pick a random default avatar from /photos, which is served by the router.
	rand.Seed(time.Now().UnixNano())
	avatarIndex := rand.Intn(10) + 1
	avatarURL := fmt.Sprintf("/photos/avatar%d.jpg", avatarIndex)

	// Create the user in the database.
	user := models.User{
		Username:  username,
		Name:      displayName,
		AvatarUrl: avatarURL, // internal field -> will be exposed as avatarUri via toAPIUser
	}
	created, err := rt.db.CreateUser(user, password)
	if err != nil {
		rt.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Build the OpenAPI-compliant response.
	var resp registerResponse
	resp.Code = http.StatusCreated
	resp.Message = "User registered successfully"
	resp.Data.User = toAPIUser(created) // map internal to outward-facing user shape
	resp.Data.Token = created.ID        // token == userID

	if err := writeJSON(w, http.StatusCreated, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode register response")
	}
}
