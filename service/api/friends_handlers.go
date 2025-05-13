package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getFriendsList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := GetUserIDFromContext(r.Context()) // ⭐️ 从Context里拿，不从URL拿！

	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	friends, err := rt.db.GetFriendsList(userID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to get friends list")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"code":    200,
		"message": "Friends list fetched successfully",
		"data": map[string]any{
			"friends": friends,
		},
	}
	writeJSON(w, http.StatusOK, resp)
}
