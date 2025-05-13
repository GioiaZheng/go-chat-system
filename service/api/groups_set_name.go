package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type SetGroupNameRequest struct {
	Name string `json:"name"`
}

func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	groupID := ps.ByName("id")

	var req SetGroupNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Group name is required", http.StatusBadRequest)
		return
	}

	err := rt.db.UpdateGroupName(groupID, req.Name)
	if err != nil {
		http.Error(w, "Failed to update group name", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"code":    200,
		"message": "Group name updated",
	}

	writeJSON(w, http.StatusOK, resp)
}
