package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// searchUsers handles GET /users/search?q=...&limit=...
// English notes:
//   - q is required; limit is optional (default 20, clamp to [1,100]).
//   - DB method signature (from your build error):
//     SearchUsers(ctx context.Context, q string, limit string)
//   - We pass r.Context() and strconv.Itoa(limit).
func (rt *_router) searchUsers(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if strings.TrimSpace(ctx.UserID) == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		rt.sendError(w, http.StatusBadRequest, "q is required")
		return
	}

	// parse & clamp limit
	limit := 20
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			if n < 1 {
				n = 1
			}
			if n > 100 {
				n = 100
			}
			limit = n
		}
	}

	items, err := rt.db.SearchUsers(r.Context(), q, strconv.Itoa(limit))
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
