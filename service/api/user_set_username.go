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
// English notes:
// - Replaced http.Error with rt.sendError (consistent JSON error responses).
// - On DB error, logs via ctx.Logger and returns 500.
// - On success, uses writeJSON with envelope.
func (rt *_router) setMyUserName(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	userID := ctx.UserID
	if userID == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req SetUserNameRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Username validation (ASCII letters, digits, underscore, hyphen)
	usernamePattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if req.Username == "" || !usernamePattern.MatchString(req.Username) || len(req.Username) > 50 {
		rt.sendError(w, http.StatusBadRequest,
			"Invalid username: must be 1-50 characters long and contain only letters, numbers, underscores, or hyphens")
		return
	}

	// Update username
	if err := rt.db.UpdateUserName(userID, req.Username); err != nil {
		ctx.Logger.WithError(err).Error("failed to update username")
		rt.sendError(w, http.StatusInternalServerError, "Failed to update username")
		return
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Username updated successfully",
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode set username response")
	}
}
