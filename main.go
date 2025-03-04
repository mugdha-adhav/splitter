package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"splitter/db"

	"golang.org/x/crypto/acme/autocert"
)

var repository *db.Repository

func main() {
	// Initialize database
	dbPath := filepath.Join(".data", "splitter.db")
	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close()

	// Create repository instance
	repository = db.NewRepository(database)

	// Set up HTTP handlers
	setupRoutes()

	// Check environment
	if os.Getenv("ENV") == "local" {
		serveLocal()
	} else {
		serve()
	}
}

func setupRoutes() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world"))
	})

	http.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("bar"))
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("wip"))
	})
}

func serve() {
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("splitter.mriyam.com"),
		Cache:      autocert.DirCache("certs"),
	}

	server := &http.Server{
		Addr: ":https",
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		},
	}

	go http.ListenAndServe(":http", certManager.HTTPHandler(nil))
	log.Printf("Starting production server on HTTPS")
	log.Fatal(server.ListenAndServeTLS("", ""))
}

func serveLocal() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting local server on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
