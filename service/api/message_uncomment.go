package api

import (
	"net/http"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 获取消息 ID
	messageID := ps.ByName("id")

	if messageID == "" {
		http.Error(w, `{"code": 400, "message": "Missing message ID"}`, http.StatusBadRequest)
		return
	}

	// 调用数据库方法删除评论
	err := rt.db.UncommentMessage(messageID)
	if err != nil {
		http.Error(w, `{"code": 500, "message": "Failed to uncomment message"}`, http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	resp := map[string]interface{}{
		"code":    200,
		"message": "Comment removed successfully",
	}
	writeJSON(w, http.StatusOK, resp)
}
