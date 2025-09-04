package api

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// searchUsers handles GET /users/search?q=...
func (rt *_router) searchUsers(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code": 401, "message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Parse query string
	queryParams, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, `{"code": 400, "message": "Invalid query parameters"}`, http.StatusBadRequest)
		return
	}

	query := strings.TrimSpace(queryParams.Get("q"))
	if query == "" {
		http.Error(w, `{"code": 400, "message": "Query parameter 'q' is required"}`, http.StatusBadRequest)
		return
	}

	// Pattern deliberately does NOT allow spaces or non-ASCII
	queryPattern := regexp.MustCompile(`^[a-zA-Z0-9_\-@.]+$`)
	if !queryPattern.MatchString(query) || len(query) > 255 {
		http.Error(w, `{"code": 400, "message": "Invalid query"}`, http.StatusBadRequest)
		return
	}

	users, err := rt.db.SearchUsers(r.Context(), userID, query)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to search users")
		http.Error(w, `{"code": 500, "message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	items := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		items = append(items, map[string]interface{}{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"gender":    user.Gender,
			"avatarUrl": user.AvatarUrl,
		})
	}

	// Ensure empty array instead of null
	if items == nil {
		items = make([]map[string]interface{}, 0)
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Search successful",
		"items":   items,
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode user search response")
	}
}
