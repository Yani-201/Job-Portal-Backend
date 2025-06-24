package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"job-portal-backend/api/router"
	"job-portal-backend/config"
)

func main() {
	// Load configuration
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	cfg := config.GetEnv()

	// Set Gin mode based on environment
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize MongoDB connection
	mongoClient, err := config.NewMongoClient()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer config.Disconnect(mongoClient)

	db := config.GetDatabase(mongoClient)

	// Initialize router with database connection
	appRouter := router.NewRouter(db)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: appRouter.SetupRoutes(),
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server is running on http://localhost:%s\n", cfg.Port)
		log.Printf("Environment: %s\n", cfg.Environment)
		log.Printf("Database: %s\n", cfg.DatabaseName)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait until the timeout
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
