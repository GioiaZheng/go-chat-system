package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	messageID := ps.ByName("id")

	if messageID == "" {
		http.Error(w, "Missing message ID", http.StatusBadRequest)
		return
	}

	err := rt.db.UncommentMessage(messageID)
	if err != nil {
		http.Error(w, "Failed to uncomment message", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"code":    200,
		"message": "Comment removed successfully",
	}
	writeJSON(w, http.StatusOK, resp)
}
