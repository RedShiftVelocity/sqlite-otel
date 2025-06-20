package logging

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRotationConfig(t *testing.T) {
	config := DefaultRotationConfig()
	
	if config.MaxSize != 100*1024*1024 {
		t.Errorf("Expected MaxSize to be 104857600, got %d", config.MaxSize)
	}
}

func TestBasicLogRotation(t *testing.T) {
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
	for i := 0; i < 10; i++ {
		logger.Info("Test log message number %d", i)
	}

	// Check that rotation occurred
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Count log files
	var currentLog, backupFiles int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if entry.Name() == "test.log" {
			currentLog++
		} else if len(entry.Name()) > 8 && entry.Name()[:8] == "test.log" {
			backupFiles++
		}
	}

	// Should have exactly 1 current log file
	if currentLog != 1 {
		t.Errorf("Expected exactly 1 current log file, got %d", currentLog)
	}

	// Should have at least 1 backup file after rotation
	if backupFiles < 1 {
		t.Errorf("Expected at least 1 backup file after rotation, got %d", backupFiles)
	}
}