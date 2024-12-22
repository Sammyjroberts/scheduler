package scheduler

import (
	"os"
	"testing"
)

// TestMain is the entry point for all tests
func TestMain(m *testing.M) {
	// Setup code if needed (e.g., initialize test data, create temp files)

	// Run all tests
	exitCode := m.Run()

	// Cleanup code if needed

	// Exit with the test result code
	os.Exit(exitCode)
}
