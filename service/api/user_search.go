package api

import (
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) searchUsers(
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

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		rt.sendError(w, http.StatusBadRequest, "q is required")
		return
	}

	items, err := rt.db.SearchUsers(r.Context(), userID, q)
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to search users")
		rt.sendError(w, http.StatusInternalServerError, "Failed to search users")
		return
	}
	if items == nil {
		items = make([]models.User, 0)
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Users search retrieved",
		"items":   items,
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode search users response")
	}
}
