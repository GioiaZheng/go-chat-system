package api

import (
	"encoding/json"
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type getCommentsResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// GET /api/v1/messages/:id/comment
func (rt *_router) getMessageComments(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	msgID := ps.ByName("id")
	if msgID == "" {
		rt.sendError(w, http.StatusBadRequest, "message id is required")
		return
	}

	// 先校验消息是否存在
	if _, err := rt.db.GetMessageByID(msgID); err != nil {
		rt.sendError(w, http.StatusNotFound, "message not found")
		return
	}

	// 占位返回（后续接入真实评论数据）
	payload := map[string]interface{}{
		"message_id": msgID,
		"comments":   []interface{}{}, // TODO: 替换为真实的评论数组
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(getCommentsResponse{
		Code:    200,
		Message: "ok",
		Data:    payload,
	})
}
