package tests

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/revenium/revenium-middleware-anthropic-go/revenium"
)

// TestLogger tests the logger functionality
func TestLogger(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	// Create a new logger
	logger := revenium.NewDefaultLogger()

	// Test different log levels
	logger.SetLevel(revenium.LogLevelDebug)

	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warning message")
	logger.Error("Error message")

	output := buf.String()

	// Check that all messages are present
	if !strings.Contains(output, "Debug message") {
		t.Error("Debug message not found in output")
	}
	if !strings.Contains(output, "Info message") {
		t.Error("Info message not found in output")
	}
	if !strings.Contains(output, "Warning message") {
		t.Error("Warning message not found in output")
	}
	if !strings.Contains(output, "Error message") {
		t.Error("Error message not found in output")
	}

	// Check log level prefixes
	if !strings.Contains(output, "[Revenium DEBUG]") {
		t.Error("Debug prefix not found")
	}
	if !strings.Contains(output, "[Revenium INFO]") {
		t.Error("Info prefix not found")
	}
	if !strings.Contains(output, "[Revenium WARN]") {
		t.Error("Warn prefix not found")
	}
	if !strings.Contains(output, "[Revenium ERROR]") {
		t.Error("Error prefix not found")
	}
}

// TestLoggerLevels tests that log levels work correctly
func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := revenium.NewDefaultLogger()

	// Set to INFO level - should not show DEBUG
	logger.SetLevel(revenium.LogLevelInfo)
	buf.Reset()

	logger.Debug("Debug message")
	logger.Info("Info message")

	output := buf.String()

	// Debug should not appear
	if strings.Contains(output, "Debug message") {
		t.Error("Debug message should not appear at INFO level")
	}

	// Info should appear
	if !strings.Contains(output, "Info message") {
		t.Error("Info message should appear at INFO level")
	}
}

// TestLoggerFormatting tests message formatting
func TestLoggerFormatting(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := revenium.NewDefaultLogger()
	logger.SetLevel(revenium.LogLevelDebug)

	// Test formatted messages
	logger.Info("User %s has %d tokens", "test-user", 100)

	output := buf.String()

	if !strings.Contains(output, "User test-user has 100 tokens") {
		t.Error("Formatted message not correct")
	}
}

// TestGlobalLogger tests the global logger functions
func TestGlobalLogger(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	// Set environment variable for testing
	os.Setenv("REVENIUM_LOG_LEVEL", "DEBUG")
	defer os.Unsetenv("REVENIUM_LOG_LEVEL")

	// Initialize logger
	revenium.InitializeLogger()

	// Test global functions
	revenium.Debug("Global debug")
	revenium.Info("Global info")
	revenium.Warn("Global warn")
	revenium.Error("Global error")

	output := buf.String()

	// Check that all messages are present
	if !strings.Contains(output, "Global debug") {
		t.Error("Global debug message not found")
	}
	if !strings.Contains(output, "Global info") {
		t.Error("Global info message not found")
	}
	if !strings.Contains(output, "Global warn") {
		t.Error("Global warn message not found")
	}
	if !strings.Contains(output, "Global error") {
		t.Error("Global error message not found")
	}
}

// TestLogLevelParsing tests log level parsing from strings
func TestLogLevelParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected revenium.LogLevel
	}{
		{"DEBUG", revenium.LogLevelDebug},
		{"debug", revenium.LogLevelDebug},
		{"INFO", revenium.LogLevelInfo},
		{"info", revenium.LogLevelInfo},
		{"WARN", revenium.LogLevelWarn},
		{"warn", revenium.LogLevelWarn},
		{"WARNING", revenium.LogLevelWarn},
		{"ERROR", revenium.LogLevelError},
		{"error", revenium.LogLevelError},
		{"INVALID", revenium.LogLevelInfo}, // Should default to INFO
		{"", revenium.LogLevelInfo},        // Should default to INFO
	}

	for _, test := range tests {
		result := revenium.ParseLogLevel(test.input)
		if result != test.expected {
			t.Errorf("ParseLogLevel(%s) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

// TestLoggerInitialization tests logger initialization from environment
func TestLoggerInitialization(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	// Test with verbose startup
	os.Setenv("REVENIUM_LOG_LEVEL", "INFO")
	os.Setenv("REVENIUM_VERBOSE_STARTUP", "true")
	defer func() {
		os.Unsetenv("REVENIUM_LOG_LEVEL")
		os.Unsetenv("REVENIUM_VERBOSE_STARTUP")
	}()

	revenium.InitializeLogger()

	output := buf.String()

	// Should contain initialization message
	if !strings.Contains(output, "Logger initialized with level: INFO") {
		t.Error("Logger initialization message not found")
	}
}
