package main

import (
	"log"
	"os"
	"path/filepath"
	"splitter/db"
	"splitter/routes"
)

func main() {
	// Create .data directory if it doesn't exist
	if err := os.MkdirAll(filepath.Join(".data"), 0755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	// Initialize database
	dbPath := filepath.Join(".data", "splitter.db")
	gormDB, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Get underlying SQL DB instance
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatal("Failed to get SQL DB instance:", err)
	}
	defer sqlDB.Close()

	// Setup and start the server
	router := routes.SetupRouter(gormDB)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
