package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/julienschmidt/httprouter"
)

// getGroupsList handles GET /groups: 获取用户群组列表
func (rt *_router) getGroupsList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userID := GetUserIDFromContext(r.Context()) // ⭐️ 正确拿用户ID

	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groups, err := rt.db.GetGroupsList(userID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to get groups list")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Groups []models.Group `json:"groups"`
		} `json:"data"`
	}{
		Code:    200,
		Message: "Groups list fetched successfully",
		Data: struct {
			Groups []models.Group `json:"groups"`
		}{
			Groups: groups,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode groups list response")
		w.WriteHeader(http.StatusInternalServerError)
	}
}
