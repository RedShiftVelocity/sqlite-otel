package logging

import (
	"fmt"
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
	
	if !config.Compress {
		t.Error("Expected Compress to be true by default")
	}
	
	if config.MaxBackups != 7 {
		t.Errorf("Expected MaxBackups to be 7, got %d", config.MaxBackups)
	}
	
	if config.MaxAge != 30 {
		t.Errorf("Expected MaxAge to be 30, got %d", config.MaxAge)
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

func TestMaxBackupsRetention(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "log-retention-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "test.log")

	// Create some old backup files manually
	now := time.Now()
	for i := 0; i < 10; i++ {
		ts := now.Add(-time.Duration(i) * time.Minute)
		backupName := fmt.Sprintf("test.log.%s.gz", ts.Format("20060102-150405.000000"))
		backupPath := filepath.Join(tmpDir, backupName)
		if err := os.WriteFile(backupPath, []byte("old log data"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create logger with MaxBackups=3
	config := &RotationConfig{
		MaxSize:    1024 * 1024, // 1MB (won't trigger rotation in this test)
		MaxBackups: 3,
		MaxAge:     30,
		Compress:   true,
	}

	logger, err := newLoggerWithRotation(logPath, config)
	if err != nil {
		t.Fatal(err)
	}

	// Trigger cleanup by writing a message (won't rotate due to size)
	logger.Info("Test message")
	
	// Manually trigger cleanup
	logger.cleanupOldBackups()
	
	logger.Close()

	// Check remaining files
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Count backup files
	var backupCount int
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".gz") {
			backupCount++
		}
	}

	// Should have exactly MaxBackups files
	if backupCount != config.MaxBackups {
		t.Errorf("Expected %d backup files, got %d", config.MaxBackups, backupCount)
	}
}

func TestMaxAgeRetention(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "log-age-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "test.log")

	// Create some old backup files
	now := time.Now()
	oldTimestamps := []time.Time{
		now.AddDate(0, 0, -40), // 40 days old
		now.AddDate(0, 0, -35), // 35 days old  
		now.AddDate(0, 0, -25), // 25 days old (should be kept)
		now.AddDate(0, 0, -10), // 10 days old (should be kept)
	}

	for _, ts := range oldTimestamps {
		backupName := fmt.Sprintf("test.log.%s.gz", ts.Format("20060102-150405.000000"))
		backupPath := filepath.Join(tmpDir, backupName)
		if err := os.WriteFile(backupPath, []byte("old log data"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create logger with MaxAge=30 days
	config := &RotationConfig{
		MaxSize:    1024 * 1024,
		MaxBackups: 10, // High enough to not affect this test
		MaxAge:     30,
		Compress:   true,
	}

	logger, err := newLoggerWithRotation(logPath, config)
	if err != nil {
		t.Fatal(err)
	}

	// Manually trigger cleanup
	logger.cleanupOldBackups()
	
	logger.Close()

	// Check remaining files
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Check that old files were removed
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".gz") {
			// Parse timestamp from filename
			tsPart := strings.TrimPrefix(entry.Name(), "test.log.")
			tsPart = strings.TrimSuffix(tsPart, ".gz")
			ts, err := time.Parse("20060102-150405.000000", tsPart)
			if err != nil {
				continue
			}

			age := now.Sub(ts).Hours() / 24 // Age in days
			if age > 30 {
				t.Errorf("Found backup file older than %d days: %s (%.0f days old)", 
					config.MaxAge, entry.Name(), age)
			}
		}
	}
}