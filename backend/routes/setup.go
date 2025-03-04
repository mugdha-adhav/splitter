package routes

import (
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes all routes and returns the router instance
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// API versioning group
	v1 := router.Group("/api/v1")
	{
		setupUserRoutes(v1)
		setupExpenseRoutes(v1)
	}

	return router
}
