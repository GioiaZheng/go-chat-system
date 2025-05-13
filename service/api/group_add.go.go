package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type AddGroupMembersRequest struct {
	UserIDs []string `json:"user_ids"`
}

// addToGroup handles POST /groups/:id/members
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	groupID := ps.ByName("groupId")

	var req AddGroupMembersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := rt.db.AddGroupMembers(groupID, req.UserIDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to add members to group")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"code":    201,
		"message": "Members added successfully",
	})
}
