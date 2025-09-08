package api

import (
	"net/http"
	"strings"

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
// English notes:
// - Validate auth and body.
// - Normalize members (dedupe, remove empties), always include creator.
// - Use rt.sendError on failure; writeJSON on success.
func (rt *_router) createGroup(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	userID := strings.TrimSpace(ctx.UserID)
	if userID == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateGroupRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Normalize members
	seen := map[string]bool{}
	members := make([]string, 0, len(req.Members)+1)
	for _, m := range req.Members {
		m = strings.TrimSpace(m)
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

	groupName := strings.TrimSpace(req.Name)
	if groupName == "" {
		groupName = "Group"
	}

	// Generate group ID (TEXT PK)
	gid, err := uuid.NewV4()
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to generate group id")
		rt.sendError(w, http.StatusInternalServerError, "Failed to generate group id")
		return
	}
	group := models.Group{
		ID:   gid.String(),
		Name: groupName,
	}

	// Create group row
	if err := rt.db.CreateGroup(group); err != nil {
		ctx.Logger.WithError(err).Error("failed to create group")
		rt.sendError(w, http.StatusInternalServerError, "Failed to create group")
		return
	}

	// Add creator first; if this fails we must abort
	if err := rt.db.AddGroupMembers(group.ID, []string{userID}); err != nil {
		ctx.Logger.WithError(err).Error("failed to add creator to group")
		rt.sendError(w, http.StatusInternalServerError, "Failed to add creator to group")
		return
	}
	// Add the rest; ignore individual failures (non-fatal)
	for _, m := range members {
		if m == userID {
			continue
		}
		_ = rt.db.AddGroupMembers(group.ID, []string{m})
	}

	resp := map[string]interface{}{
		"code":    http.StatusCreated,
		"message": "Group created",
		"data": map[string]interface{}{
			"group": map[string]interface{}{
				"id":      group.ID,
				"name":    group.Name,
				"members": members,
			},
		},
	}
	if err := writeJSON(w, http.StatusCreated, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode create group response")
	}
}
