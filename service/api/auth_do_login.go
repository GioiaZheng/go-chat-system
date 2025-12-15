// auth_do_login.go implements the session login endpoint and its response
// envelope to match the OpenAPI user/token contract. Authentication is based on
// a display name and returns a token equal to the user ID for simplicity.
package api

import (
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/julienschmidt/httprouter"
)

// loginRequest represents the POST /session payload.
type loginRequest struct {
	Name string `json:"name"`
}

// apiUser mirrors the OpenAPI User shape (id, name, avatarUri).
type apiUser struct {
	ID        string `json:"id"`
	Name      string `json:"name,omitempty"`
	AvatarURI string `json:"avatarUri,omitempty"`
}

// authEnvelope is the response wrapper for successful login requests.
type authEnvelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		User  apiUser `json:"user"`
		Token string  `json:"token"`
	} `json:"data"`
}

func toAPIUser(u models.User) apiUser {
	return apiUser{
		ID:        u.ID,
		Name:      u.Name,
		AvatarURI: u.AvatarUrl,
	}
}

// doLogin implements POST /session.
// It authenticates by display name only, creating the user on first login,
// and always returns a 200 response containing {user, token} where token ==
// userID.
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

	// Check whether the username is already registered.
	exists, err := rt.db.CheckUserExists(name)
	if err != nil {
		rt.writeErrorResponse(w, http.StatusInternalServerError, "Database error")
		return
	}

	var user models.User
	if exists {
		// Find the user by identifier (name) then load its full record.
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
		// Create a new user on first login.
		u, err := rt.db.CreateUser(models.User{
			Name: name,
		})
		if err != nil {
			rt.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create user")
			return
		}
		user = u
	}

	var resp authEnvelope
	resp.Code = http.StatusOK
	resp.Message = "Login successful"
	resp.Data.User = toAPIUser(user)
	resp.Data.Token = user.ID // token == userID

	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode login response")
	}
}
