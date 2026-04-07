package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sampath/signaling-server/internal/server"
)

func main() {
	// Parse command-line flags
	addr := flag.String("addr", getEnv("ADDR", ":8080"), "Server address")
	sessionTTL := flag.Duration("session-ttl", 1*time.Hour, "Session time-to-live")
	monitorInterval := flag.Duration("monitor-interval", 30*time.Second, "Health monitor interval")
	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("=== Signaling Server ===")
	log.Printf("Configuration:")
	log.Printf("  Address: %s", *addr)
	log.Printf("  Session TTL: %v", *sessionTTL)
	log.Printf("  Monitor Interval: %v", *monitorInterval)
	log.Println("========================")

	// Create and start server
	srv := server.NewServer(*addr, *sessionTTL, *monitorInterval)

	// Start server in a goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
