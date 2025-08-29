package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	messageID := ps.ByName("id")
	userID := ctx.UserID

	err := rt.db.DeleteMessage(userID, messageID)
	if err != nil {
		writeError(w, http.StatusForbidden, "Delete failed: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    200,
		"message": "Message deleted",
	})

}
