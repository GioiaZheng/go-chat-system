package api

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// UpdateGroupNameRequest matches OpenAPI spec
type UpdateGroupNameRequest struct {
	Name string `json:"name"`
}

// setGroupName handles PUT /groups/:id/name
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("id") // Changed from "groupId" to "id"
	if groupID == "" {
		rt.sendError(w, http.StatusBadRequest, "Group ID required", "validation.failed", []string{"id parameter is required"})
		return
	}

	// Validate group ID format
	groupIDPattern := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	if !groupIDPattern.MatchString(groupID) || len(groupID) > 64 {
		rt.sendError(w, http.StatusBadRequest, "Invalid group ID", "validation.failed", []string{"id: Invalid format"})
		return
	}

	// Parse request body
	var req UpdateGroupNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "Invalid request body", "validation.failed", []string{"Invalid JSON format"})
		return
	}

	// Validate group name
	namePattern := regexp.MustCompile(`^[a-zA-Z0-9\s_-]+$`)
	if req.Name == "" || !namePattern.MatchString(req.Name) || len(req.Name) > 100 {
		rt.sendError(w, http.StatusBadRequest, "Invalid name", "validation.failed",
			[]string{"name: Must be 1-100 characters and match pattern [a-zA-Z0-9\\s_-]+"})
		return
	}

	// Update group name
	if err := rt.db.UpdateGroupName(groupID, req.Name); err != nil {
		rt.baseLogger.WithError(err).Error("Failed to update group name")
		rt.sendError(w, http.StatusInternalServerError, "Failed to update group name", "group.update_failed", []string{err.Error()})
		return
	}

	// Return success response
	response := map[string]interface{}{
		"code":    200,
		"message": "Group name updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
