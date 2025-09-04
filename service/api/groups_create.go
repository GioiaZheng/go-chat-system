package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
)

// CreateGroupRequest defines the request body for creating a group
type CreateGroupRequest struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

// createGroup handles POST /groups
// OpenAPI: respond with { code, message, data: { group: {...} } }
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, `{"code":401,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req CreateGroupRequest
	if err := readJSON(r, &req); err != nil {
		http.Error(w, `{"code":400,"message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Normalize members (unique, non-empty) and ensure creator is included
	seen := map[string]bool{}
	members := make([]string, 0, len(req.Members)+1)
	for _, m := range req.Members {
		if m == "" || seen[m] {
			continue
		}
		seen[m] = true
		members = append(members, m)
	}
	if !seen[userID] {
		seen[userID] = true
		members = append(members, userID)
	}

	groupName := req.Name
	if groupName == "" {
		groupName = "Group"
	}

	// Generate group ID (TEXT PK)
	gid, err := uuid.NewV4()
	if err != nil {
		http.Error(w, `{"code":500,"message":"Failed to generate group id"}`, http.StatusInternalServerError)
		return
	}
	group := models.Group{
		ID:   gid.String(),
		Name: groupName,
	}

	// Create group row
	if err := rt.db.CreateGroup(group); err != nil {
		http.Error(w, `{"code":500,"message":"Failed to create group"}`, http.StatusInternalServerError)
		return
	}

	// Add creator first; if this fails we must abort
	if err := rt.db.AddGroupMembers(group.ID, []string{userID}); err != nil {
		http.Error(w, `{"code":500,"message":"Failed to add creator to group"}`, http.StatusInternalServerError)
		return
	}
	// Add the rest; ignore individual failures
	for _, m := range members {
		if m == userID {
			continue
		}
		_ = rt.db.AddGroupMembers(group.ID, []string{m})
	}

	// Build response to match OpenAPI example: data.group = { id, name, members[...] }
	resp := map[string]interface{}{
		"code":    201,
		"message": "Group created",
		"data": map[string]interface{}{
			"group": map[string]interface{}{
				"id":   group.ID,
				"name": group.Name,
				// "conversationId": "", // optional if your DB supports conversations
				"members": members,
			},
		},
	}
	if err := writeJSON(w, http.StatusCreated, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode create group response")
	}
}
