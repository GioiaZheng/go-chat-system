package api

import (
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// StartConversationRequest supports both OpenAPI (member_ids) and
// legacy (memberIds) fields for backward compatibility.
type StartConversationRequest struct {
	MemberIDs       []string `json:"member_ids,omitempty"`
	LegacyMemberIDs []string `json:"memberIds,omitempty"`
	Name            string   `json:"name,omitempty"`
}

// startConversation handles POST /conversations.
// Flow:
//  1. Ensure the request is authenticated (ctx.UserID).
//  2. Parse and validate the JSON body.
//  3. Merge member_ids and legacy memberIds, normalize and deduplicate.
//  4. Ensure the caller is included as a member.
//  5. Call the DB to create the conversation.
//  6. Return 201 Created with the conversation object.
func (rt *_router) startConversation(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// 1) Auth check
	if strings.TrimSpace(ctx.UserID) == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 2) Parse request
	var req StartConversationRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 3) Merge member_ids and legacy memberIds
	finalMembers := make([]string, 0, len(req.MemberIDs)+len(req.LegacyMemberIDs)+1)
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
	for _, id := range req.LegacyMemberIDs {
		push(id)
	}

	// 4) Ensure caller included
	push(ctx.UserID)

	// 5) Delegate to DB
	conv, err := rt.db.StartConversation(r.Context(), ctx.UserID, finalMembers, strings.TrimSpace(req.Name))
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to start conversation")
		rt.sendError(w, http.StatusInternalServerError, "Failed to start conversation")
		return
	}

	// 6) Success response
	resp := map[string]interface{}{
		"code":    http.StatusCreated,
		"message": "Conversation created",
		"data": map[string]interface{}{
			"conversation": conv,
		},
	}
	if err := writeJSON(w, http.StatusCreated, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode startConversation response")
	}
}
