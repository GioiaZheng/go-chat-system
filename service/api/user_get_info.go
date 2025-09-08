package api

import (
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getUserInfo handles GET /users/me
// English notes:
// - Use db.GetUser(userID) instead of GetUserInfo (not defined).
// - Returns current authenticated user.
func (rt *_router) getUserInfo(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	userID := strings.TrimSpace(ctx.UserID)
	if userID == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := rt.db.GetUser(userID) // ✅ fixed: use GetUser
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to get user info")
		rt.sendError(w, http.StatusInternalServerError, "Failed to fetch user info")
		return
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "User info retrieved",
		"data": map[string]interface{}{
			"user": user,
		},
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode user info response")
	}
}
