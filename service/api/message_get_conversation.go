package api

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// GET /api/v1/messages?chat_type=private|group&target_id=...
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	q := r.URL.Query()
	chatType := q.Get("chat_type") // "private" | "group"
	targetID := q.Get("target_id") // 对方用户ID 或 群ID
	if chatType == "" || targetID == "" {
		writeError(w, http.StatusBadRequest, "chat_type and target_id are required")
		return
	}

	var (
		msgs []models.Message
		err  error
	)

	switch chatType {
	case "private":
		// 先按 userID,targetID 查；查不到再反向试一次，避免实现对顺序敏感
		msgs, err = rt.db.GetPrivateConversation(userID, targetID)
		if notFoundErr(err) {
			msgs, err = rt.db.GetPrivateConversation(targetID, userID)
		}
	case "group":
		msgs, err = rt.db.GetGroupConversation(targetID)
	default:
		writeError(w, http.StatusBadRequest, "invalid chat_type: must be 'private' or 'group'")
		return
	}

	if err != nil {
		// “未找到/空结果/扫描到 NULL” → 返回 200 + 空数组
		if notFoundErr(err) {
			msgs = []models.Message{}
		} else {
			rt.baseLogger.WithError(err).Error("failed to get conversation")
			writeError(w, http.StatusInternalServerError, "Failed to get conversation")
			return
		}
	}

	// 关键：确保不是 nil，而是空切片，序列化为 []
	if msgs == nil {
		msgs = []models.Message{}
	}

	resp := map[string]interface{}{
		"code":    200,
		"message": "Conversation fetched successfully",
		"data": map[string]interface{}{
			"messages": msgs,
		},
	}
	writeJSON(w, http.StatusOK, resp)
}

// 识别“未找到/无记录/扫描到 NULL”类错误
func notFoundErr(err error) bool {
	if err == nil {
		return false
	}
	if err == sql.ErrNoRows {
		return true
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "not found") ||
		strings.Contains(s, "no rows") ||
		strings.Contains(s, "no conversation") ||
		strings.Contains(s, "record not found") ||
		strings.Contains(s, "empty result") ||
		strings.Contains(s, "converting null to string")
}
