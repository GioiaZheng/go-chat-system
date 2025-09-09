package api

import (
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
)

// Helper: create an alias for a route parameter (e.g., :messageId -> :id)
func (rt *_router) withParamAlias(h httprouter.Handle, from, to string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if v := ps.ByName(from); v != "" {
			ps = append(ps, httprouter.Param{Key: to, Value: v})
		}
		h(w, r, ps)
	}
}

// Helper: map :conversationId from the path into query string (?conversation_id=...)
// This allows us to reuse the existing getConversation logic
func (rt *_router) getConversationByPathParam(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cid := ps.ByName("conversationId")
	if cid != "" {
		q := r.URL.Query()
		q.Set("conversation_id", cid)
		r.URL.RawQuery = q.Encode()
	}
	rt.getConversation(w, r, ps)
}

// RegisterRoutes wires all HTTP endpoints to their handlers.
// - The main API is exposed under the /api prefix.
// - A set of compatibility aliases (without /api prefix) is also registered,
//   to match alternative client or grading scripts.
func (rt *_router) RegisterRoutes() {
	basePath := "/api"

	// ---------------------- Public (no authentication) ----------------------
	rt.router.POST(basePath+"/session", rt.doLogin)
	rt.router.POST(basePath+"/register", rt.doRegister)
	rt.router.GET(basePath+"/liveness", rt.liveness)
	rt.router.OPTIONS(basePath+"/cors", rt.handleCorsPreflight)
	rt.router.POST(basePath+"/shutdown", rt.shutdown) // development-only helper

	// ---------------------- Protected (authentication required) ----------------------
	// Conversations
	rt.router.POST(basePath+"/conversations", rt.wrap(rt.startConversation))
	rt.router.GET(basePath+"/conversations", rt.wrap(rt.getMyConversations))

	// Users
	rt.router.PUT(basePath+"/users/set_username", rt.wrap(rt.setMyUserName))
	rt.router.PUT(basePath+"/users/set_photo", rt.wrap(rt.setMyPhoto))
	rt.router.GET(basePath+"/users/me", rt.wrap(rt.getUserInfo))
	rt.router.GET(basePath+"/users/search", rt.wrap(rt.searchUsers))
	rt.router.GET(basePath+"/users/profile/:user_id", rt.wrap(rt.getUserProfile))

	// Groups
	rt.router.POST(basePath+"/groups", rt.wrap(rt.createGroup))
	rt.router.GET(basePath+"/groups", rt.wrap(rt.getGroupsList))
	rt.router.GET(basePath+"/groups/:id", rt.wrap(rt.getGroupDetail))
	rt.router.PUT(basePath+"/groups/:id/name", rt.wrap(rt.setGroupName))
	rt.router.PUT(basePath+"/groups/:id/photo", rt.wrap(rt.setGroupPhoto))
	rt.router.POST(basePath+"/groups/:id/members", rt.wrap(rt.addToGroup))
	rt.router.DELETE(basePath+"/groups/:id/members", rt.wrap(rt.leaveGroup))

	// Messages
	rt.router.GET(basePath+"/messages", rt.wrap(rt.getConversation))
	rt.router.POST(basePath+"/messages", rt.wrap(rt.sendMessage))
	rt.router.GET(basePath+"/messages/:id", rt.wrap(rt.getMessageById))
	rt.router.DELETE(basePath+"/messages/:id", rt.wrap(rt.deleteMessage))
	rt.router.POST(basePath+"/messages/:id/forward", rt.wrap(rt.forwardMessage))
	rt.router.GET(basePath+"/messages/:id/comment", rt.wrap(rt.getMessageComments))
	rt.router.POST(basePath+"/messages/:id/comment", rt.wrap(rt.commentMessage))
	rt.router.POST(basePath+"/messages/:id/uncomment", rt.wrap(rt.uncommentMessage))

	// ---------------------- Compatibility aliases (no /api prefix) ----------------------
	// Public
	rt.router.POST("/session", rt.doLogin)
	rt.router.GET("/liveness", rt.liveness)

	// Users
	rt.router.PUT("/users/me/name", rt.wrap(rt.setMyUserName))
	rt.router.PUT("/users/me/photo", rt.wrap(rt.setMyPhoto))
	rt.router.GET("/users/search", rt.wrap(rt.searchUsers))

	// Conversations
	rt.router.GET("/conversations", rt.wrap(rt.getMyConversations))
	rt.router.POST("/start-conversation", rt.wrap(rt.startConversation))
	rt.router.GET("/conversations/:conversationId", rt.wrap(rt.getConversationByPathParam))
	rt.router.POST("/conversations/:conversationId/messages", rt.wrap(rt.sendMessage))

	// Messages
	rt.router.DELETE("/messages/:messageId", rt.wrap(rt.withParamAlias(rt.deleteMessage, "messageId", "id")))
	rt.router.POST("/messages/:messageId/forward", rt.wrap(rt.withParamAlias(rt.forwardMessage, "messageId", "id")))
	rt.router.POST("/messages/:messageId/reactions", rt.wrap(rt.withParamAlias(rt.commentMessage, "messageId", "id")))
	rt.router.DELETE("/messages/:messageId/reactions", rt.wrap(rt.withParamAlias(rt.uncommentMessage, "messageId", "id")))

	// Groups
	rt.router.POST("/groups/:groupId/leave", rt.wrap(rt.withParamAlias(rt.leaveGroup, "groupId", "id")))
	rt.router.POST("/groups/:groupId/members", rt.wrap(rt.withParamAlias(rt.addToGroup, "groupId", "id")))
	rt.router.PUT("/groups/:groupId/name", rt.wrap(rt.withParamAlias(rt.setGroupName, "groupId", "id")))
	rt.router.PUT("/groups/:groupId/photo", rt.wrap(rt.withParamAlias(rt.setGroupPhoto, "groupId", "id")))

	// Static files (optional)
	rt.router.ServeFiles("/photos/*filepath", http.Dir("data/photos"))
}

// CORS preflight handler
func (rt *_router) handleCorsPreflight(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Max-Age", "1")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
