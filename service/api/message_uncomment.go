package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// uncommentMessage handles DELETE /messages/:id/comment
// It removes the comment associated with the given message.
func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	messageID := ps.ByName("id")
	if messageID == "" {
		http.Error(w, `{"code": 400, "message": "Missing message ID"}`, http.StatusBadRequest)
		return
	}

	if err := rt.db.UncommentMessage(messageID); err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to uncomment message"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Comment removed successfully",
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode uncomment response")
	}
}
