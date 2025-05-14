package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type StartConversationRequest struct {
	MemberIds []string `json:"memberIds"`
	Name      string   `json:"name"`
}

func (rt *_router) startConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var req StartConversationRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if req.Name == "" || len(req.MemberIds) == 0 {
		writeError(w, http.StatusBadRequest, "Missing name or memberIds")
		return
	}

	if len(req.Name) > 50 {
		writeError(w, http.StatusBadRequest, "Name too long (max 50 characters)")
		return
	}

	userID := ctx.UserID
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
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"code":    201,
		"message": "Conversation started successfully",
		"data":    conversation,
	})
}
