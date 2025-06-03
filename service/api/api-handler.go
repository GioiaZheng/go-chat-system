package api

func (rt *_router) RegisterRoutes() {
	basePath := "/api/v1"
	// 1. First register all completely static routes
	// These don't need the wrap middleware
	rt.router.POST(basePath+"/session", rt.doLogin)
	rt.router.POST(basePath+"/register", rt.doRegister)
	rt.router.GET(basePath+"/liveness", rt.liveness)
	rt.router.OPTIONS(basePath+"/api", rt.handleCorsPreflight) // Added to match OpenAPI spec
	
	// Routes that need authentication (use wrap middleware)
	rt.router.POST(basePath+"/start_conversation", rt.wrap(rt.startConversation))
	rt.router.GET(basePath+"/conversations", rt.wrap(rt.getMyConversations))

	// Users related static routes
	rt.router.PUT(basePath+"/users/set_username", rt.wrap(rt.setMyUserName))
	rt.router.PUT(basePath+"/users/set_photo", rt.wrap(rt.setMyPhoto))
	rt.router.GET(basePath+"/users/info", rt.wrap(rt.getUserInfo))
	rt.router.GET(basePath+"/users/search", rt.wrap(rt.searchUsers))

	// Friends related static routes
	rt.router.GET(basePath+"/friends", rt.wrap(rt.getUserFriends))
	rt.router.POST(basePath+"/friends/add", rt.wrap(rt.addFriend))

	// Groups related static routes
	rt.router.POST(basePath+"/groups", rt.wrap(rt.createGroup))
	rt.router.GET(basePath+"/groups", rt.wrap(rt.getGroupsList))

	// Messages related static routes
	rt.router.GET(basePath+"/messages", rt.wrap(rt.getConversation))
	rt.router.POST(basePath+"/messages", rt.wrap(rt.sendMessage))

	// Parameterized routes (using {param} syntax to match OpenAPI spec)
	// Users parameter routes
	rt.router.GET(basePath+"/users/profile/{userId}", rt.wrap(rt.getUserProfile))

	// Friends parameter routes
	rt.router.GET(basePath+"/friends/list/{userId}", rt.wrap(rt.getFriendsList))

	// Groups parameter routes
	rt.router.GET(basePath+"/groups/{id}", rt.wrap(rt.getGroupDetail))
	rt.router.PUT(basePath+"/groups/{id}/name", rt.wrap(rt.setGroupName))
	rt.router.PUT(basePath+"/groups/{id}/photo", rt.wrap(rt.setGroupPhoto))
	rt.router.POST(basePath+"/groups/{id}/members", rt.wrap(rt.addToGroup))
	rt.router.DELETE(basePath+"/groups/{id}/members", rt.wrap(rt.leaveGroup))

	// Messages parameter routes
	rt.router.GET(basePath+"/messages/{id}", rt.wrap(rt.getMessageById))
	rt.router.DELETE(basePath+"/messages/{id}", rt.wrap(rt.deleteMessage))
	rt.router.POST(basePath+"/messages/{id}/forward", rt.wrap(rt.forwardMessage))
	rt.router.POST(basePath+"/messages/{id}/comment", rt.wrap(rt.commentMessage))
	rt.router.POST(basePath+"/messages/{id}/uncomment", rt.wrap(rt.uncommentMessage))
}