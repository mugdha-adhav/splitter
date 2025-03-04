package routes

import (
	"splitter/db"

	"github.com/gin-gonic/gin"
)

var repository *db.Repository

// InitRepository initializes the repository for all routes
func InitRepository(r *db.Repository) {
	repository = r
}

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
