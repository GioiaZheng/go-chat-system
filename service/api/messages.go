// messages.go serves message read/write endpoints and the DTO helpers that map
// internal models to the OpenAPI JSON shapes.
package api

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/GioiaZheng/Wasa_proj/service/models"
	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
)

// Section: DTOs (match OpenAPI JSON shapes; keep models.* as internal)

type MessageDTO struct {
	ID             string  `json:"id"`
	Content        string  `json:"content"`
	FileUrl        string  `json:"fileUrl,omitempty"`
	SenderID       string  `json:"senderId"`
	ConversationID string  `json:"conversationId,omitempty"`
	CreatedAt      string  `json:"createdAt"`
	Type           string  `json:"type,omitempty"`
	Status         string  `json:"status,omitempty"`
	ReplyToID      *string `json:"replyToId,omitempty"`
}

// toMessageDTO maps internal models.Message (snake_case tags) to public MessageDTO.
func toMessageDTO(m models.Message) MessageDTO {
	return MessageDTO{
		ID:             m.ID,
		Content:        m.Content,
		FileUrl:        m.FileURL,
		SenderID:       m.SenderID,
		ConversationID: m.ConversationID,
		CreatedAt:      m.CreatedAt,
		Type:           m.Type,
		Status:         m.Status,
		ReplyToID:      m.ReplyToID,
	}
}

// CommentDTO matches OpenAPI Comment: id, authorId, content, createdAt.
type CommentDTO struct {
	ID        string `json:"id"`
	AuthorID  string `json:"authorId"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

// messageRowToCommentDTO converts an internal models.Message row used as a comment
// into the outward-facing CommentDTO.
func messageRowToCommentDTO(m models.Message) CommentDTO {
	return CommentDTO{
		ID:        m.ID,
		AuthorID:  m.SenderID,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
}

// Section: Helpers

// newMsgID generates a random message ID (TEXT).
func newMsgID() (string, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

// parseLimit converts a query string into a safe page size.
func parseLimit(raw string, dflt, min, max int) int {
	if raw == "" {
		return dflt
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return dflt
	}
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

// timeBefore reports a < b assuming RFC3339 timestamps (lex compare is OK for full RFC3339Z).
func timeBefore(a, b string) bool { return a < b }
func timeAfter(a, b string) bool  { return a > b }

// Endpoint: POST /messages -> sendMessage (OpenAPI)

// sendMessageRequest matches the OpenAPI request body for sending a message.
type sendMessageRequest struct {
	ConversationID string  `json:"conversationId"`
	Content        string  `json:"content"`
	FileUrl        string  `json:"fileUrl,omitempty"`
	Type           string  `json:"type,omitempty"`
	ReplyToID      *string `json:"replyToId"`
}

// sendMessage sends a message to a conversation and returns MessageResourceEnvelope.
func (rt *_router) sendMessage(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	userID := strings.TrimSpace(ctx.UserID)
	if userID == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req sendMessageRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	req.ConversationID = strings.TrimSpace(req.ConversationID)
	req.Content = strings.TrimSpace(req.Content)
	req.FileUrl = strings.TrimSpace(req.FileUrl)

	if req.ConversationID == "" {
		rt.sendError(w, http.StatusBadRequest, "conversationId is required")
		return
	}

	if req.Content == "" && req.FileUrl == "" {
		rt.sendError(w, http.StatusBadRequest, "content or fileUrl is required")
		return
	}

	if req.FileUrl != "" && req.Type == "" {
		req.Type = "image"
	} else if req.Type == "" {
		req.Type = "text"
	}

	if req.Type == "image" && req.FileUrl == "" {
		rt.sendError(w, http.StatusBadRequest, "fileUrl is required for image messages")
		return
	}

	id, err := newMsgID()
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "Failed to generate message ID")
		return
	}

	// use RFC3339Nano to avoid same-second collisions in ordering/cursors
	now := time.Now().UTC().Format(time.RFC3339Nano)

	msg := models.Message{
		ID:             id,
		Content:        req.Content,
		FileURL:        req.FileUrl,
		SenderID:       userID,
		ConversationID: req.ConversationID,
		CreatedAt:      now,
		Type:           req.Type,
		Status:         "sent",
		ReplyToID:      req.ReplyToID,
	}
	if err := rt.db.SendMessageToConversation(msg); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "Failed to send message")
		return
	}

	// MessageResourceEnvelope
	resp := map[string]interface{}{
		"code":    http.StatusCreated,
		"message": "Message sent successfully",
		"data": map[string]interface{}{
			"resource": toMessageDTO(msg),
		},
	}
	_ = writeJSON(w, http.StatusCreated, resp)
}

// Endpoint: GET /messages -> getConversation (OpenAPI)

// getMessages returns a cursor-paginated slice for a conversation.
// Prefer database pagination; fall back to in-memory filtering to keep a 200 on failures.
func (rt *_router) getMessages(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	_ reqcontext.RequestContext,
) {
	q := r.URL.Query()
	convID := strings.TrimSpace(q.Get("conversationId"))
	if convID == "" {
		rt.sendError(w, http.StatusBadRequest, "conversationId is required")
		return
	}
	limit := parseLimit(q.Get("limit"), 20, 1, 100)
	before := strings.TrimSpace(q.Get("beforeCursor"))
	after := strings.TrimSpace(q.Get("afterCursor"))

	// 1) Primary path: database query per conversation
	items, err := rt.db.GetMessagesByConversation(convID, before, after, limit)
	if err != nil {
		// Log details for troubleshooting without changing the response shape
		rt.baseLogger.WithError(err).Error("getMessages: db.GetMessagesByConversation failed; falling back to in-memory filter")

		// 2) Fallback: load all messages into memory and filter to avoid 500s
		all, e2 := rt.db.GetAllMessages()
		if e2 != nil {
			rt.baseLogger.WithError(e2).Error("getMessages: db.GetAllMessages failed")
			rt.sendError(w, http.StatusInternalServerError, "Failed to fetch messages")
			return
		}
		// Keep only the messages belonging to the requested conversation
		buf := make([]models.Message, 0, len(all))
		for _, m := range all {
			if strings.TrimSpace(m.ConversationID) == convID {
				buf = append(buf, m)
			}
		}
		// Sort newest first; if timestamps tie, sort by id for determinism
		sort.Slice(buf, func(i, j int) bool {
			if buf[i].CreatedAt == buf[j].CreatedAt {
				return buf[i].ID > buf[j].ID
			}
			return buf[i].CreatedAt > buf[j].CreatedAt
		})
		// Apply cursors
		filtered := buf[:0]
		for _, m := range buf {
			if before != "" && !timeBefore(m.CreatedAt, before) {
				continue
			}
			if after != "" && !timeAfter(m.CreatedAt, after) {
				continue
			}
			filtered = append(filtered, m)
		}
		// Truncate to the requested page size
		if len(filtered) > limit {
			items = filtered[:limit]
		} else {
			items = filtered
		}
	}

	// Build cursors for the response
	var nextCursor, prevCursor *string
	if len(items) == limit {
		nc := items[len(items)-1].CreatedAt
		nextCursor = &nc
	}
	if after != "" && len(items) > 0 {
		pc := items[0].CreatedAt
		prevCursor = &pc
	}

	// DTO
	out := make([]MessageDTO, 0, len(items))
	for _, m := range items {
		out = append(out, toMessageDTO(m))
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Messages retrieved successfully",
		"data": map[string]interface{}{
			"messages":   out,
			"nextCursor": nextCursor,
			"prevCursor": prevCursor,
		},
	}
	_ = writeJSON(w, http.StatusOK, resp)
}

// Endpoint: GET /messages/{id} -> getMessageByID (OpenAPI)

func (rt *_router) getMessageByID(
	w http.ResponseWriter,
	_ *http.Request,
	ps httprouter.Params,
	_ reqcontext.RequestContext,
) {
	id := strings.TrimSpace(ps.ByName("id"))
	if id == "" {
		rt.sendError(w, http.StatusBadRequest, "Message ID is required")
		return
	}
	m, err := rt.db.GetMessageByID(id)
	if err != nil {
		rt.sendError(w, http.StatusNotFound, "Message not found")
		return
	}
	resp := map[string]interface{}{
		"code": http.StatusOK,
		"data": map[string]interface{}{
			"resource": toMessageDTO(m),
		},
	}
	_ = writeJSON(w, http.StatusOK, resp)
}

// Endpoint: POST /messages/{id}/forward -> forwardMessage (OpenAPI)

// forwardRequest matches OpenAPI: target conversationId only.
type forwardRequest struct {
	ConversationID string `json:"conversationId"`
}

// forwardMessage copies/forwards an existing message into another conversation.
// NOTE: DB expects (toUserID, toGroupID). Here we map conversationId => toGroupID.
// If you support private-conversation IDs, extend the mapping logic accordingly.
func (rt *_router) forwardMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	userID := strings.TrimSpace(ctx.UserID)
	if userID == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	msgID := strings.TrimSpace(ps.ByName("id"))
	if msgID == "" {
		rt.sendError(w, http.StatusBadRequest, "Message ID is required")
		return
	}

	var req forwardRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	req.ConversationID = strings.TrimSpace(req.ConversationID)
	if req.ConversationID == "" {
		rt.sendError(w, http.StatusBadRequest, "conversationId is required")
		return
	}

	// Map to existing DB API: treat target conversation as a "group" sink.
	// (If you have a real conversation->group lookup, add it here.)
	if err := rt.db.ForwardMessage(userID, msgID /*toUserID*/, "" /*toGroupID*/, req.ConversationID); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "Failed to forward message")
		return
	}

	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Message forwarded successfully",
	}
	_ = writeJSON(w, http.StatusOK, resp)
}

// Section: Comments (GET, POST, POST /uncomment)

// commentAddRequest matches OpenAPI CommentMessageRequest.
type commentAddRequest struct {
	Type    string `json:"type"`    // "text" or "emoji"
	Content string `json:"content"` // required
}

// getMessageComments returns the list of comments for a message.
// OpenAPI requires a plain object: { "comments": [...] } (no envelope).
func (rt *_router) getMessageComments(
	w http.ResponseWriter,
	_ *http.Request,
	ps httprouter.Params,
	_ reqcontext.RequestContext,
) {
	msgID := strings.TrimSpace(ps.ByName("id"))
	if msgID == "" {
		rt.sendError(w, http.StatusBadRequest, "Message ID is required")
		return
	}
	rows, err := rt.db.GetMessageComments(msgID)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "Failed to fetch message comments")
		return
	}
	out := make([]CommentDTO, 0, len(rows))
	for _, m := range rows {
		out = append(out, messageRowToCommentDTO(m))
	}

	resp := map[string]interface{}{
		"comments": out,
	}
	_ = writeJSON(w, http.StatusOK, resp)
}

func (rt *_router) commentMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	userID := strings.TrimSpace(ctx.UserID)
	if userID == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	msgID := strings.TrimSpace(ps.ByName("id"))
	if msgID == "" {
		rt.sendError(w, http.StatusBadRequest, "Message ID is required")
		return
	}

	var req commentAddRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		rt.sendError(w, http.StatusBadRequest, "content is required")
		return
	}

	// Default to text comments when the client does not provide a type.
	if req.Type == "" {
		req.Type = "text"
	}

	// Only allow text or emoji comments to avoid unexpected database writes.
	if req.Type != "text" && req.Type != "emoji" {
		rt.sendError(w, http.StatusBadRequest, "invalid type")
		return
	}

	// Persist the comment using the normalized type and sanitized content.
	if err := rt.db.CommentMessage(msgID, userID, req.Type, req.Content); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "Failed to add comment")
		return
	}

	resp := map[string]interface{}{
		"code":    http.StatusCreated,
		"message": "Comment added successfully",
	}
	_ = writeJSON(w, http.StatusCreated, resp)
}

func (rt *_router) uncommentMessage(
	w http.ResponseWriter,
	_ *http.Request,
	ps httprouter.Params,
	_ reqcontext.RequestContext,
) {
	msgID := strings.TrimSpace(ps.ByName("id"))
	if msgID == "" {
		rt.sendError(w, http.StatusBadRequest, "Message ID is required")
		return
	}
	if err := rt.db.UncommentMessage(msgID); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "Failed to remove comment(s)")
		return
	}
	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Comment removed successfully",
	}
	_ = writeJSON(w, http.StatusOK, resp)
}

// Endpoint: DELETE /messages/{id} -> deleteMessage (OpenAPI)

func (rt *_router) deleteMessage(
	w http.ResponseWriter,
	_ *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	userID := strings.TrimSpace(ctx.UserID)
	if userID == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	msgID := strings.TrimSpace(ps.ByName("id"))
	if msgID == "" {
		rt.sendError(w, http.StatusBadRequest, "Message ID is required")
		return
	}

	// Optional: verify ownership before DB call.
	ok, err := rt.db.IsMessageOwner(userID, msgID)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "Failed to verify ownership")
		return
	}
	if !ok {
		rt.sendError(w, http.StatusForbidden, "You are not the owner of this message")
		return
	}

	if err := rt.db.DeleteMessage(userID, msgID); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "Failed to delete message")
		return
	}
	resp := map[string]interface{}{
		"code":    http.StatusOK,
		"message": "Message deleted successfully",
	}
	_ = writeJSON(w, http.StatusOK, resp)
}

// POST /messages/upload
func (rt *_router) uploadMessageFile(
	w http.ResponseWriter,
	r *http.Request,
	_ httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	userID := strings.TrimSpace(ctx.UserID)
	if userID == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Read the uploaded file from multipart field "upload"
	file, header, err := r.FormFile("upload")
	if err != nil {
		rt.sendError(w, http.StatusBadRequest, "file required")
		return
	}
	defer file.Close()

	// Ensure the uploads directory exists
	baseDir := "./uploads"
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		rt.sendError(w, http.StatusInternalServerError, "Failed to prepare upload directory")
		return
	}

	// Persist the file under ./uploads/
	fname := "msg_" + uuid.Must(uuid.NewV4()).String() + "_" + header.Filename
	path := filepath.Join(baseDir, fname)

	out, err := os.Create(path)
	if err != nil {
		rt.sendError(w, http.StatusInternalServerError, "Failed to save file")
		return
	}
	defer out.Close()

	_, _ = io.Copy(out, file)

	// Return a public URL for the frontend to use as message content
	url := "/uploads/" + fname

	resp := map[string]interface{}{
		"code": http.StatusCreated,
		"data": map[string]string{
			"fileUrl":  url,
			"filename": header.Filename,
		},
	}
	_ = writeJSON(w, http.StatusCreated, resp)
}
