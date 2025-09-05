package api

import (
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// StartConversationRequest models the expected JSON payload.
// NOTE: Keep field names as-is to remain compatible with the existing OpenAPI and clients.
type StartConversationRequest struct {
	Name      string   `json:"name,omitempty"`
	MemberIDs []string `json:"memberIds,omitempty"`
}

// startConversation handles POST /conversations.
// Flow:
//  1) Ensure the request is authenticated (ctx.UserID set by wrap middleware).
//  2) Parse JSON body and normalize the member list (trim, dedupe, drop empties).
//  3) Ensure the caller is included among the members.
//  4) Delegate to db.StartConversation(...).
//  5) Reply with 201 Created and the created conversation.
//
// IMPORTANT: We keep the response shape stable to match the rest of the project:
//   { code, message, data: { conversation: ... } }
func (rt *_router) startConversation(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// 1) Auth check (wrap middleware should have set ctx.UserID; we double check for safety).
	if strings.TrimSpace(ctx.UserID) == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 2) Parse & validate JSON body.
	var req StartConversationRequest
	if err := readJSON(r, &req); err != nil {
		// READABLE: return a clear 400 on malformed JSON or disallowed fields.
		rt.sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// 2.1) Normalize member IDs: trim spaces, drop empties, dedupe.
	finalMembers := make([]string, 0, len(req.MemberIDs)+1)
	seen := make(map[string]struct{})
	push := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		finalMembers = append(finalMembers, id)
	}
	for _, id := range req.MemberIDs {
		push(id)
	}

	// 3) Ensure caller is included.
	push(ctx.UserID)

	// 4) Delegate to the DB.
	conv, err := rt.db.StartConversation(r.Context(), ctx.UserID, finalMembers, strings.TrimSpace(req.Name))
	if err != nil {
		// READABLE: hide internals, log for diagnostics.
		ctx.Logger.WithError(err).Error("failed to start conversation")
		rt.sendError(w, http.StatusInternalServerError, "Failed to start conversation")
		return
	}

	// 5) Success response (201 Created) with a stable envelope.
	resp := map[string]interface{}{
		"code":    http.StatusCreated,
		"message": "Conversation created",
		"data": map[string]interface{}{
			"conversation": conv,
		},
	}
	if err := writeJSON(w, http.StatusCreated, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode start conversation response")
	}
}
