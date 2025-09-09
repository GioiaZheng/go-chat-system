package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getMessageComments(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	_ reqcontext.RequestContext,
) {
	msgID := ps.ByName("id")
	if msgID == "" {
		rt.sendError(w, http.StatusBadRequest, "message id is required")
		return
	}

	// Validate existence (optional but helpful)
	if _, err := rt.db.GetMessageByID(msgID); err != nil {
		rt.sendError(w, http.StatusNotFound, "message not found")
		return
	}

	// OpenAPI: 200 -> { comments: [ {id, author_id, content, created_at}, ... ] }
	// 这里先返回空列表（等 DB 实现后替换）
	payload := map[string]interface{}{
		"comments": []interface{}{},
	}

	if err := writeJSON(w, http.StatusOK, payload); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode message comments response")
	}
}
