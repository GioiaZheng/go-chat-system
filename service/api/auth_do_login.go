package api

import (
	"net/http"
	"strings"
	// "regexp" // uncomment if you want to enforce the pattern ^[a-zA-Z0-9\s'-]+$

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/julienschmidt/httprouter"
)

// loginRequest models the POST /session request body (simplified login).
type loginRequest struct {
	Name string `json:"name"`
}

// apiUser is the outward-facing user shape that matches the OpenAPI schema
// (notably: avatarUri instead of avatar_url).
type apiUser struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name,omitempty"`
	AvatarURI string `json:"avatarUri,omitempty"`
}

// authEnvelope matches the OpenAPI AuthEnvelope: { code, message, data: { user, token } }.
type authEnvelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		User  apiUser `json:"user"`
		Token string  `json:"token"` // In this assignment the token equals the user ID.
	} `json:"data"`
}

// toAPIUser converts an internal models.User into the outward-facing apiUser.
func toAPIUser(u models.User) apiUser {
	return apiUser{
		ID:        u.ID,
		Username:  u.Username,
		Name:      u.Name,
		AvatarURI: u.AvatarUrl, // map avatar_url -> avatarUri for OpenAPI compliance
	}
}

// doLogin handles POST /session.
//
// Behavior per api.yaml:
// - If the user does not exist, create it (simplified login, no password).
// - Otherwise, "log in" the existing user.
// Request:  { "name": "<username>" }
// Response: 200 + { code, message, data: { user, token } }
//
// IMPORTANT:
// For simplified login we treat token == user.ID (opaque string), which matches the
// rest of this codebase and the auth middleware (wrap). If you later switch to real JWTs,
// update both this handler and the auth extractor accordingly.
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req loginRequest
	if err := readJSON(r, &req); err != nil {
		rt.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		rt.writeErrorResponse(w, http.StatusBadRequest, "Name is required")
		return
	}
	if l := len(name); l < 3 || l > 16 {
		rt.writeErrorResponse(w, http.StatusBadRequest, "Name must be 3-16 characters long")
		return
	}
	// Optionally enforce the OpenAPI pattern: ^[a-zA-Z0-9\s'-]+$
	// re := regexp.MustCompile(`^[a-zA-Z0-9\s'-]+$`)
	// if !re.MatchString(name) {
	// 	rt.writeErrorResponse(w, http.StatusBadRequest, "Name contains invalid characters")
	// 	return
	// }

	// 1) Check if the user already exists by username.
	exists, err := rt.db.CheckUserExists(name)
	if err != nil {
		rt.writeErrorResponse(w, http.StatusInternalServerError, "Database error")
		return
	}

	var user models.User

	if exists {
		// 2a) Simplified login path for existing users:
		//     Keep compatibility with existing DB functions by using empty password.
		u, err := rt.db.GetUserByCredentials(name, "")
		if err != nil {
			// If there are legacy users with non-empty passwords, they won't pass
			// this simplified login. We return 401 to reflect invalid credentials.
			rt.writeErrorResponse(w, http.StatusUnauthorized, "Invalid credentials for simplified login")
			return
		}
		user = u
	} else {
		// 2b) Create the user with an empty password to support simplified login.
		newUser := models.User{
			Username: name,
			Name:     name,
		}
		u, err := rt.db.CreateUser(newUser, "")
		if err != nil {
			rt.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create user")
			return
		}
		user = u
	}

	// 3) Build the response (200 OK per the current OpenAPI).
	var resp authEnvelope
	resp.Code = http.StatusOK
	resp.Message = "Login successful"
	resp.Data.User = toAPIUser(user)
	resp.Data.Token = user.ID // token == userID in this simplified model

	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode login response")
	}
}
