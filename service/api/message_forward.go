package api

import (
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// ForwardRequest supports both official and legacy payloads.
// Official (api.yaml):
//   { "conversation_id": "<conv-id>" }
// Legacy (existing):
//   { "toUserId": "<uid>" } or { "toGroupId": "<gid>" }
type ForwardRequest struct {
	ConversationID string `json:"conversation_id,omitempty"` // official
	ToUserID       string `json:"toUserId,omitempty"`        // legacy
	ToGroupID      string `json:"toGroupId,omitempty"`       // legacy
}

// forwardMessage handles POST /messages/:id/forward
func (rt *_router) forwardMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if strings.TrimSpace(ctx.UserID) == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	msgID := strings.TrimSpace(ps.ByName("id"))
	if msgID == "" {
		rt.sendError(w, http.StatusBadRequest, "missing message id")
		return
	}

	var req ForwardRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Normalize: prefer official conversation_id if provided
	toUser := strings.TrimSpace(req.ToUserID)
	toGroup := strings.TrimSpace(req.ToGroupID)

	convID := strings.TrimSpace(req.ConversationID)
	if convID != "" && toUser == "" && toGroup == "" {
		// Heuristic inference to keep DB API unchanged
		lc := strings.ToLower(convID)
		switch {
		case strings.HasPrefix(lc, "g_"),
			strings.HasPrefix(lc, "grp-"),
			strings.HasPrefix(lc, "group-"):
			toGroup = convID
		case strings.HasPrefix(lc, "u_"),
			strings.HasPrefix(lc, "usr-"),
			strings.HasPrefix(lc, "user-"):
			toUser = convID
		default:
			// We don't know how to map this conversation_id to user/group
			rt.sendError(w, http.StatusBadRequest, "conversation_id not recognized; please send toUserId or toGroupId")
			return
		}
	}

	// Validate target
	if (toUser == "" && toGroup == "") || (toUser != "" && toGroup != "") {
		rt.sendError(w, http.StatusBadRequest, "specify exactly one of conversation_id, toUserId, or toGroupId")
		return
	}

	// Call existing DB method (kept unchanged)
	if err := rt.db.ForwardMessage(ctx.UserID, msgID, toUser, toGroup); err != nil {
		ctx.Logger.WithError(err).Error("failed to forward message")
		rt.sendError(w, http.StatusInternalServerError, "failed to forward message")
		return
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Message forwarded successfully",
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode forward response")
	}
}
