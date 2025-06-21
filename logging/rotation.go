package logging

import (
	"fmt"
	"io"
	"os"
	"time"
)

// RotationConfig defines log rotation parameters
type RotationConfig struct {
	MaxSize int64 // Maximum file size in bytes before rotation (default: 100MB)
}

// DefaultRotationConfig returns default rotation configuration
func DefaultRotationConfig() *RotationConfig {
	return &RotationConfig{
		MaxSize: 100 * 1024 * 1024, // 100MB
	}
}

// needsRotationLocked checks if the current log file needs rotation
// Must be called with l.mu held
func (l *Logger) needsRotationLocked() bool {
	if l.file == nil || l.rotationConfig == nil {
		return false
	}

	stat, err := l.file.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stat log file for rotation check: %v\n", err)
		return false
	}

	return stat.Size() >= l.rotationConfig.MaxSize
}

// rotateLocked performs log file rotation
// Must be called with l.mu held
func (l *Logger) rotateLocked() error {
	if l.file == nil || l.logPath == "" {
		return nil
	}

	// Close current file
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("failed to close log file: %w", err)
	}

	// Generate backup filename with timestamp and microseconds for uniqueness
	timestamp := time.Now().Format("20060102-150405.000000")
	backupPath := fmt.Sprintf("%s.%s", l.logPath, timestamp)

	// Rename current log file to backup
	if err := os.Rename(l.logPath, backupPath); err != nil {
		return fmt.Errorf("failed to rotate log file: %w", err)
	}

	// Open new log file immediately to unblock logging
	file, err := os.OpenFile(l.logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open new log file: %w", err)
	}

	l.file = file
	// Update multi-writer
	multiWriter := io.MultiWriter(os.Stdout, file)
	l.fileLogger.SetOutput(multiWriter)

	return nil
}