package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	// Pre-initialize with a default logger to avoid race conditions
	globalLogger = &Logger{
		stdLogger: log.New(os.Stdout, "", log.LstdFlags),
	}
	loggerMu sync.RWMutex
	initOnce sync.Once
)

// Logger handles application logging
type Logger struct {
	file       *os.File
	fileLogger *log.Logger
	stdLogger  *log.Logger
	mu         sync.Mutex
}

// Init initializes the logger with the given log file path
func Init(logFilePath string) error {
	var err error
	initOnce.Do(func() {
		var newL *Logger
		newL, err = newLogger(logFilePath)
		if err == nil {
			loggerMu.Lock()
			globalLogger = newL
			loggerMu.Unlock()
		}
	})
	return err
}

// newLogger creates a new logger instance
func newLogger(logFilePath string) (*Logger, error) {
	l := &Logger{
		stdLogger: log.New(os.Stdout, "", log.LstdFlags),
	}

	if logFilePath != "" {
		// Ensure directory exists
		logDir := filepath.Dir(logFilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Open log file
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}

		l.file = file
		// Create multi-writer to write to both stdout and file
		multiWriter := io.MultiWriter(os.Stdout, file)
		l.fileLogger = log.New(multiWriter, "", log.LstdFlags)
	}

	return l, nil
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	return globalLogger
}

// log is the internal logging method
func (l *Logger) log(level, format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	msg := fmt.Sprintf(level+" "+format, v...)
	if l.fileLogger != nil {
		l.fileLogger.Println(msg)
	} else {
		l.stdLogger.Println(msg)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	l.log("[INFO]", format, v...)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.log("[ERROR]", format, v...)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	l.log("[DEBUG]", format, v...)
}

// LogStartup logs application startup information
func (l *Logger) LogStartup(port int, dbPath string) {
	l.Info("=== SQLite OTEL Collector Starting ===")
	l.Info("Version: v0.5")
	l.Info("Port: %d", port)
	l.Info("Database: %s", dbPath)
	l.Info("Started at: %s", time.Now().Format(time.RFC3339))
}

// LogShutdown logs application shutdown
func (l *Logger) LogShutdown() {
	l.Info("=== SQLite OTEL Collector Shutting Down ===")
	l.Info("Stopped at: %s", time.Now().Format(time.RFC3339))
}

// Close closes the log file if open
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Close closes the global logger
func Close() error {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	
	if globalLogger != nil {
		return globalLogger.Close()
	}
	return nil
}

// Info logs an info message using the global logger
func Info(format string, v ...interface{}) {
	GetLogger().Info(format, v...)
}

// Error logs an error message using the global logger
func Error(format string, v ...interface{}) {
	GetLogger().Error(format, v...)
}

// Debug logs a debug message using the global logger
func Debug(format string, v ...interface{}) {
	GetLogger().Debug(format, v...)
}