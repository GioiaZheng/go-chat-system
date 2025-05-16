package api

func (rt *_router) RegisterRoutes() {
	basePath := "/api/v1"
	// 1. 首先注册所有完全静态的路由
	rt.router.POST(basePath+"/session", rt.wrap(rt.doLogin))
	rt.router.POST(basePath+"/register", rt.wrap(rt.doRegister))
	// 直接注册 liveness，不需要 wrap
	rt.router.GET(basePath+"/liveness", rt.liveness)
	rt.router.POST(basePath+"/start_conversation", rt.wrap(rt.startConversation))
	rt.router.GET(basePath+"/conversations", rt.wrap(rt.getMyConversations))

	// Users 相关静态路由
	rt.router.PUT(basePath+"/users/set_username", rt.wrap(rt.setMyUserName))
	rt.router.PUT(basePath+"/users/set_photo", rt.wrap(rt.setMyPhoto))
	rt.router.GET(basePath+"/users/info", rt.wrap(rt.getUserInfo))
	rt.router.GET(basePath+"/users/search", rt.wrap(rt.searchUsers))

	// Friends 相关静态路由
	rt.router.GET(basePath+"/friends", rt.wrap(rt.getUserFriends))
	rt.router.POST(basePath+"/friends/add", rt.wrap(rt.addFriend))

	// Groups 相关静态路由
	rt.router.POST(basePath+"/groups", rt.wrap(rt.createGroup))
	rt.router.GET(basePath+"/groups", rt.wrap(rt.getGroupsList))

	// Messages 相关静态路由
	rt.router.GET(basePath+"/messages", rt.wrap(rt.getConversation))
	rt.router.POST(basePath+"/messages", rt.wrap(rt.sendMessage))

	// Users 参数路由
	rt.router.GET(basePath+"/users/profile/:userId", rt.wrap(rt.getUserProfile)) // 修改了路径结构

	// Friends 参数路由
	rt.router.GET(basePath+"/friends/list/:userId", rt.wrap(rt.getFriendsList)) // 修改了路径结构

	// Groups 参数路由
	rt.router.GET(basePath+"/groups/:groupId", rt.wrap(rt.getGroupDetail))
	rt.router.PUT(basePath+"/groups/:groupId/name", rt.wrap(rt.setGroupName))
	rt.router.PUT(basePath+"/groups/:groupId/photo", rt.wrap(rt.setGroupPhoto))
	rt.router.POST(basePath+"/groups/:groupId/members", rt.wrap(rt.addToGroup))
	rt.router.DELETE(basePath+"/groups/:groupId/members", rt.wrap(rt.leaveGroup))

	// Messages 参数路由
	rt.router.GET(basePath+"/messages/:messageId", rt.wrap(rt.getMessageById))
	rt.router.DELETE(basePath+"/messages/:messageId", rt.wrap(rt.deleteMessage))
	rt.router.POST(basePath+"/messages/:messageId/forward", rt.wrap(rt.forwardMessage))
	rt.router.POST(basePath+"/messages/:messageId/comment", rt.wrap(rt.commentMessage))
	rt.router.POST(basePath+"/messages/:messageId/uncomment", rt.wrap(rt.uncommentMessage))
}
