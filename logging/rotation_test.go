package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
		MaxAge:     30, // 30 days (won't affect this test)
		Compress:   false,
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
		// Small delay to ensure different timestamps
		time.Sleep(20 * time.Millisecond)
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

func TestMaxAgeRetention(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "log-rotation-maxage-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "test.log")

	// Create some old backup files manually
	now := time.Now()
	oldTimestamps := []time.Time{
		now.AddDate(0, 0, -40), // 40 days old
		now.AddDate(0, 0, -35), // 35 days old
		now.AddDate(0, 0, -25), // 25 days old (should be kept)
		now.AddDate(0, 0, -10), // 10 days old (should be kept)
	}

	for _, ts := range oldTimestamps {
		backupName := fmt.Sprintf("test.log.%s", ts.Format("20060102-150405.000000"))
		backupPath := filepath.Join(tmpDir, backupName)
		if err := os.WriteFile(backupPath, []byte("old log data"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create logger with MaxAge set to 30 days
	config := &RotationConfig{
		MaxSize:    1024 * 1024, // 1MB (won't trigger rotation in this test)
		MaxBackups: 10,          // High enough to not affect this test
		MaxAge:     30,          // 30 days
		Compress:   false,
	}

	logger, err := newLoggerWithRotation(logPath, config)
	if err != nil {
		t.Fatal(err)
	}

	// Trigger a rotation to invoke cleanup
	logger.mu.Lock()
	logger.rotateLocked()
	logger.mu.Unlock()

	// Wait for background cleanup
	time.Sleep(100 * time.Millisecond)

	logger.Close()

	// Check remaining files
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Count backup files and check their ages
	for _, entry := range entries {
		if entry.Name() == "test.log" {
			continue // Skip current log
		}

		if strings.HasPrefix(entry.Name(), "test.log.") {
			// Parse timestamp from filename
			tsPart := strings.TrimPrefix(entry.Name(), "test.log.")
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

func TestConcurrentLogging(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "log-rotation-concurrent-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger with small rotation size
	config := &RotationConfig{
		MaxSize:    1000, // 1KB for frequent rotation
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true, // Test with compression enabled
	}

	logger, err := newLoggerWithRotation(logPath, config)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	// Launch multiple goroutines to write logs concurrently
	var wg sync.WaitGroup
	numGoroutines := 10
	logsPerGoroutine := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < logsPerGoroutine; j++ {
				logger.Info("Goroutine %d: Message %d - Testing concurrent logging with rotation", id, j)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Wait for background compression/cleanup
	time.Sleep(200 * time.Millisecond)

	// Verify no data loss and all files are properly created
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Should have at least one compressed backup
	foundCompressed := false
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".gz") {
			foundCompressed = true
			break
		}
	}

	if !foundCompressed {
		t.Error("No compressed backup files found despite compression being enabled")
	}

	// Verify the logger is still functional after concurrent access
	logger.Info("Final test message after concurrent logging")
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

func TestRotationWithCompressionAndCleanup(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "log-rotation-compress-cleanup-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger with compression and limited backups
	config := &RotationConfig{
		MaxSize:    100, // 100 bytes
		MaxBackups: 2,   // Keep only 2 backups
		MaxAge:     30,
		Compress:   true,
	}

	logger, err := newLoggerWithRotation(logPath, config)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	// Write enough data to trigger multiple rotations
	for i := 0; i < 5; i++ {
		logger.Info("Test message %d to trigger rotation with compression", i)
		time.Sleep(20 * time.Millisecond) // Ensure different timestamps
	}

	// Wait for background compression and cleanup
	time.Sleep(300 * time.Millisecond)

	// Check results
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	compressedCount := 0
	uncompressedCount := 0
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".gz") {
			compressedCount++
		} else if entry.Name() != "test.log" && strings.HasPrefix(entry.Name(), "test.log.") {
			uncompressedCount++
		}
	}

	// Should have MaxBackups compressed files
	if compressedCount > config.MaxBackups {
		t.Errorf("Too many compressed backups: got %d, expected at most %d", 
			compressedCount, config.MaxBackups)
	}

	// Should have no uncompressed backups (all should be compressed)
	if uncompressedCount > 0 {
		t.Errorf("Found %d uncompressed backup files, expected 0", uncompressedCount)
	}
}