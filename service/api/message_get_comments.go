package api

import (
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getMessages(
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

	q := r.URL.Query()
	conversationID := strings.TrimSpace(q.Get("conversation_id"))
	chatType := strings.ToLower(strings.TrimSpace(q.Get("chat_type")))
	targetID := strings.TrimSpace(q.Get("target_id"))

	var (
		items []models.Message
		err   error
	)

	switch {
	// 1) Prefer conversation_id (infer whether it's group or private)
	case conversationID != "":
		members, mErr := rt.db.GetConversationMembers(conversationID)
		if mErr != nil {
			ctx.Logger.WithError(mErr).Error("failed to get conversation members")
			rt.sendError(w, http.StatusNotFound, "Conversation not found")
			return
		}
		if len(members) >= 3 {
			// group
			items, err = rt.db.GetGroupConversation(conversationID)
		} else if len(members) == 2 {
			// private: pick the "other" user
			var other string
			for _, mid := range members {
				if mid != userID {
					other = mid
					break
				}
			}
			if other == "" {
				rt.sendError(w, http.StatusBadRequest, "Invalid conversation members")
				return
			}
			items, err = rt.db.GetPrivateConversation(userID, other)
		} else {
			rt.sendError(w, http.StatusBadRequest, "Invalid conversation (member count)")
			return
		}

	// 2) Legacy: private
	case chatType == "private" && targetID != "":
		items, err = rt.db.GetPrivateConversation(userID, targetID)

	// 3) Legacy: group
	case chatType == "group" && targetID != "":
		items, err = rt.db.GetGroupConversation(targetID)

	default:
		rt.sendError(w, http.StatusBadRequest, "Provide conversation_id or (chat_type + target_id)")
		return
	}

	if err != nil {
		ctx.Logger.WithError(err).Error("failed to get messages")
		rt.sendError(w, http.StatusInternalServerError, "Failed to get messages")
		return
	}
	if items == nil {
		items = make([]models.Message, 0)
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Messages retrieved",
		"items":   items,
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode get messages response")
	}
}
