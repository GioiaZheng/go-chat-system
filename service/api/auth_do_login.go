package api

import (
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/julienschmidt/httprouter"
)

// loginRequest matches POST /session body (simplified login)
type loginRequest struct {
	Name string `json:"name"`
}

// authEnvelope matches the OpenAPI AuthEnvelope: {code, message, data:{user, token}}
type authEnvelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		User  models.User `json:"user"`
		Token string      `json:"token"` // user.ID used as bearer token
	} `json:"data"`
}

// doLogin handles POST /session
// Behavior per api.yaml: if the user doesn't exist -> create; otherwise return existing.
// Request: { "name": "<username>" }    Response: 201 + {code,message,data:{user,token}}
// IMPORTANT: To avoid changing the DB interface, we rely on GetUserByCredentials(name, "")
// for the simplified login (empty password). When creating new users we also store "" as password.
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

	// 1) Check if the user already exists
	exists, err := rt.db.CheckUserExists(name)
	if err != nil {
		rt.writeErrorResponse(w, http.StatusInternalServerError, "Database error")
		return
	}

	var user models.User

	if exists {
		// 2a) Simplified login path for existing users:
		//     Use empty password as the credential (no DB API changes required).
		u, err := rt.db.GetUserByCredentials(name, "")
		if err != nil {
			// If your dataset contains legacy users with real passwords,
			// they won't be accessible via simplified login. In that case,
			// we explicitly return 401 to match the assignment behavior.
			rt.writeErrorResponse(w, http.StatusUnauthorized, "Invalid credentials for simplified login")
			return
		}
		user = u
	} else {
		// 2b) Create the user with empty password (assignment requires simplified login)
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

	// 3) Build response according to OpenAPI (201 + single object in data)
	var resp authEnvelope
	resp.Code = http.StatusCreated
	resp.Message = "Login successful"
	resp.Data.User = user
	resp.Data.Token = user.ID // use user ID as bearer token

	if err := writeJSON(w, http.StatusCreated, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode login response")
	}
}
