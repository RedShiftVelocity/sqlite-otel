package logging

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRotationConfig(t *testing.T) {
	config := DefaultRotationConfig()
	
	if config.MaxSize != 100*1024*1024 {
		t.Errorf("Expected MaxSize to be 104857600, got %d", config.MaxSize)
	}
	
	if !config.Compress {
		t.Error("Expected Compress to be true by default")
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

func TestCompressFile(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "compress-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer os.Remove(tmpFile.Name() + ".gz")

	// Write test data
	testData := "Test data for compression"
	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Compress file
	if err := compressFile(tmpFile.Name()); err != nil {
		t.Fatal(err)
	}

	// Check compressed file exists
	if _, err := os.Stat(tmpFile.Name() + ".gz"); err != nil {
		t.Error("Compressed file not found")
	}
}

func TestRotationWithCompression(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "log-rotation-compress-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger with compression enabled
	config := &RotationConfig{
		MaxSize:  50, // Small size to trigger rotation
		Compress: true,
	}

	logger, err := newLoggerWithRotation(logPath, config)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	// Write enough data to trigger rotation
	for i := 0; i < 5; i++ {
		logger.Info("Test message %d to trigger rotation with compression", i)
	}

	// Give time for async compression
	time.Sleep(100 * time.Millisecond)

	// Check for compressed backup
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	foundCompressed := false
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".gz" {
			foundCompressed = true
			break
		}
	}

	if !foundCompressed {
		t.Error("No compressed backup file found")
	}
}