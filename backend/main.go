package main

import (
	"log"

	"example.com/backend/database"
	"example.com/backend/handlers"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db, err := database.New()
	if err != nil {
		log.Fatal(err)
	}

	// Create handler instance with db
	h := handlers.New(db)

	engine := gin.Default()

	// Add database to context middleware for JWT auth
	engine.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// the jwt middleware
	authMiddleware, err := jwt.New(handlers.InitParams())
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	// register middleware
	engine.Use(func(context *gin.Context) {
		errInit := authMiddleware.MiddlewareInit()
		if errInit != nil {
			log.Fatal("authMiddleware.MiddlewareInit() Error:" + errInit.Error())
		}
	})

	// setup CORS
	engine.Use(cors.Default())

	// Register routes with handler instance
	handlers.Register(engine, authMiddleware, h)

	engine.Run()
}
