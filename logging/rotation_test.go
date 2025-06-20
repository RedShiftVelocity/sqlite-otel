package logging

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRotationConfig(t *testing.T) {
	config := DefaultRotationConfig()
	
	if config.MaxSize != 100*1024*1024 {
		t.Errorf("Expected MaxSize to be 100MB, got %d", config.MaxSize)
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
	// Create temp directory
	tmpDir, err := ioutil.TempDir("", "log-rotation-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	logPath := filepath.Join(tmpDir, "test.log")
	
	// Create logger with small rotation size
	config := &RotationConfig{
		MaxSize:    1024, // 1KB for testing
		MaxBackups: 3,
		MaxAge:     1,
		Compress:   false, // Disable compression for testing
	}
	
	logger, err := newLoggerWithRotation(logPath, config)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()
	
	// Write enough data to trigger rotation
	longMessage := strings.Repeat("x", 100)
	for i := 0; i < 20; i++ {
		logger.Info("Test message %d: %s", i, longMessage)
	}
	
	// Give rotation time to happen
	time.Sleep(100 * time.Millisecond)
	
	// Check for backup files
	files, err := filepath.Glob(filepath.Join(tmpDir, "test.log.*"))
	if err != nil {
		t.Fatal(err)
	}
	
	if len(files) == 0 {
		t.Error("Expected at least one backup file, found none")
	}
}

func TestCompressionAndCleanup(t *testing.T) {
	// Create temp directory
	tmpDir, err := ioutil.TempDir("", "log-compression-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Test file compression
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("This is a test file for compression")
	if err := ioutil.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatal(err)
	}
	
	if err := compressFile(testFile); err != nil {
		t.Fatal(err)
	}
	
	// Check compressed file exists
	compressedFile := testFile + ".gz"
	if _, err := os.Stat(compressedFile); err != nil {
		t.Error("Compressed file not found")
	}
	
	// Test cleanup with MaxBackups
	logPath := filepath.Join(tmpDir, "app.log")
	logger := &Logger{
		logPath: logPath,
		rotationConfig: &RotationConfig{
			MaxBackups: 2,
			MaxAge:     0, // Disable age-based cleanup
		},
	}
	
	// Create dummy backup files with timestamp format including microseconds
	for i := 0; i < 5; i++ {
		timestamp := time.Now().Add(time.Duration(-i) * time.Hour).Format("20060102-150405.000000")
		backupPath := fmt.Sprintf("%s.%s", logPath, timestamp)
		if err := ioutil.WriteFile(backupPath, []byte("backup"), 0644); err != nil {
			t.Fatal(err)
		}
		// Add delay to ensure different modification times
		time.Sleep(10 * time.Millisecond)
	}
	
	if err := logger.cleanupOldBackupsLocked(); err != nil {
		t.Fatal(err)
	}
	
	// Check remaining files
	files, err := filepath.Glob(filepath.Join(tmpDir, "app.log.*"))
	if err != nil {
		t.Fatal(err)
	}
	
	if len(files) > 2 {
		t.Errorf("Expected at most 2 backup files, found %d", len(files))
	}
}

func TestNeedsRotation(t *testing.T) {
	// Create temp file
	tmpFile, err := ioutil.TempFile("", "rotation-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	
	logger := &Logger{
		file: tmpFile,
		rotationConfig: &RotationConfig{
			MaxSize: 100, // 100 bytes
		},
	}
	
	// Initially should not need rotation
	if logger.needsRotationLocked() {
		t.Error("Empty file should not need rotation")
	}
	
	// Write data
	data := []byte(strings.Repeat("x", 101))
	if _, err := tmpFile.Write(data); err != nil {
		t.Fatal(err)
	}
	
	// Now should need rotation
	if !logger.needsRotationLocked() {
		t.Error("File exceeding MaxSize should need rotation")
	}
}