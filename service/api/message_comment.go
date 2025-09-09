package api

import (
	"net/http"
	"strings"

	"github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// CommentRequest supports both the official schema and legacy fields.
// Official (api.yaml):
//   { "type": "text"|"emoji", "content": "<text or emoji>" }
// Legacy (existing clients):
//   { "comment": "<text>", "emoji": "😊" }
type CommentRequest struct {
	Type    string `json:"type,omitempty"`    // "text" or "emoji"
	Content string `json:"content,omitempty"` // preferred content field
	Comment string `json:"comment,omitempty"` // legacy alias for text content
	Emoji   string `json:"emoji,omitempty"`   // legacy alias for emoji content
}

// commentMessage handles POST /messages/:id/comment
func (rt *_router) commentMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	messageID := strings.TrimSpace(ps.ByName("id"))
	if messageID == "" {
		rt.sendError(w, http.StatusBadRequest, "missing message id")
		return
	}
	if strings.TrimSpace(ctx.UserID) == "" {
		rt.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CommentRequest
	if err := readJSON(r, &req); err != nil {
		rt.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Normalize inputs: prefer Content; fallback to legacy fields
	content := strings.TrimSpace(req.Content)
	if content == "" {
		if v := strings.TrimSpace(req.Comment); v != "" {
			content = v
			if req.Type == "" {
				req.Type = "text"
			}
		}
	}
	if content == "" && strings.TrimSpace(req.Emoji) != "" {
		content = strings.TrimSpace(req.Emoji)
		if req.Type == "" {
			req.Type = "emoji"
		}
	}

	if content == "" {
		rt.sendError(w, http.StatusBadRequest, "content is required")
		return
	}

	// Persist comment (DB expects: messageID, userID, content)
	if err := rt.db.CommentMessage(messageID, ctx.UserID, content); err != nil {
		ctx.Logger.WithError(err).Error("failed to comment message")
		rt.sendError(w, http.StatusInternalServerError, "failed to comment message")
		return
	}

	// api.yaml uses 201 for created
	resp := map[string]interface{}{
		"code":    http.StatusCreated,
		"message": "Comment added successfully",
	}
	if err := writeJSON(w, http.StatusCreated, resp); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode comment response")
	}
}
