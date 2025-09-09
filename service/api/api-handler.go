package api

import (
	"net/http"

	reqcontext "github.com/GioiaZheng/Wasa_proj/service/reqcontext"
	"github.com/julienschmidt/httprouter"
)
// withParamAlias renames a path param (e.g., :messageId -> :id) then calls the inner handler.
func (rt *_router) withParamAlias(h httpRouterHandler, from, to string) httpRouterHandler {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
		if v := ps.ByName(from); v != "" {
			ps = append(ps, httprouter.Param{Key: to, Value: v})
		}
		h(w, r, ps, ctx)
	}
}

// GET /conversations/:conversationId -> reuse GET /messages?conversation_id=...
func (rt *_router) getConversationByPathParam(
	w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext,
) {
	if cid := ps.ByName("conversationId"); cid != "" {
		q := r.URL.Query()
		q.Set("conversation_id", cid)
		r.URL.RawQuery = q.Encode()
	}
	// Reuse the canonical messages fetch handler
	rt.getMessages(w, r, ps, ctx)
}

// POST /conversations/:conversationId/messages -> reuse POST /messages (via query)
func (rt *_router) sendMessageByPathParam(
	w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext,
) {
	if cid := ps.ByName("conversationId"); cid != "" {
		q := r.URL.Query()
		q.Set("conversation_id", cid)
		r.URL.RawQuery = q.Encode()
	}
	rt.sendMessage(w, r, ps, ctx)
}

// RegisterRoutes wires all endpoints to handlers.
// Keep it simple (no /api prefix), and avoid duplicate registrations.
// Only add compatibility aliases when the path is actually different.
func (rt *_router) RegisterRoutes() {
	// ------------- Public (no auth) -------------
	rt.router.POST("/session", rt.doLogin)
	rt.router.POST("/register", rt.doRegister)
	rt.router.GET("/liveness", rt.liveness)
	rt.router.OPTIONS("/cors", rt.handleCorsPreflight)
	rt.router.POST("/shutdown", rt.shutdown) // dev-only helper

	// ------------- Protected (auth required) -------------

	// Conversations (canonical)
	rt.router.POST("/conversations", rt.wrap(rt.startConversation))
	rt.router.GET("/conversations", rt.wrap(rt.getMyConversations))

	// Users (canonical)
	rt.router.PUT("/users/set_username", rt.wrap(rt.setMyUserName))
	rt.router.PUT("/users/set_photo", rt.wrap(rt.setMyPhoto))
	rt.router.GET("/users/me", rt.wrap(rt.getUserInfo))
	rt.router.GET("/users/search", rt.wrap(rt.searchUsers))
	rt.router.GET("/users/profile/:user_id", rt.wrap(rt.getUserProfile))

	// Groups (canonical)
	rt.router.POST("/groups", rt.wrap(rt.createGroup))
	rt.router.GET("/groups", rt.wrap(rt.getGroupsList))
	rt.router.GET("/groups/:id", rt.wrap(rt.getGroupDetail))
	rt.router.PUT("/groups/:id/name", rt.wrap(rt.setGroupName))
	rt.router.PUT("/groups/:id/photo", rt.wrap(rt.setGroupPhoto))
	rt.router.POST("/groups/:id/members", rt.wrap(rt.addToGroup))
	rt.router.DELETE("/groups/:id/members", rt.wrap(rt.leaveGroup))

	// Messages (canonical)
	rt.router.GET("/messages", rt.wrap(rt.getMessages))
	rt.router.POST("/messages", rt.wrap(rt.sendMessage))
	rt.router.GET("/messages/:id", rt.wrap(rt.getMessageById))
	rt.router.DELETE("/messages/:id", rt.wrap(rt.deleteMessage))
	rt.router.POST("/messages/:id/forward", rt.wrap(rt.forwardMessage))
	rt.router.GET("/messages/:id/comment", rt.wrap(rt.getMessageComments))
	rt.router.POST("/messages/:id/comment", rt.wrap(rt.commentMessage))
	rt.router.POST("/messages/:id/uncomment", rt.wrap(rt.uncommentMessage))

	// ------------- Compatibility aliases (only when path differs) -------------

	// Users: keep different aliases only
	rt.router.PUT("/users/me/name", rt.wrap(rt.setMyUserName))   // alias of set_username
	rt.router.PUT("/users/me/photo", rt.wrap(rt.setMyPhoto))     // alias of set_photo
	// (Do NOT re-register /users/search; it's identical to canonical one.)

	// Conversations: add only different paths
	rt.router.POST("/start-conversation", rt.wrap(rt.startConversation))                          // extra alias
	rt.router.GET("/conversations/:conversationId", rt.wrap(rt.getConversationByPathParam))       // maps to /messages?conversation_id=...
	rt.router.POST("/conversations/:conversationId/messages", rt.wrap(rt.sendMessageByPathParam)) // maps to POST /messages

	// Static files
	rt.router.ServeFiles("/photos/*filepath", http.Dir("data/photos"))
}

// CORS preflight handler for manual OPTIONS checks.
func (rt *_router) handleCorsPreflight(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Max-Age", "1")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
