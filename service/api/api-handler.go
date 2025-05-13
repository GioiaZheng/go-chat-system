package api

// RegisterRoutes sets up all API routes.
func (rt *_router) RegisterRoutes() {
	// Authentication
	rt.router.POST("/api/v1/session", rt.wrap(rt.doLogin))
	rt.router.POST("/api/v1/register", rt.wrap(rt.doRegister))

	// User Management
	rt.router.PUT("/api/v1/users/set_username", rt.wrap(rt.setMyUserName))
	rt.router.PUT("/api/v1/users/set_photo", rt.wrap(rt.setMyPhoto))
	rt.router.GET("/api/v1/users/info", rt.wrap(rt.getUserInfo))
	rt.router.GET("/api/v1/users/search", rt.wrap(rt.searchUsersHandler))
	rt.router.GET("/api/v1/users/:userId/profile", rt.wrap(rt.getUserProfile))

	// Friends
	rt.router.GET("/api/v1/friends", rt.wrap(rt.getUserFriends))
	rt.router.POST("/api/v1/friends/add", rt.wrap(rt.addFriend))
	rt.router.GET("/api/v1/users/:userId/friends", rt.wrap(rt.getFriendsList))

	// Groups
	rt.router.POST("/api/v1/groups", rt.wrap(rt.createGroup))
	rt.router.GET("/api/v1/groups", rt.wrap(rt.getGroupsList))
	rt.router.GET("/api/v1/groups/:id", rt.wrap(rt.getGroupDetail))
	rt.router.PUT("/api/v1/groups/:id/name", rt.wrap(rt.setGroupName))
	rt.router.PUT("/api/v1/groups/:id/photo", rt.wrap(rt.setGroupPhoto))
	rt.router.POST("/api/v1/groups/:id/members", rt.wrap(rt.addToGroup))
	rt.router.DELETE("/api/v1/groups/:id/members", rt.wrap(rt.leaveGroup))

	// Messages
	rt.router.GET("/api/v1/messages", rt.wrap(rt.getConversation))
	rt.router.POST("/api/v1/messages", rt.wrap(rt.sendMessage))
	rt.router.GET("/api/v1/messages/:id", rt.wrap(rt.getMessageById))
	rt.router.DELETE("/api/v1/messages/:id", rt.wrap(rt.deleteMessage))
	rt.router.POST("/api/v1/messages/:id/forward", rt.wrap(rt.forwardMessage))
	rt.router.POST("/api/v1/messages/:id/comment", rt.wrap(rt.commentMessage))
	rt.router.POST("/api/v1/messages/:id/uncomment", rt.wrap(rt.uncommentMessage))

	// Conversation
	rt.router.POST("/api/v1/start_conversation", rt.wrap(rt.startConversation))

	// Health Check
	rt.router.GET("/api/v1/liveness", rt.wrap(rt.liveness))
}
