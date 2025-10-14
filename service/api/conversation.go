package api

import (
	"net/http"
	"strings"

	reqcontext "github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// StartConversationRequest accepts both the OpenAPI field (memberIds)
// and a legacy snake_case field (member_ids) for backward compatibility.
type StartConversationRequest struct {
	MemberIDs       []string `json:"memberIds,omitempty"`  // OpenAPI (camelCase)
	LegacyMemberIDs []string `json:"member_ids,omitempty"` // legacy (snake_case)
	Name            string   `json:"name,omitempty"`
}

// startConversation handles POST /conversations.
//
// Flow:
//  1. Ensure the request is authenticated (ctx.UserID).
//  2. Parse and validate JSON body.
//  3. Merge memberIds and legacy member_ids, normalize and deduplicate.
//  4. Ensure the caller is included as a member.
//  5. Delegate to the DB to create the conversation.
//  6. Return 201 Created with { code, message, data: { conversation } }.
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

	// 3) Merge memberIds (OpenAPI) and member_ids (legacy)
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

	// 4) Ensure caller is included as a member
	push(ctx.UserID)

	// Optional minimal validation to mirror the spec spirit:
	// require at least one member (the caller will always be there after push).
	if len(finalMembers) == 0 {
		rt.sendError(w, http.StatusBadRequest, "At least one member is required")
		return
	}

	// 5) Create conversation via DB
	conv, err := rt.db.StartConversation(r.Context(), ctx.UserID, finalMembers, strings.TrimSpace(req.Name))
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to start conversation")
		rt.sendError(w, http.StatusInternalServerError, "Failed to start conversation")
		return
	}

	// 6) Success response aligned with OpenAPI (ConversationEnvelope)
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

// getMyConversations handles GET /conversations
// It returns the current user's conversation list.
func (rt *_router) getMyConversations(
	w http.ResponseWriter,
	_ *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	uid := strings.TrimSpace(ctx.UserID)
	if uid == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	convs, err := rt.db.GetMyConversations(uid)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "Failed to fetch conversations")
		return
	}

	// OpenAPI 里外层是 {code, message?, data:{items:[]}}
	resp := map[string]interface{}{
		"code": 200,
		"data": map[string]interface{}{
			"items": convs,
		},
	}
	_ = writeJSON(w, http.StatusOK, resp)
}
