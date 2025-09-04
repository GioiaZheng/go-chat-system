package api

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (rt *_router) RegisterRoutes() {
	basePath := "/api/v1"

	// Public (no auth)
	rt.router.POST(basePath+"/session", rt.doLogin)
	rt.router.POST(basePath+"/register", rt.doRegister)
	rt.router.GET(basePath+"/liveness", rt.liveness)
	rt.router.OPTIONS(basePath+"/cors", rt.handleCorsPreflight)

	// Auth-required
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
	rt.router.GET(basePath+"/messages", rt.wrap(rt.getMessages))
	rt.router.POST(basePath+"/messages", rt.wrap(rt.sendMessage))
	rt.router.GET(basePath+"/messages/:id", rt.wrap(rt.getMessageById))
	rt.router.DELETE(basePath+"/messages/:id", rt.wrap(rt.deleteMessage))
	rt.router.POST(basePath+"/messages/:id/forward", rt.wrap(rt.forwardMessage))
	rt.router.GET(basePath+"/messages/:id/comment", rt.wrap(rt.getMessageComments)) // ← 放在这里
	rt.router.POST(basePath+"/messages/:id/comment", rt.wrap(rt.commentMessage))
	rt.router.POST(basePath+"/messages/:id/uncomment", rt.wrap(rt.uncommentMessage))
}

func (rt *_router) handleCorsPreflight(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Max-Age", "1")
	w.WriteHeader(http.StatusOK)
}
