package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/julienschmidt/httprouter"
)

type StartConversationRequest struct {
	MemberIds []string `json:"memberIds"`
	Name      string   `json:"name"`
}

func (rt *_router) startConversation(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req StartConversationRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if req.Name == "" || len(req.MemberIds) == 0 {
		writeError(w, http.StatusBadRequest, "Missing name or memberIds")
		return
	}

	userID := getUserIDFromContext(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	conversation, err := rt.db.StartConversation(r.Context(), userID, req.MemberIds, req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to start conversation")
		return
	}

	// 返回 JSON 格式的 conversation
	writeJSON(w, http.StatusCreated, models.Conversation{
		ID:        conversation.ID,
		Name:      conversation.Name,
		AvatarURL: conversation.AvatarURL,
		LastMsg:   conversation.LastMsg,
	})
}
