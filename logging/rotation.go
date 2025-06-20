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

	// Compress if enabled
	if l.rotationConfig.Compress {
		if err := compressFile(backupPath); err != nil {
			// Log error but don't fail rotation
			fmt.Fprintf(os.Stderr, "Failed to compress log file: %v\n", err)
		} else {
			// Remove uncompressed file after successful compression
			if err := os.Remove(backupPath); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to remove uncompressed log file %s: %v\n", backupPath, err)
			}
			backupPath += ".gz"
		}
	}

	// Clean up old backups
	if err := l.cleanupOldBackupsLocked(); err != nil {
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

// cleanupOldBackupsLocked removes old backup files based on MaxBackups and MaxAge
// Must be called with l.mu held
func (l *Logger) cleanupOldBackupsLocked() error {
	if l.logPath == "" || l.rotationConfig == nil {
		return nil
	}

	dir := filepath.Dir(l.logPath)
	base := filepath.Base(l.logPath)
	expectedPrefix := base + "."

	// Find all backup files
	var backups []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking path %s during log cleanup: %v\n", path, err)
			return nil // Continue walking
		}

		if info.IsDir() {
			return nil
		}

		name := filepath.Base(path)
		if !strings.HasPrefix(name, expectedPrefix) {
			return nil
		}

		// Extract timestamp part
		timestampPart := strings.TrimPrefix(name, expectedPrefix)
		if strings.HasSuffix(timestampPart, ".gz") {
			timestampPart = strings.TrimSuffix(timestampPart, ".gz")
		}

		// Validate timestamp format (YYYYMMDD-HHMMSS or YYYYMMDD-HHMMSS.UUUUUU)
		if _, err := time.Parse("20060102-150405", timestampPart); err != nil {
			// Try with microseconds
			if _, err := time.Parse("20060102-150405.000000", timestampPart); err != nil {
				// Not a valid backup file
				return nil
			}
		}

		backups = append(backups, path)
		return nil
	})

	if err != nil {
		return err
	}

	// Sort backups by modification time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		iStat, errI := os.Stat(backups[i])
		jStat, errJ := os.Stat(backups[j])
		if errI != nil || errJ != nil {
			// If we can't stat one of the files, maintain current order
			if errI != nil {
				fmt.Fprintf(os.Stderr, "Error statting backup file %s: %v\n", backups[i], errI)
			}
			if errJ != nil {
				fmt.Fprintf(os.Stderr, "Error statting backup file %s: %v\n", backups[j], errJ)
			}
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
				fmt.Fprintf(os.Stderr, "Error statting backup file %s during age cleanup: %v\n", backup, err)
				continue
			}
			if stat.ModTime().Before(cutoff) {
				if err := os.Remove(backup); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to remove old backup %s: %v\n", backup, err)
				}
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

