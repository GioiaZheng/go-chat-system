package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getMyConversations(
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

	conversations, err := rt.db.GetMyConversations(userID)
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to get my conversations")
		rt.sendError(w, http.StatusInternalServerError, "Failed to fetch conversations")
		return
	}

	// OpenAPI: ConversationListEnvelope -> { code, message, items: [...] }
	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Conversations list retrieved",
		"items":   conversations,
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode conversations response")
	}
}
