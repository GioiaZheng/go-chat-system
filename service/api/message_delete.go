package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	messageID := ps.ByName("messageId")
	userID := ctx.UserID

	err := rt.db.DeleteMessage(userID, messageID)
	if err != nil {
		writeError(w, http.StatusForbidden, "Delete failed: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
