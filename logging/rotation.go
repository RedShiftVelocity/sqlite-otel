package logging

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// RotationConfig defines log rotation parameters
type RotationConfig struct {
	MaxSize    int64 // Maximum file size in bytes before rotation (default: 100MB)
	MaxBackups int   // Maximum number of backup files to keep (default: 7)
	MaxAge     int   // Maximum age in days to keep backup files (default: 30)
	Compress   bool  // Whether to compress rotated files (default: true)
}

// DefaultRotationConfig returns default rotation configuration
func DefaultRotationConfig() *RotationConfig {
	return &RotationConfig{
		MaxSize:    100 * 1024 * 1024, // 100MB
		MaxBackups: 7,
		MaxAge:     30,
		Compress:   true,
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

	// Perform slow operations in the background
	go l.compressAndCleanup(backupPath)

	return nil
}

// compressAndCleanup runs compression and cleanup in the background
func (l *Logger) compressAndCleanup(backupPath string) {
	if l.rotationConfig != nil && l.rotationConfig.Compress {
		if err := compressFile(backupPath); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to compress log file %s: %v\n", backupPath, err)
		} else {
			// Remove uncompressed file after successful compression
			if err := os.Remove(backupPath); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to remove uncompressed log file %s: %v\n", backupPath, err)
			}
		}
	}

	// Clean up old backups
	if err := l.cleanupOldBackups(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to cleanup old backups: %v\n", err)
	}
}

// backupFile represents a backup log file with parsed timestamp
type backupFile struct {
	path string
	ts   time.Time
}

// cleanupOldBackups removes old backup files based on MaxBackups and MaxAge
// This version is safe to run in a goroutine
func (l *Logger) cleanupOldBackups() error {
	if l.logPath == "" || l.rotationConfig == nil {
		return nil
	}

	dir := filepath.Dir(l.logPath)
	base := filepath.Base(l.logPath)
	expectedPrefix := base + "."

	// Read directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %w", err)
	}

	var backups []backupFile
	now := time.Now()

	// Find and parse backup files
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, expectedPrefix) {
			continue
		}

		// Parse timestamp from filename
		timestampPart := strings.TrimPrefix(name, expectedPrefix)
		if strings.HasSuffix(timestampPart, ".gz") {
			timestampPart = strings.TrimSuffix(timestampPart, ".gz")
		}

		// Try parsing with microseconds first
		ts, err := time.Parse("20060102-150405.000000", timestampPart)
		if err != nil {
			// Try without microseconds for backward compatibility
			if ts, err = time.Parse("20060102-150405", timestampPart); err != nil {
				continue // Not a valid backup file format
			}
		}
		backups = append(backups, backupFile{path: filepath.Join(dir, name), ts: ts})
	}

	// Sort backups by timestamp, newest first
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].ts.After(backups[j].ts)
	})

	// Determine which files to delete
	var toDelete []string

	// Apply MaxBackups policy
	if l.rotationConfig.MaxBackups > 0 && len(backups) > l.rotationConfig.MaxBackups {
		for i := l.rotationConfig.MaxBackups; i < len(backups); i++ {
			toDelete = append(toDelete, backups[i].path)
		}
	}

	// Apply MaxAge policy
	if l.rotationConfig.MaxAge > 0 {
		cutoff := now.AddDate(0, 0, -l.rotationConfig.MaxAge)
		for _, backup := range backups {
			if backup.ts.Before(cutoff) {
				// Check if already marked for deletion
				alreadyMarked := false
				for _, path := range toDelete {
					if path == backup.path {
						alreadyMarked = true
						break
					}
				}
				if !alreadyMarked {
					toDelete = append(toDelete, backup.path)
				}
			}
		}
	}

	// Delete identified files
	for _, path := range toDelete {
		if err := os.Remove(path); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove old backup %s: %v\n", path, err)
		}
	}

	return nil
}

// compressFile compresses a file using gzip
func compressFile(path string) error {
	source, err := os.Open(path)
	if err != nil {
		return err
	}
	defer source.Close()

	destPath := path + ".gz"
	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	gz := gzip.NewWriter(dest)
	defer gz.Close()

	// Copy file contents
	if _, err := io.Copy(gz, source); err != nil {
		os.Remove(destPath) // Clean up on error
		return err
	}

	return nil
}