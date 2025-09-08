package api

import (
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// AddGroupMembersRequest aligns with OpenAPI which expects an array:
//
//	{ "member_ids": ["u1","u2", ...] }
//
// For backward compatibility we also accept legacy single-field:
//
//	{ "userId": "u1" }   // legacy
//
// English notes:
// - Prefer member_ids (snake_case) per OpenAPI.
// - If legacy userId is present and member_ids is empty, we upgrade it to a single-element array.
type AddGroupMembersRequest struct {
	MemberIDs []string `json:"member_ids,omitempty"`
	LegacyUID string   `json:"userId,omitempty"` // compatibility only
}

// addToGroup handles POST /groups/:id/members
// Behavior:
//   - Validates path param and auth.
//   - Accepts either {member_ids: [...]} or legacy {userId: "..."}.
//   - Calls db.AddGroupMembers with the normalized list.
//   - Returns 200 with a simple envelope.
func (rt *_router) addToGroup(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	groupID := strings.TrimSpace(ps.ByName("id"))
	if groupID == "" {
		rt.sendError(w, http.StatusBadRequest, "missing group id")
		return
	}
	if strings.TrimSpace(ctx.UserID) == "" {
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req AddGroupMembersRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Build the final member list:
	final := make([]string, 0, len(req.MemberIDs)+1)
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
		final = append(final, id)
	}

	// Prefer official field
	for _, id := range req.MemberIDs {
		push(id)
	}
	// Backward compatible: upgrade single userId into the list if needed
	if len(final) == 0 && strings.TrimSpace(req.LegacyUID) != "" {
		push(req.LegacyUID)
	}

	if len(final) == 0 {
		rt.sendError(w, http.StatusBadRequest, "member_ids is required (or legacy userId)")
		return
	}

	// Delegate to DB
	if err := rt.db.AddGroupMembers(groupID, final); err != nil {
		ctx.Logger.WithError(err).Error("failed to add members to group")
		rt.sendError(w, http.StatusInternalServerError, "failed to add member(s) to group")
		return
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Member(s) added to group",
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode add-to-group response")
	}
}
