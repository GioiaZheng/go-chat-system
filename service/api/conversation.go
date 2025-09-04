package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type StartConversationRequest struct {
	Name      string   `json:"name,omitempty"`
	MemberIDs []string `json:"memberIds,omitempty"`
}

// startConversation handles POST /conversations
func (rt *_router) startConversation(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	if ctx.UserID == "" {
		http.Error(w, `{"code":401,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req StartConversationRequest
	if err := readJSON(r, &req); err != nil {
		http.Error(w, `{"code":400,"message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Ensure caller is always added to the conversation
	found := false
	for _, id := range req.MemberIDs {
		if id == ctx.UserID {
			found = true
			break
		}
	}
	if !found {
		req.MemberIDs = append(req.MemberIDs, ctx.UserID)
	}

	conv, err := rt.db.StartConversation(r.Context(), ctx.UserID, req.MemberIDs, req.Name)
	if err != nil {
		rt.baseLogger.WithError(err).Error("failed to start conversation")
		http.Error(w, `{"code":500,"message":"Failed to start conversation"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"code":    201,
		"message": "Conversation created",
		"data": map[string]interface{}{
			"conversation": conv,
		},
	}
	if err := writeJSON(w, http.StatusCreated, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode start conversation response")
	}
}
