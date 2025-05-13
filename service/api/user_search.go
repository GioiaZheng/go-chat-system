package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type searchUserRequest struct {
	Query string `form:"q" binding:"required,min=1"`
}

type searchedUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Avatar   string `json:"avatar_url,omitempty"`
}

func (rt *_router) searchUsersHandler(c *gin.Context) {
	var req searchUserRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid query parameter"})
		return
	}

	userID, exists := getUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "Unauthorized"})
		return
	}

	results, err := rt.db.SearchUsers(userID, req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to search users"})
		return
	}

	var response []searchedUser
	for _, u := range results {
		response = append(response, searchedUser{
			ID:       u.ID,
			Username: u.Username,
			Name:     u.Name,
			Email:    u.Email,
			Avatar:   u.AvatarURL,
		})
	}

	c.JSON(http.StatusOK, response)
}

func registerUserSearchRoutes(r *_router, rg *gin.RouterGroup) {
	rg.GET("/users/search", r.searchUsersHandler)
}
