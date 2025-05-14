package api

func (rt *_router) RegisterRoutes() {
	// 1. 首先注册所有完全静态的路由
	rt.router.POST("/session", rt.wrap(rt.doLogin))
	rt.router.POST("/register", rt.wrap(rt.doRegister))
	rt.router.GET("/liveness", rt.wrap(rt.liveness))
	rt.router.POST("/start_conversation", rt.wrap(rt.startConversation))
	rt.router.GET("/conversations", rt.wrap(rt.getMyConversations))

	// Users 相关静态路由
	rt.router.PUT("/users/set_username", rt.wrap(rt.setMyUserName))
	rt.router.PUT("/users/set_photo", rt.wrap(rt.setMyPhoto))
	rt.router.GET("/users/info", rt.wrap(rt.getUserInfo))
	rt.router.GET("/users/search", rt.wrap(rt.searchUsers))

	// Friends 相关静态路由
	rt.router.GET("/friends", rt.wrap(rt.getUserFriends))
	rt.router.POST("/friends/add", rt.wrap(rt.addFriend))

	// Groups 相关静态路由
	rt.router.POST("/groups", rt.wrap(rt.createGroup))
	rt.router.GET("/groups", rt.wrap(rt.getGroupsList))

	// Messages 相关静态路由
	rt.router.GET("/messages", rt.wrap(rt.getConversation))
	rt.router.POST("/messages", rt.wrap(rt.sendMessage))

	// Users 参数路由
	rt.router.GET("/users/profile/:userId", rt.wrap(rt.getUserProfile)) // 修改了路径结构

	// Friends 参数路由
	rt.router.GET("/friends/list/:userId", rt.wrap(rt.getFriendsList)) // 修改了路径结构

	// Groups 参数路由
	rt.router.GET("/groups/:groupId", rt.wrap(rt.getGroupDetail))
	rt.router.PUT("/groups/:groupId/name", rt.wrap(rt.setGroupName))
	rt.router.PUT("/groups/:groupId/photo", rt.wrap(rt.setGroupPhoto))
	rt.router.POST("/groups/:groupId/members", rt.wrap(rt.addToGroup))
	rt.router.DELETE("/groups/:groupId/members", rt.wrap(rt.leaveGroup))

	// Messages 参数路由
	rt.router.GET("/messages/:messageId", rt.wrap(rt.getMessageById))
	rt.router.DELETE("/messages/:messageId", rt.wrap(rt.deleteMessage))
	rt.router.POST("/messages/:messageId/forward", rt.wrap(rt.forwardMessage))
	rt.router.POST("/messages/:messageId/comment", rt.wrap(rt.commentMessage))
	rt.router.POST("/messages/:messageId/uncomment", rt.wrap(rt.uncommentMessage))
}
