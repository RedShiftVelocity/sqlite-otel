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
	"github.com/RedShiftVelocity/sqlite-otel/logging"
)

func main() {
	// Define command-line flags
	port := flag.Int("port", 4318, "Port to listen on (default: 4318, OTLP/HTTP standard)")
	
	// Determine default database path following XDG Base Directory specification
	defaultDBPath := getDefaultDBPath()
	dbPath := flag.String("db-path", defaultDBPath, "Path to SQLite database file (default: "+defaultDBPath+")")
	
	// Determine default log file path
	defaultLogPath := getDefaultLogPath()
	logFile := flag.String("log-file", defaultLogPath, "Path to log file for execution metadata (default: "+defaultLogPath+")")
	
	// Log rotation flag
	logMaxSize := flag.Int64("log-max-size", 100, "Maximum size in MB before log rotation (default: 100)")
	
	flag.Parse()

	// Initialize logging with rotation
	rotationConfig := &logging.RotationConfig{
		MaxSize: *logMaxSize * 1024 * 1024, // Convert MB to bytes
	}
	if err := logging.InitWithRotation(*logFile, rotationConfig); err != nil {
		log.Fatalf("Failed to initialize logging: %v", err)
	}
	defer logging.Close()

	if err := run(*port, *dbPath); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func run(port int, dbPath string) error {
	logger := logging.GetLogger()
	logger.LogStartup(port, dbPath)
	// Ensure directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		logger.Error("Failed to create database directory: %v", err)
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Initialize database
	if err := database.InitDB(dbPath); err != nil {
		logger.Error("Failed to initialize database: %v", err)
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer database.CloseDB()

	logger.Info("SQLite database initialized at: %s", dbPath)

	// Create a listener on specified port
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Error("Failed to create listener on port %d: %v", port, err)
		if port == 4318 {
			fmt.Fprintf(os.Stderr, "Port 4318 appears to be in use. Try:\n  %s -port 4319\n  %s -port 0  (for random port)\n", 
				os.Args[0], os.Args[0])
		}
		return fmt.Errorf("failed to create listener on port %d: %w", port, err)
	}
	
	// Get the actual port that was assigned
	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return fmt.Errorf("listener address is not TCP")
	}
	actualPort := tcpAddr.Port
	logger.Info("OTLP/HTTP receiver listening on port %d", actualPort)
	
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
		IdleTimeout:  120 * time.Second,
	}
	
	// Channel to listen for interrupt signals and server errors
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	errChan := make(chan error, 1)
	
	// Start server in a goroutine
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed: %v", err)
			errChan <- err
		}
		close(errChan)
	}()
	
	// Wait for interrupt signal or server error
	select {
	case err := <-errChan:
		return fmt.Errorf("server failed: %w", err)
	case <-sigChan:
		logger.Info("Received shutdown signal")
		logger.LogShutdown()
	}
	
	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Shutdown the server gracefully
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown error: %v", err)
		return fmt.Errorf("server shutdown error: %w", err)
	}
	
	logger.Info("Server stopped successfully")
	return nil
}

// getDefaultDBPath returns the default database path following XDG Base Directory specification
func getDefaultDBPath() string {
	// Detect if running in service mode (no home directory or systemd)
	_, err := os.UserHomeDir()
	isServiceMode := err != nil || os.Getenv("INVOCATION_ID") != ""
	
	if isServiceMode {
		// Running as a service, use system directory
		return "/var/lib/sqlite-otel-collector/otel-collector.db"
	}
	
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

// getDefaultLogPath returns the default log file path
func getDefaultLogPath() string {
	// Detect if running in service mode (no home directory or systemd)
	_, err := os.UserHomeDir()
	isServiceMode := err != nil || os.Getenv("INVOCATION_ID") != ""
	
	if isServiceMode {
		// Running as a service, use system directory
		return "/var/log/sqlite-otel-collector.log"
	}
	
	// For user mode, write to user's log directory
	// First check XDG_STATE_HOME (for logs/state)
	stateHome := os.Getenv("XDG_STATE_HOME")
	
	if stateHome == "" {
		// If XDG_STATE_HOME is not set, use ~/.local/state
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current directory if home directory can't be determined
			return "sqlite-otel-collector.log"
		}
		stateHome = filepath.Join(homeDir, ".local", "state")
	}
	
	// Create the sqlite-otel subdirectory path
	return filepath.Join(stateHome, "sqlite-otel", "execution.log")
}