package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	messageID := ps.ByName("messageId")
	userID := getUserIDFromContext(r)

	err := rt.db.DeleteMessage(userID, messageID)
	if err != nil {
		writeError(w, http.StatusForbidden, "Delete failed: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
