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
	
	if config.MaxBackups != 7 {
		t.Errorf("Expected MaxBackups to be 7, got %d", config.MaxBackups)
	}
	
	if config.MaxAge != 30 {
		t.Errorf("Expected MaxAge to be 30, got %d", config.MaxAge)
	}
	
	if !config.Compress {
		t.Error("Expected Compress to be true")
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
		MaxSize:    100, // 100 bytes for easy testing
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   false, // Disable for testing
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