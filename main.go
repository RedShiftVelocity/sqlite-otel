package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/RedShiftVelocity/sqlite-otel/database"
	"github.com/RedShiftVelocity/sqlite-otel/handlers"
)

func main() {
	// Define command-line flags
	port := flag.Int("port", 4318, "Port to listen on (default: 4318, OTLP/HTTP standard)")
	
	// Determine default database path following XDG Base Directory specification
	defaultDBPath := getDefaultDBPath()
	dbPath := flag.String("db-path", defaultDBPath, "Path to SQLite database file (default: "+defaultDBPath+")")
	flag.Parse()

	// Ensure directory exists
	dbDir := filepath.Dir(*dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	// Initialize database
	if err := database.InitDB(*dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	fmt.Printf("SQLite database initialized at: %s\n", *dbPath)

	// Create a listener on specified port
	address := fmt.Sprintf(":%d", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		if *port == 4318 {
			log.Fatalf("Failed to create listener on port %d: %v\nPort 4318 appears to be in use. Try:\n  %s -port 4319\n  %s -port 0  (for random port)", 
				*port, err, os.Args[0], os.Args[0])
		} else {
			log.Fatalf("Failed to create listener on port %d: %v", *port, err)
		}
	}
	
	// Get the actual port that was assigned
	actualPort := listener.Addr().(*net.TCPAddr).Port
	fmt.Printf("OTLP/HTTP receiver listening on port %d\n", actualPort)
	
	// Create HTTP mux and register OTLP endpoints
	mux := http.NewServeMux()
	
	// Register OTLP endpoints
	mux.HandleFunc("/v1/traces", handlers.HandleTraces)
	mux.HandleFunc("/v1/metrics", handlers.HandleMetrics)
	mux.HandleFunc("/v1/logs", handlers.HandleLogs)
	
	// Create HTTP server
	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	
	// Channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	// Channel to signal server shutdown completion
	done := make(chan bool, 1)
	
	// Start server in a goroutine
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	<-sigChan
	fmt.Println("\nShutting down server...")
	
	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Shutdown the server gracefully
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	
	close(done)
	fmt.Println("Server stopped")
}

// getDefaultDBPath returns the default database path following XDG Base Directory specification
func getDefaultDBPath() string {
	// First check XDG_DATA_HOME
	dataHome := os.Getenv("XDG_DATA_HOME")
	
	if dataHome == "" {
		// If XDG_DATA_HOME is not set, use ~/.local/share
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current directory if home directory can't be determined
			return "otel-collector.db"
		}
		dataHome = filepath.Join(homeDir, ".local", "share")
	}
	
	// Create the sqlite-otel subdirectory path
	return filepath.Join(dataHome, "sqlite-otel", "otel-collector.db")
}