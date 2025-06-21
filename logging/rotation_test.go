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

func TestMaxAgeRetention(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "log-rotation-maxage-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger with MaxAge policy (keep files for 7 days)
	config := &RotationConfig{
		MaxSize:    100, // Small size for easy testing
		MaxBackups: 10,  // Large number so MaxAge is the limiting factor
		MaxAge:     7,   // Keep files for 7 days
	}

	logger, err := newLoggerWithRotation(logPath, config)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	// Create mock backup files with different ages
	now := time.Now()
	
	// Files that should be kept (within 7 days) - generate names from actual dates
	recentFiles := []struct {
		name string
		age  time.Duration
	}{
		{now.Add(-24 * time.Hour).Format("test.log.20060102-150405"), 24 * time.Hour},      // 1 day old
		{now.Add(-48 * time.Hour).Format("test.log.20060102-150405"), 48 * time.Hour},     // 2 days old  
		{now.Add(-96 * time.Hour).Format("test.log.20060102-150405"), 96 * time.Hour},     // 4 days old
		{now.Add(-144 * time.Hour).Format("test.log.20060102-150405"), 144 * time.Hour},   // 6 days old
	}

	// Files that should be deleted (older than 7 days) - generate names from actual dates
	oldFiles := []struct {
		name string
		age  time.Duration
	}{
		{now.Add(-192 * time.Hour).Format("test.log.20060102-150405"), 192 * time.Hour},   // 8 days old
		{now.Add(-240 * time.Hour).Format("test.log.20060102-150405"), 240 * time.Hour},  // 10 days old
		{now.Add(-288 * time.Hour).Format("test.log.20060102-150405"), 288 * time.Hour},  // 12 days old
	}

	// Create recent backup files (should be kept)
	for _, file := range recentFiles {
		filePath := filepath.Join(tmpDir, file.name)
		content := fmt.Sprintf("Log content for %s", file.name)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Create old backup files (should be deleted)
	for _, file := range oldFiles {
		filePath := filepath.Join(tmpDir, file.name)
		content := fmt.Sprintf("Log content for %s", file.name)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Verify all files exist before cleanup
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	
	initialFileCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "test.log") {
			initialFileCount++
		}
	}
	
	expectedInitialFiles := len(recentFiles) + len(oldFiles) + 1 // +1 for current log
	if initialFileCount != expectedInitialFiles {
		t.Errorf("Expected %d initial files, got %d", expectedInitialFiles, initialFileCount)
	}

	// Trigger rotation to activate cleanup
	logger.Info("Test message to trigger rotation and cleanup")
	
	// Wait for background cleanup to complete
	time.Sleep(200 * time.Millisecond)

	// Check which files remain after cleanup
	entries, err = os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	var remainingFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "test.log") {
			remainingFiles = append(remainingFiles, entry.Name())
		}
	}

	t.Logf("Remaining files after MaxAge cleanup: %v", remainingFiles)

	// Verify old files were deleted
	for _, oldFile := range oldFiles {
		found := false
		for _, remaining := range remainingFiles {
			if remaining == oldFile.name {
				found = true
				break
			}
		}
		if found {
			t.Errorf("Old file %s should have been deleted but still exists", oldFile.name)
		}
	}

	// Verify recent files were kept
	for _, recentFile := range recentFiles {
		found := false
		for _, remaining := range remainingFiles {
			if remaining == recentFile.name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Recent file %s should have been kept but was deleted", recentFile.name)
		}
	}

	// Should have: current log + recent backup files + 1 new backup from rotation
	expectedFinalFiles := 1 + len(recentFiles) + 1 // current + recent + new backup
	if len(remainingFiles) != expectedFinalFiles {
		t.Errorf("Expected %d files after cleanup, got %d", expectedFinalFiles, len(remainingFiles))
	}
}