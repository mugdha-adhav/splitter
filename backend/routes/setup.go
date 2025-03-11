package routes

import (
	"splitter/db"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var repository *db.Repository

func SetupRouter(database *gorm.DB) *gin.Engine {
	repository = db.NewRepository(database)
	router := gin.Default()

	// Enable CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := router.Group("/api")
	{
		// Public routes (no authentication required)
		setupAuthRoutes(api)

		// Protected routes (require authentication)
		protected := api.Group("")
		protected.Use(AuthMiddleware())
		{
			setupUserRoutes(protected)
			setupExpenseRoutes(protected)
		}
	}

	return router
}
