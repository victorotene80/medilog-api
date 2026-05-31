package main

import (
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"

	"github.com/victorotene80/medilog-api/internal/bootstrap"
)

func main() {

	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found or failed to load:", err)
	}

	app, err := bootstrap.InitializeApp()
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:         ":8080",
		Handler:      app.Router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Println("Server running on :8080")
	log.Fatal(server.ListenAndServe())
}
