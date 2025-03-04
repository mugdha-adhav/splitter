package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"splitter/db"
	"splitter/routes"

	"golang.org/x/crypto/acme/autocert"
)

var repository *db.Repository

func main() {
	// Create .data directory if it doesn't exist
	if err := os.MkdirAll(filepath.Join(".data"), 0755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	// Initialize database
	dbPath := filepath.Join(".data", "splitter.db")
	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close()

	// Create repository instance and initialize routes with it
	repository = db.NewRepository(database)
	routes.InitRepository(repository)

	// Set up router using the routes package
	router := routes.SetupRouter()

	// Check environment
	if os.Getenv("ENV") == "local" {
		serveLocal(router)
	} else {
		serve(router)
	}
}

func serve(router http.Handler) {
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("splitter.mriyam.com"),
		Cache:      autocert.DirCache("certs"),
	}

	server := &http.Server{
		Addr:    ":https",
		Handler: router,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		},
	}

	// Handle HTTP-01 challenge
	go http.ListenAndServe(":http", certManager.HTTPHandler(router))

	log.Printf("Starting production server on HTTPS")
	log.Fatal(server.ListenAndServeTLS("", ""))
}

func serveLocal(router http.Handler) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting local server on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
