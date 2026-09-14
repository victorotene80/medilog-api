package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	_ "github.com/victorotene80/medilog-api/docs"
	"github.com/victorotene80/medilog-api/internal/bootstrap"
)

// @title			MediLog API
// @version		1.0
// @description	HTTP API for MediLog authentication, health records, medication tracking, AI conversations, drug verification, and reference data.
// @host			localhost:8080
// @schemes		http
// @BasePath		/api/v1
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Println("Server running on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received, starting graceful shutdown...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	app.Stop()
	app.Close()
	log.Println("Server stopped")
}
