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

// needsRotation checks if the current log file needs rotation
func (l *Logger) needsRotation() bool {
	if l.file == nil || l.rotationConfig == nil {
		return false
	}

	stat, err := l.file.Stat()
	if err != nil {
		return false
	}

	return stat.Size() >= l.rotationConfig.MaxSize
}

// rotate performs log file rotation
func (l *Logger) rotate() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil || l.logPath == "" {
		return nil
	}

	// Close current file
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("failed to close log file: %w", err)
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.%s", l.logPath, timestamp)

	// Rename current log file to backup
	if err := os.Rename(l.logPath, backupPath); err != nil {
		return fmt.Errorf("failed to rotate log file: %w", err)
	}

	// Compress if enabled
	if l.rotationConfig.Compress {
		if err := compressFile(backupPath); err != nil {
			// Log error but don't fail rotation
			fmt.Fprintf(os.Stderr, "Failed to compress log file: %v\n", err)
		} else {
			// Remove uncompressed file after successful compression
			os.Remove(backupPath)
			backupPath += ".gz"
		}
	}

	// Clean up old backups
	if err := l.cleanupOldBackups(); err != nil {
		// Log error but don't fail rotation
		fmt.Fprintf(os.Stderr, "Failed to cleanup old backups: %v\n", err)
	}

	// Open new log file
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

// cleanupOldBackups removes old backup files based on MaxBackups and MaxAge
func (l *Logger) cleanupOldBackups() error {
	if l.logPath == "" || l.rotationConfig == nil {
		return nil
	}

	dir := filepath.Dir(l.logPath)
	base := filepath.Base(l.logPath)

	// Find all backup files
	var backups []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		if info.IsDir() {
			return nil
		}

		name := filepath.Base(path)
		// Match backup files (with or without .gz extension)
		if strings.HasPrefix(name, base+".") && 
		   (strings.Contains(name, "-") || strings.HasSuffix(name, ".gz")) {
			backups = append(backups, path)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Sort backups by modification time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		iStat, _ := os.Stat(backups[i])
		jStat, _ := os.Stat(backups[j])
		if iStat == nil || jStat == nil {
			return false
		}
		return iStat.ModTime().After(jStat.ModTime())
	})

	// Remove backups exceeding MaxBackups
	if l.rotationConfig.MaxBackups > 0 && len(backups) > l.rotationConfig.MaxBackups {
		for _, backup := range backups[l.rotationConfig.MaxBackups:] {
			os.Remove(backup)
		}
	}

	// Remove backups older than MaxAge
	if l.rotationConfig.MaxAge > 0 {
		cutoff := time.Now().AddDate(0, 0, -l.rotationConfig.MaxAge)
		for _, backup := range backups {
			stat, err := os.Stat(backup)
			if err != nil {
				continue
			}
			if stat.ModTime().Before(cutoff) {
				os.Remove(backup)
			}
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

// checkRotation checks if rotation is needed and performs it
func (l *Logger) checkRotation() {
	if l.needsRotation() {
		if err := l.rotate(); err != nil {
			fmt.Fprintf(os.Stderr, "Log rotation failed: %v\n", err)
		}
	}
}