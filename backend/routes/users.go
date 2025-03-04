package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// setupUserRoutes configures all user-related routes
func setupUserRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		users.GET("", listUsers)
		users.POST("", createUser)
		users.GET("/:userId", getUser)
		users.PUT("/:userId", updateUser)
		users.DELETE("/:userId", deleteUser)
	}
}

func listUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GET /users endpoint hit"})
}

func createUser(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "POST /users endpoint hit"})
}

func getUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GET /users/:userId endpoint hit", "userId": c.Param("userId")})
}

func updateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "PUT /users/:userId endpoint hit", "userId": c.Param("userId")})
}

func deleteUser(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}
