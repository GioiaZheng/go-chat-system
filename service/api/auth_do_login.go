// file: service/api/auth_do_login.go
package api

import (
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/julienschmidt/httprouter"
)

// loginRequest for POST /session
type loginRequest struct {
	Name string `json:"name"`
}

// apiUser matches OpenAPI User (id, name, avatarUri)
type apiUser struct {
	ID        string `json:"id"`
	Name      string `json:"name,omitempty"`
	AvatarURI string `json:"avatarUri,omitempty"`
}

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

// POST /session
// Spec: login by name only; create if not exists; return 200 with {user, token} (token == userID)
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

	// exists?
	exists, err := rt.db.CheckUserExists(name)
	if err != nil {
		rt.writeErrorResponse(w, http.StatusInternalServerError, "Database error")
		return
	}

	var user models.User
	if exists {
		// find by identifier (name) -> id -> load
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
		// create
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
