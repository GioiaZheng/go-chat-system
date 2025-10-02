package api

import (
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
)

//
// ────────────────────────────────────────────────────────────────────────────────
//  Helpers
// ────────────────────────────────────────────────────────────────────────────────
//

// newMsgID generates a random message ID (TEXT). DB schema should accept TEXT PKs.
func newMsgID() (string, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

// readCommentBody parses a minimal body with a "comment" field.
type commentBody struct {
	Comment string `json:"comment"`
}

// sendMessageBody is a flexible payload used by multiple send handlers.
// Fields are optional; each handler validates what it needs.
type sendMessageBody struct {
	Content        string `json:"content"`
	ToUserID       string `json:"to_user_id,omitempty"`      // for private messages
	GroupID        string `json:"group_id,omitempty"`        // (legacy) if needed
	ConversationID string `json:"conversation_id,omitempty"` // for conversation messages
	Type           string `json:"type,omitempty"`            // "text" by default
	Status         string `json:"status,omitempty"`          // "sent" by default
}

//
// ────────────────────────────────────────────────────────────────────────────────
//  SEND
// ────────────────────────────────────────────────────────────────────────────────
//

// sendPrivateMessage handles POST /messages/private
// Body: { "content": "...", "to_user_id": "..." }
func (rt *_router) sendPrivateMessage(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	sender := strings.TrimSpace(ctx.UserID)
	if sender == "" {
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body sendMessageBody
	if err := readJSON(r, &body); err != nil {
		rt.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	body.Content = strings.TrimSpace(body.Content)
	body.ToUserID = strings.TrimSpace(body.ToUserID)
	if body.Content == "" || body.ToUserID == "" {
		rt.sendError(w, http.StatusBadRequest, "content and to_user_id are required")
		return
	}
	if body.Type == "" {
		body.Type = "text"
	}
	if body.Status == "" {
		body.Status = "sent"
	}

	id, err := newMsgID()
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "cannot generate message id")
		return
	}

	msg := models.Message{
		ID:         id,
		Content:    body.Content,
		SenderID:   sender,
		ReceiverID: body.ToUserID,
		Type:       body.Type,
		Status:     body.Status,
	}
	if err := rt.db.SendPrivateMessage(msg); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to send private message")
		return
	}

	_ = writeJSON(w, http.StatusCreated, map[string]interface{}{
		"code":    http.StatusCreated,
		"message": "private message sent",
		"data":    msg,
	})
}

// sendMessageToConversation handles POST /conversations/:id/messages
// Body: { "content": "..." }
// Path: :id is the conversation ID.
func (rt *_router) sendMessageToConversation(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	sender := strings.TrimSpace(ctx.UserID)
	if sender == "" {
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	convID := strings.TrimSpace(ps.ByName("id"))
	if convID == "" {
		// fallback: allow body to provide conversation_id
		var peek sendMessageBody
		_ = readJSON(r, &peek)
		convID = strings.TrimSpace(peek.ConversationID)
	}
	if convID == "" {
		rt.sendError(w, http.StatusBadRequest, "conversation id is required")
		return
	}

	var body sendMessageBody
	if err := readJSON(r, &body); err != nil {
		rt.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	body.Content = strings.TrimSpace(body.Content)
	if body.Content == "" {
		rt.sendError(w, http.StatusBadRequest, "content is required")
		return
	}
	if body.Type == "" {
		body.Type = "text"
	}
	if body.Status == "" {
		body.Status = "sent"
	}

	id, err := newMsgID()
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "cannot generate message id")
		return
	}

	msg := models.Message{
		ID:             id,
		Content:        body.Content,
		SenderID:       sender,
		ConversationID: convID,
		Type:           body.Type,
		Status:         body.Status,
	}
	if err := rt.db.SendMessageToConversation(msg); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to send message to conversation")
		return
	}

	_ = writeJSON(w, http.StatusCreated, map[string]interface{}{
		"code":    http.StatusCreated,
		"message": "message sent to conversation",
		"data":    msg,
	})
}

// sendGroupMessage is kept as a compatibility alias that delegates to sendMessageToConversation.
// Some legacy routes may post to a group but actually rely on conversation_id in DB.
func (rt *_router) sendGroupMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	rt.sendMessageToConversation(w, r, ps, ctx)
}

// sendMessage is a generic alias for "send to conversation".
func (rt *_router) sendMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	rt.sendMessageToConversation(w, r, ps, ctx)
}

//
// ────────────────────────────────────────────────────────────────────────────────
//  READ
// ────────────────────────────────────────────────────────────────────────────────
//

// getMessages handles GET /messages (admin/test helper)
// Returns all messages using db.GetAllMessages() if available.
func (rt *_router) getMessages(
	w http.ResponseWriter,
	_ *http.Request,
	_ httprouter.Params,
	_ reqcontext.RequestContext,
) {
	msgs, err := rt.db.GetAllMessages()
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to fetch messages")
		return
	}
	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"data": msgs,
	})
}

// getPrivateConversation handles GET /messages/private/:userId
// Returns ordered messages between current user and the specified user.
func (rt *_router) getPrivateConversation(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	me := strings.TrimSpace(ctx.UserID)
	if me == "" {
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	other := strings.TrimSpace(ps.ByName("userId"))
	if other == "" {
		other = strings.TrimSpace(ps.ByName("id")) // fallback param name
	}
	if other == "" {
		rt.sendError(w, http.StatusBadRequest, "user id required")
		return
	}

	msgs, err := rt.db.GetPrivateConversation(me, other)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to fetch private conversation")
		return
	}
	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"data": msgs,
	})
}

// getGroupConversation handles GET /groups/:id/messages (or legacy path)
// Returns ordered messages in a group conversation (DB side may map by conversation).
func (rt *_router) getGroupConversation(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	groupID := strings.TrimSpace(ps.ByName("groupId"))
	if groupID == "" {
		groupID = strings.TrimSpace(ps.ByName("id"))
	}
	if groupID == "" {
		rt.sendError(w, http.StatusBadRequest, "group id required")
		return
	}

	msgs, err := rt.db.GetGroupConversation(groupID)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to fetch group conversation")
		return
	}
	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"data": msgs,
	})
}

// getConversation / getConversationMessages
// If your router expects a "getConversation" handler for /conversations/:id/messages,
// you can wire to this function name. Here we only return 501 when no DB method exists.
func (rt *_router) getConversation(
	w http.ResponseWriter,
	_ *http.Request,
	_ httprouter.Params,
	_ reqcontext.RequestContext,
) {
	_ = writeJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"code":    http.StatusNotImplemented,
		"message": "Get conversation by conversation_id is not implemented in DB interface",
	})
}
func (rt *_router) getConversationMessages(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	rt.getConversation(w, r, ps, ctx)
}

// getMyConversations handles GET /me/conversations
func (rt *_router) getMyConversations(
	w http.ResponseWriter,
	_ *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	me := strings.TrimSpace(ctx.UserID)
	if me == "" {
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	convs, err := rt.db.GetMyConversations(me)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to fetch conversations")
		return
	}
	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"data": convs,
	})
}

// getMessageByID handles GET /messages/:id
func (rt *_router) getMessageByID(
	w http.ResponseWriter,
	_ *http.Request,
	ps httprouter.Params,
	_ reqcontext.RequestContext,
) {
	id := strings.TrimSpace(ps.ByName("messageId"))
	if id == "" {
		id = strings.TrimSpace(ps.ByName("id"))
	}
	if id == "" {
		rt.sendError(w, http.StatusBadRequest, "message id required")
		return
	}
	msg, err := rt.db.GetMessageByID(id)
	if err != nil {
		rt.sendError(w, http.StatusNotFound, "message not found")
		return
	}
	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"data": msg,
	})
}

// getMessageById is a compatibility alias (lowercase 'd').
func (rt *_router) getMessageById(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	rt.getMessageByID(w, r, ps, ctx)
}

// getMessageComments handles GET /messages/:id/comments
func (rt *_router) getMessageComments(
	w http.ResponseWriter,
	_ *http.Request,
	ps httprouter.Params,
	_ reqcontext.RequestContext,
) {
	id := strings.TrimSpace(ps.ByName("messageId"))
	if id == "" {
		id = strings.TrimSpace(ps.ByName("id"))
	}
	if id == "" {
		rt.sendError(w, http.StatusBadRequest, "message id required")
		return
	}
	items, err := rt.db.GetMessageComments(id)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to fetch message comments")
		return
	}
	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"data": items,
	})
}

//
// ────────────────────────────────────────────────────────────────────────────────
//  UPDATE / COMMENT / FORWARD / DELETE
// ────────────────────────────────────────────────────────────────────────────────
//

// commentMessage handles POST /messages/:id/comments
func (rt *_router) commentMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	uid := strings.TrimSpace(ctx.UserID)
	if uid == "" {
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(ps.ByName("messageId"))
	if id == "" {
		id = strings.TrimSpace(ps.ByName("id"))
	}
	if id == "" {
		rt.sendError(w, http.StatusBadRequest, "message id required")
		return
	}

	var body commentBody
	if err := readJSON(r, &body); err != nil {
		rt.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	body.Comment = strings.TrimSpace(body.Comment)
	if body.Comment == "" {
		rt.sendError(w, http.StatusBadRequest, "comment is required")
		return
	}

	if err := rt.db.CommentMessage(id, uid, body.Comment); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to comment the message")
		return
	}
	_ = writeJSON(w, http.StatusCreated, map[string]interface{}{
		"code":    http.StatusCreated,
		"message": "comment added",
	})
}

// uncommentMessage handles DELETE /messages/:id/comments
func (rt *_router) uncommentMessage(
	w http.ResponseWriter,
	_ *http.Request,
	ps httprouter.Params,
	_ reqcontext.RequestContext,
) {
	id := strings.TrimSpace(ps.ByName("messageId"))
	if id == "" {
		id = strings.TrimSpace(ps.ByName("id"))
	}
	if id == "" {
		rt.sendError(w, http.StatusBadRequest, "message id required")
		return
	}

	if err := rt.db.UncommentMessage(id); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to delete comment(s)")
		return
	}
	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    http.StatusOK,
		"message": "comment(s) removed",
	})
}

// forwardMessage handles POST /messages/:id/forward
// Body: { "to_user_id": "...", "to_group_id": "..." }  (provide exactly one)
type forwardBody struct {
	ToUserID  string `json:"to_user_id,omitempty"`
	ToGroupID string `json:"to_group_id,omitempty"`
}

func (rt *_router) forwardMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	uid := strings.TrimSpace(ctx.UserID)
	if uid == "" {
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(ps.ByName("messageId"))
	if id == "" {
		id = strings.TrimSpace(ps.ByName("id"))
	}
	if id == "" {
		rt.sendError(w, http.StatusBadRequest, "message id required")
		return
	}

	var body forwardBody
	if err := readJSON(r, &body); err != nil {
		rt.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	body.ToUserID = strings.TrimSpace(body.ToUserID)
	body.ToGroupID = strings.TrimSpace(body.ToGroupID)
	if (body.ToUserID == "" && body.ToGroupID == "") || (body.ToUserID != "" && body.ToGroupID != "") {
		rt.sendError(w, http.StatusBadRequest, "provide either to_user_id or to_group_id")
		return
	}

	if err := rt.db.ForwardMessage(uid, id, body.ToUserID, body.ToGroupID); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to forward message")
		return
	}
	_ = writeJSON(w, http.StatusCreated, map[string]interface{}{
		"code":    http.StatusCreated,
		"message": "message forwarded",
	})
}

// deleteMessage handles DELETE /messages/:id
func (rt *_router) deleteMessage(
	w http.ResponseWriter,
	_ *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	uid := strings.TrimSpace(ctx.UserID)
	if uid == "" {
		rt.sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(ps.ByName("messageId"))
	if id == "" {
		id = strings.TrimSpace(ps.ByName("id"))
	}
	if id == "" {
		rt.sendError(w, http.StatusBadRequest, "message id required")
		return
	}

	// Optional: double-check ownership in API before hitting DB.
	ok, err := rt.db.IsMessageOwner(uid, id)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to verify ownership")
		return
	}
	if !ok {
		rt.sendError(w, http.StatusForbidden, "you are not the owner of this message")
		return
	}

	if err := rt.db.DeleteMessage(uid, id); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "failed to delete message")
		return
	}
	_ = writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    http.StatusOK,
		"message": "message deleted",
	})
}
