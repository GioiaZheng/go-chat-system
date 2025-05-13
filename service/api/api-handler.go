package api

import (
	"net/http"
)

func (rt *_router) Handler() http.Handler {
	h := rt

	// Authentication
	h.router.POST("/login", h.doLogin)
	h.router.POST("/register", h.doRegister)

	// User Management
	h.router.PUT("/users/set-username", h.wrap(h.setMyUserName))
	h.router.PUT("/user/set-photo", h.wrap(h.setMyPhoto))
	h.router.GET("/user/info", h.wrap(h.getUserInfo))
	h.router.GET("/users/search", h.wrap(h.searchUsersHandler))
	h.router.GET("/users/:userId/profile", h.wrap(h.getUserProfile))

	// Friends
	h.router.GET("/friends", h.wrap(h.getUserFriends))
	h.router.POST("/friends/add", h.wrap(h.addFriend))
	h.router.GET("/users/:userId/friends", h.wrap(h.getFriendsList))

	// Groups
	h.router.POST("/groups", h.wrap(h.createGroup))
	h.router.GET("/groups", h.wrap(h.getGroupsList))
	h.router.GET("/groups/:groupId", h.wrap(h.getGroupDetail))
	h.router.PUT("/groups/:groupId/name", h.wrap(h.setGroupName))
	h.router.PUT("/groups/:groupId/photo", h.wrap(h.setGroupPhoto))
	h.router.POST("/groups/:groupId/members", h.wrap(h.addToGroup))
	h.router.DELETE("/groups/:groupId/members", h.wrap(h.leaveGroup))

	// Messages
	h.router.GET("/messages", h.wrap(h.getConversation))
	h.router.POST("/messages", h.wrap(h.sendMessage))
	h.router.GET("/messages/:messageId", h.wrap(h.getMessageById))
	h.router.DELETE("/messages/:messageId", h.wrap(h.deleteMessage))
	h.router.POST("/messages/:messageId/forward", h.wrap(h.forwardMessage))
	h.router.POST("/messages/:messageId/comment", h.wrap(h.commentMessage))
	h.router.POST("/messages/:messageId/uncomment", h.wrap(h.uncommentMessage))

	// Conversation
	h.router.POST("/start-conversation", h.wrap(h.startConversation))

	// Health Check
	h.router.GET("/liveness", h.liveness)

	return h.router
}
