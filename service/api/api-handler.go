// api-handler.go centralizes route registration so OpenAPI endpoints stay
// consistent and easy to audit in one place.
package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

/*
RegisterRoutes wires all HTTP endpoints to their handlers.

This router is kept 1:1 aligned with the OpenAPI spec:
  - No global "/api" prefix.
  - No CORS preflight helper endpoint (the spec removed /cors).
  - Only the endpoints required by the assignment are exposed here.
    (If you need dev-only helpers, put them in a separate file.)
*/
func (rt *_router) RegisterRoutes() {
	// Section: public routes that do not require authentication.

	// POST /session -> doLogin (returns 200 on success per spec)
	rt.router.POST("/session", rt.doLogin)

	// GET /liveness -> simple service liveness check
	rt.router.GET("/liveness", rt.liveness)
	// GET /healthz -> nginx /api/healthz proxy
	rt.router.GET("/healthz", rt.healthz)

	// Section: authenticated routes; every handler is wrapped to inject auth/context.

	// Conversations
	// POST /conversations -> startConversation
	rt.router.POST("/conversations", rt.wrap(rt.startConversation))
	// GET /conversations -> getMyConversations
	rt.router.GET("/conversations", rt.wrap(rt.getMyConversations))
	// DELETE /conversations/{id} -> deleteConversation
	rt.router.DELETE("/conversations/:id", rt.wrap(rt.deleteConversation))

	// Users
	// GET /users/me -> getUserInfo
	rt.router.GET("/users/me", rt.wrap(rt.getUserInfo))
	// PUT /users/me/name -> setMyUserName
	rt.router.PUT("/users/me/name", rt.wrap(rt.setMyUserName))
	// PUT /users/me/photo -> setMyPhoto
	rt.router.PUT("/users/me/photo", rt.wrap(rt.setMyPhoto))
	// GET /users/search -> searchUsers
	rt.router.GET("/users/search", rt.wrap(rt.searchUsers))
	// GET /users/profile/{userId} -> getUserProfile
	rt.router.GET("/users/profile/:userId", rt.wrap(rt.getUserProfile))

	// Groups
	// POST /groups -> createGroup
	rt.router.POST("/groups", rt.wrap(rt.createGroup))
	// GET /groups -> getGroupsList
	rt.router.GET("/groups", rt.wrap(rt.getGroupsList))
	// GET /groups/{id} -> getGroupDetail
	rt.router.GET("/groups/:id", rt.wrap(rt.getGroupDetail))
	// PUT /groups/{id}/name -> setGroupName
	rt.router.PUT("/groups/:id/name", rt.wrap(rt.setGroupName))
	// PUT /groups/{id}/photo -> setGroupPhoto (multipart/form-data)
	rt.router.PUT("/groups/:id/photo", rt.wrap(rt.setGroupPhoto))
	// POST /groups/{id}/members -> addToGroup
	rt.router.POST("/groups/:id/members", rt.wrap(rt.addToGroup))
	// DELETE /groups/{id}/members -> leaveGroup
	rt.router.DELETE("/groups/:id/members", rt.wrap(rt.leaveGroup))

	// Messages
	// GET /messages?conversationId=...&limit=...&beforeCursor=...&afterCursor=... -> getConversation
	// NOTE: the query key is "conversationId" (camelCase) to match the spec.
	rt.router.GET("/messages", rt.wrap(rt.getMessages))
	// POST /messages -> sendMessage
	rt.router.POST("/messages", rt.wrap(rt.sendMessage))
	// POST /message-attachments -> uploadMessageFile
	rt.router.POST("/message-attachments", rt.wrap(rt.uploadMessageFile))
	// GET /messages/{id} -> getMessageByID
	rt.router.GET("/messages/:id", rt.wrap(rt.getMessageByID))
	// DELETE /messages/{id} -> deleteMessage
	rt.router.DELETE("/messages/:id", rt.wrap(rt.deleteMessage))
	// POST /messages/{id}/forwards -> forwardMessage
	rt.router.POST("/messages/:id/forwards", rt.wrap(rt.forwardMessage))
	// GET /messages/{id}/comment -> getMessageComments
	rt.router.GET("/messages/:id/comment", rt.wrap(rt.getMessageComments))
	// POST /messages/{id}/comment -> commentMessage
	rt.router.POST("/messages/:id/comment", rt.wrap(rt.commentMessage))
	// DELETE /messages/{id}/comment -> uncommentMessage
	rt.router.DELETE("/messages/:id/comment", rt.wrap(rt.uncommentMessage))

	// Section: static files served from the local uploads directory.
	// The handler matches publicURL(...) outputs (e.g. /uploads/users/...).
	rt.router.ServeFiles("/uploads/*filepath", http.Dir("uploads"))
}

// healthz returns a simple OK JSON for health checks (for nginx /api/healthz proxy).
func (rt *_router) healthz(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}
