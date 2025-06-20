package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRotationConfig(t *testing.T) {
	config := DefaultRotationConfig()
	
	if config.MaxSize != 100*1024*1024 {
		t.Errorf("Expected MaxSize to be 104857600, got %d", config.MaxSize)
	}
	
	if config.MaxBackups != 7 {
		t.Errorf("Expected MaxBackups to be 7, got %d", config.MaxBackups)
	}
	
	if config.MaxAge != 30 {
		t.Errorf("Expected MaxAge to be 30, got %d", config.MaxAge)
	}
}

func TestLogRotation(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "log-rotation-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	logPath := filepath.Join(tmpDir, "test.log")
	
	// Create logger with small rotation size for testing
	config := &RotationConfig{
		MaxSize: 100, // 100 bytes for easy testing
	}
	
	logger, err := newLoggerWithRotation(logPath, config)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()
	
	// Write enough data to trigger rotation
	for i := 0; i < 20; i++ {
		logger.Info("Test log message number %d", i)
	}
	
	// Check that rotation occurred
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(entries) < 2 {
		t.Errorf("Expected at least 2 files after rotation, got %d", len(entries))
	}
	
	// Verify backup file exists
	foundBackup := false
	for _, entry := range entries {
		if entry.Name() != "test.log" && !entry.IsDir() {
			foundBackup = true
			break
		}
	}
	
	if !foundBackup {
		t.Error("No backup file found after rotation")
	}
}

func TestMaxBackupsRetention(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "log-rotation-maxbackups-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger with small rotation size and limited backups
	config := &RotationConfig{
		MaxSize:    50, // 50 bytes for easy testing
		MaxBackups: 3,  // Keep only 3 backups
	}

	logger, err := newLoggerWithRotation(logPath, config)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	// Write enough data to create multiple rotations
	// Need to create at least 5 rotations to trigger cleanup of old files
	for i := 0; i < 6; i++ {
		// Write enough to exceed MaxSize
		logger.Info("Test log message number %d to trigger rotation. Adding extra text to ensure we exceed the size limit.", i)
		// Delay to ensure different timestamps (second precision)
		time.Sleep(1100 * time.Millisecond)
	}

	// Wait for background cleanup to complete
	time.Sleep(500 * time.Millisecond)

	// Check that we have exactly MaxBackups + 1 files (current + backups)
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	logFiles := 0
	var fileNames []string
	for _, entry := range entries {
		fileNames = append(fileNames, entry.Name())
		if !entry.IsDir() && (strings.HasPrefix(entry.Name(), "test.log") || entry.Name() == "test.log") {
			logFiles++
		}
	}

	t.Logf("Files in directory: %v", fileNames)
	
	// Should have current log + MaxBackups
	expectedFiles := config.MaxBackups + 1
	if logFiles != expectedFiles {
		t.Errorf("Expected %d log files, got %d", expectedFiles, logFiles)
	}
}