package tests

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
)

// TestMain runs before all tests and loads the .env file
func TestMain(m *testing.M) {
	// Load .env file from the parent directory
	envPath := filepath.Join("..", ".env")

	if err := godotenv.Load(envPath); err != nil {
		log.Printf("Warning: Could not load .env file from %s: %v", envPath, err)
		log.Printf("Tests will use environment variables from the system")
	} else {
		log.Printf("Successfully loaded .env file from %s", envPath)
	}

	// Run all tests
	code := m.Run()

	// Exit with the test result code
	os.Exit(code)
}
