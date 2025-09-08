package api

import (
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// StartConversationRequest models the expected JSON payload.
// ✅ OpenAPI 对齐：主字段使用 snake_case: `member_ids`
// 🤝 兼容旧客户端：同时接受 legacy 字段 `memberIds`，在处理逻辑里合并。
// English notes:
// - We keep both fields in the struct. Only `member_ids` is "official" per OpenAPI.
// - If clients still send `memberIds`, we merge them so nothing breaks during transition.
type StartConversationRequest struct {
	// New / official (OpenAPI)
	MemberIDs []string `json:"member_ids,omitempty"`
	// Legacy / compatibility only
	LegacyMemberIDs []string `json:"memberIds,omitempty"`
	Name            string   `json:"name,omitempty"`
}

// startConversation handles POST /conversations.
// Flow:
//  1. Ensure the request is authenticated (ctx.UserID set by wrap middleware).
//  2. Parse JSON body and normalize the member list (trim, dedupe, drop empties).
//  3. Merge legacy field `memberIds` into `member_ids` for backward compatibility.
//  4. Ensure the caller is included among the members.
//  5. Delegate to db.StartConversation(...).
//  6. Reply with 201 Created and the created conversation.
//
// IMPORTANT: Response shape kept stable across the project:
//
//	{ code, message, data: { conversation: ... } }
func (rt *_router) startConversation(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// 1) Auth check (wrap middleware should have set ctx.UserID; double check for safety).
	if strings.TrimSpace(ctx.UserID) == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 2) Parse & validate JSON body.
	var req StartConversationRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// 3) Merge legacy `memberIds` into `member_ids` and normalize.
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

	// 4) Ensure caller is included.
	push(ctx.UserID)

	// 5) Delegate to the DB.
	conv, err := rt.db.StartConversation(r.Context(), ctx.UserID, finalMembers, strings.TrimSpace(req.Name))
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to start conversation")
		rt.sendError(w, http.StatusInternalServerError, "Failed to start conversation")
		return
	}

	// 6) Success response (201 Created) with a stable envelope.
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
