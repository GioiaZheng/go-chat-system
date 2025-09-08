package api

import (
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getUserProfile handles GET /users/profile/:user_id
// English notes:
// - Your DB doesn't have GetUserProfile(); use GetUser(userID) instead.
// - If your real method is named differently (e.g., GetUserByID), just rename here.
func (rt *_router) getUserProfile(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if strings.TrimSpace(ctx.UserID) == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID := strings.TrimSpace(ps.ByName("user_id"))
	if userID == "" {
		rt.sendError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	user, err := rt.db.GetUser(userID) // <— 如果你的方法叫 GetUserByID，请改名
	if err != nil {
		// Treat as not found to avoid leaking details
		rt.sendError(w, http.StatusNotFound, "User not found")
		return
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "User profile retrieved",
		"data": map[string]interface{}{
			"user": user,
		},
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode user profile response")
	}
}
