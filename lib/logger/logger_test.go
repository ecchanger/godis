package logger

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewStdoutLogger(t *testing.T) {
	logger := NewStdoutLogger()
	if logger == nil {
		t.Fatal("NewStdoutLogger should not return nil")
	}
	if logger.logFile != nil {
		t.Error("stdout logger should not have logFile")
	}
	if logger.logger == nil {
		t.Error("stdout logger should have logger")
	}
	if logger.entryChan == nil {
		t.Error("stdout logger should have entryChan")
	}
	if logger.entryPool == nil {
		t.Error("stdout logger should have entryPool")
	}
}

func TestNewFileLogger(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := ioutil.TempDir("", "logger_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	settings := &Settings{
		Path:       tmpDir,
		Name:       "test",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	}

	logger, err := NewFileLogger(settings)
	if err != nil {
		t.Fatal(err)
	}
	if logger == nil {
		t.Fatal("NewFileLogger should not return nil")
	}
	if logger.logFile == nil {
		t.Error("file logger should have logFile")
	}
	if logger.logger == nil {
		t.Error("file logger should have logger")
	}
}

func TestLoggerOutput(t *testing.T) {
	logger := NewStdoutLogger()
	
	// Test different log levels
	logger.Output(DEBUG, 1, "debug message")
	logger.Output(INFO, 1, "info message")
	logger.Output(WARNING, 1, "warning message")
	logger.Output(ERROR, 1, "error message")
	logger.Output(FATAL, 1, "fatal message")
	
	// Give some time for async processing
	time.Sleep(100 * time.Millisecond)
}

func TestDebugFunctions(t *testing.T) {
	// Test Debug function
	Debug("test debug message")
	Debug("test", "debug", "with", "multiple", "params")
	
	// Test Debugf function
	Debugf("test debug formatted %s %d", "message", 123)
	
	// Give some time for async processing
	time.Sleep(100 * time.Millisecond)
}

func TestInfoFunctions(t *testing.T) {
	// Test Info function
	Info("test info message")
	Info("test", "info", "with", "multiple", "params")
	
	// Test Infof function
	Infof("test info formatted %s %d", "message", 456)
	
	// Give some time for async processing
	time.Sleep(100 * time.Millisecond)
}

func TestWarnFunction(t *testing.T) {
	// Test Warn function
	Warn("test warn message")
	Warn("test", "warn", "with", "multiple", "params")
	
	// Give some time for async processing
	time.Sleep(100 * time.Millisecond)
}

func TestErrorFunctions(t *testing.T) {
	// Test Error function
	Error("test error message")
	Error("test", "error", "with", "multiple", "params")
	
	// Test Errorf function
	Errorf("test error formatted %s %d", "message", 789)
	
	// Give some time for async processing
	time.Sleep(100 * time.Millisecond)
}

func TestFatalFunction(t *testing.T) {
	// Test Fatal function (but don't call it as it would terminate the program)
	// We can only test the formatting logic by using a custom logger
	logger := NewStdoutLogger()
	logger.Output(FATAL, 1, "test fatal message")
	
	// Give some time for async processing
	time.Sleep(100 * time.Millisecond)
}

func TestSetup(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := ioutil.TempDir("", "logger_setup_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	settings := &Settings{
		Path:       tmpDir,
		Name:       "setup_test",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	}

	// Save original DefaultLogger
	originalLogger := DefaultLogger
	defer func() {
		DefaultLogger = originalLogger
	}()

	Setup(settings)
	
	// Verify DefaultLogger was set
	if DefaultLogger == nil {
		t.Error("Setup should set DefaultLogger")
	}
}

func TestFileLoggerWithRotation(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := ioutil.TempDir("", "logger_rotation_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	settings := &Settings{
		Path:       tmpDir,
		Name:       "rotation_test",
		Ext:        "log",
		TimeFormat: "2006-01-02-15-04-05", // Include seconds for quick rotation test
	}

	logger, err := NewFileLogger(settings)
	if err != nil {
		t.Fatal(err)
	}

	// Log a message
	logger.Output(INFO, 1, "first message")
	
	// Wait a second to ensure time difference
	time.Sleep(1 * time.Second)
	
	// Log another message (should potentially create new file due to time format)
	logger.Output(INFO, 1, "second message")
	
	// Give some time for async processing
	time.Sleep(100 * time.Millisecond)
	
	// Check if log files were created
	files, err := filepath.Glob(filepath.Join(tmpDir, "rotation_test-*.log"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Error("No log files were created")
	}
}

func TestLogLevels(t *testing.T) {
	levels := []LogLevel{DEBUG, INFO, WARNING, ERROR, FATAL}
	expectedFlags := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
	
	for i, level := range levels {
		if int(level) != i {
			t.Errorf("LogLevel %v should equal %d", level, i)
		}
		if levelFlags[level] != expectedFlags[i] {
			t.Errorf("levelFlags[%v] should be %s, got %s", level, expectedFlags[i], levelFlags[level])
		}
	}
}

func TestLogEntryPool(t *testing.T) {
	logger := NewStdoutLogger()
	
	// Get entry from pool
	entry := logger.entryPool.Get().(*logEntry)
	if entry == nil {
		t.Error("should get logEntry from pool")
	}
	
	// Use entry
	entry.msg = "test message"
	entry.level = INFO
	
	// Put back to pool
	logger.entryPool.Put(entry)
	
	// Get again to verify pooling works
	entry2 := logger.entryPool.Get().(*logEntry)
	if entry2 == nil {
		t.Error("should get logEntry from pool again")
	}
}

func TestInvalidFileLogger(t *testing.T) {
	// Test with invalid path
	settings := &Settings{
		Path:       "/invalid/path/that/does/not/exist/and/cannot/be/created",
		Name:       "test",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	}

	_, err := NewFileLogger(settings)
	if err == nil {
		t.Error("NewFileLogger should return error for invalid path")
	}
	if !strings.Contains(err.Error(), "logging.Join err") {
		t.Errorf("Error message should contain 'logging.Join err', got: %s", err.Error())
	}
}