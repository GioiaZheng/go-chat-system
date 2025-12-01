// file: service/api/conversations.go
package api

import (
	"net/http"
	"regexp"
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
//  2. Parse and validate JSON body (strictly follow OpenAPI):
//     - name: length 1..100 & pattern ^[a-zA-Z0-9\s'-]+$
//     - memberIds (or member_ids): at least 1 (minItems:1)
//  3. Merge memberIds and legacy member_ids, normalize and deduplicate.
//  4. Ensure the caller is included as a member（校验在加入自己之前已完成）.
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

	// Validate name by OpenAPI spec: 1..100 and pattern ^[a-zA-Z0-9\s'-]+$
	name := strings.TrimSpace(req.Name)
	if name == "" {
		rt.sendError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if ln := len(name); ln < 1 || ln > 100 {
		rt.sendError(w, http.StatusBadRequest, "Name must be 1-100 characters long")
		return
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9\s'-]+$`).MatchString(name) {
		rt.sendError(w, http.StatusBadRequest, "Name contains invalid characters")
		return
	}

	// 合并但先不加入自己，用于严格遵循 YAML 的 minItems:1
	rawMembers := make([]string, 0, len(req.MemberIDs)+len(req.LegacyMemberIDs))
	rawMembers = append(rawMembers, req.MemberIDs...)
	rawMembers = append(rawMembers, req.LegacyMemberIDs...)

	// OpenAPI: memberIds 是必填且 minItems:1
	if len(rawMembers) == 0 {
		rt.sendError(w, http.StatusBadRequest, "memberIds is required and must contain at least 1 item")
		return
	}

	// 3) 去重与清洗
	finalMembers := make([]string, 0, len(rawMembers)+1)
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
	for _, id := range rawMembers {
		push(id)
	}

	// 4) 确保把自己加入（不计入上述 minItems:1 的校验）
	push(ctx.UserID)

	// 5) Create conversation via DB
	conv, err := rt.db.StartConversation(r.Context(), ctx.UserID, finalMembers, name)
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
		ctx.Logger.WithError(err).Error("failed to fetch conversations")
		rt.sendError(w, http.StatusInternalServerError, "Failed to fetch conversations")
		return
	}

	// OpenAPI: {code, data:{items:[]}}
	resp := map[string]interface{}{
		"code": http.StatusOK,
		"data": map[string]interface{}{
			"items": convs,
		},
	}
	_ = writeJSON(w, http.StatusOK, resp)
}


func (rt *_router) deleteConversation(
    w http.ResponseWriter,
    r *http.Request,
    ps httprouter.Params,
    ctx reqcontext.RequestContext,
) {
    userID := strings.TrimSpace(ctx.UserID)
    if userID == "" {
        rt.writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    cid := ps.ByName("id")
    if cid == "" {
        rt.writeErrorResponse(w, http.StatusBadRequest, "missing conversation id")
        return
    }

    // Check membership
    members, err := rt.db.GetConversationMembers(cid)
    if err != nil {
        rt.writeErrorResponse(w, http.StatusInternalServerError, "db error")
        return
    }

    ok := false
    for _, m := range members {
        if m == userID {
            ok = true
            break
        }
    }
    if !ok {
        rt.writeErrorResponse(w, http.StatusForbidden, "not your conversation")
        return
    }

    // Delete conversation
    if err := rt.db.DeleteConversation(cid); err != nil {
        rt.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
