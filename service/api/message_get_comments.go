package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// getMessages supports both:
// - Official: GET /messages?conversation_id=...&limit=...&before=ISO8601
// - Legacy:   GET /messages?chat_type=private|group&target_id=...
func (rt *_router) getMessages(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if ctx.UserID == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	q := r.URL.Query()
	convID := strings.TrimSpace(q.Get("conversation_id"))
	limitStr := strings.TrimSpace(q.Get("limit"))
	beforeStr := strings.TrimSpace(q.Get("before"))

	var (
		messages interface{}
		err      error
	)

	if convID != "" {
		// Official path: fetch by conversation_id (DB 侧你可按需改造；这里先用启发式路由)
		// 约定：以 u_/usr-/user- 开头 -> 私聊；以 g_/grp-/group- 开头 -> 群聊
		lc := strings.ToLower(convID)
		switch {
		case strings.HasPrefix(lc, "u_") || strings.HasPrefix(lc, "usr-") || strings.HasPrefix(lc, "user-"):
			var list []models.Message
			list, err = rt.db.GetPrivateConversation(ctx.UserID, convID)
			if err == nil && list == nil {
				list = make([]models.Message, 0)
			}
			messages = list
		case strings.HasPrefix(lc, "g_") || strings.HasPrefix(lc, "grp-") || strings.HasPrefix(lc, "group-"):
			var list []models.Message
			list, err = rt.db.GetGroupConversation(convID)
			if err == nil && list == nil {
				list = make([]models.Message, 0)
			}
			messages = list
		default:
			// 如果无法从 conversation_id 判别类型，这里先返回 400，前端可切回 legacy 参
			rt.sendError(w, http.StatusBadRequest, "conversation_id not recognized; use legacy chat_type/target_id if needed")
			return
		}

		// limit/before 目前仅做解析校验（如需分页可在 DB 层扩展）
		if limitStr != "" {
			if _, err2 := strconv.Atoi(limitStr); err2 != nil {
				rt.sendError(w, http.StatusBadRequest, "invalid limit")
				return
			}
		}
		if beforeStr != "" {
			if _, err2 := time.Parse(time.RFC3339, beforeStr); err2 != nil {
				rt.sendError(w, http.StatusBadRequest, "invalid before (expect RFC3339)")
				return
			}
		}
	} else {
		// Legacy path: chat_type + target_id
		chatType := strings.TrimSpace(q.Get("chat_type"))
		targetID := strings.TrimSpace(q.Get("target_id"))
		if chatType == "" || targetID == "" {
			rt.sendError(w, http.StatusBadRequest, "conversation_id or (chat_type & target_id) is required")
			return
		}
		switch chatType {
		case "private":
			var list []models.Message
			list, err = rt.db.GetPrivateConversation(ctx.UserID, targetID)
			if err == nil && list == nil {
				list = make([]models.Message, 0)
			}
			messages = list
		case "group":
			var list []models.Message
			list, err = rt.db.GetGroupConversation(targetID)
			if err == nil && list == nil {
				list = make([]models.Message, 0)
			}
			messages = list
		default:
			rt.sendError(w, http.StatusBadRequest, "chat_type must be 'private' or 'group'")
			return
		}
	}

	if err != nil {
		ctx.Logger.WithError(err).Error("failed to get conversation")
		rt.sendError(w, http.StatusInternalServerError, "Failed to fetch conversation")
		return
	}

	// OpenAPI：MessageResourceEnvelopeCollection -> { code, message, data: { messages, pagination } }
	// 先返回 messages，pagination 可后续接入
	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Messages retrieved successfully",
		"data": map[string]interface{}{
			"messages": messages,
		},
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode conversation response")
	}
}
