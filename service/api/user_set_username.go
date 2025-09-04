package api

import (
	"net/http"
	"regexp"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type SetUserNameRequest struct {
	Username string `json:"username"`
}

// setMyUserName handles PUT /users/set_username
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req SetUserNameRequest
	if err := readJSON(r, &req); err != nil {
		http.Error(w, `{"code": 400, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Username validation (ASCII letters, digits, underscore, hyphen)
	usernamePattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if req.Username == "" || !usernamePattern.MatchString(req.Username) || len(req.Username) > 50 {
		http.Error(w, `{"code": 400, "message": "Invalid username: must be 1-50 characters long and contain only letters, numbers, underscores, or hyphens"}`, http.StatusBadRequest)
		return
	}

	// Update username
	if err := rt.db.UpdateUserName(userID, req.Username); err != nil {
		rt.baseLogger.WithError(err).Error("failed to update username")
		http.Error(w, `{"code": 500, "message": "Failed to update username"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Username updated successfully",
	}

	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode set username response")
	}
}
