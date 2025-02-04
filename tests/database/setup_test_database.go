package database

import (
	"fmt"
	"log"

	"path/filepath"

	"subsnotifpro-go/database"

	"github.com/joho/godotenv"
)

// SetupTestDatabase loads environment variables and initializes the test database
func SetupTestDatabase() {
	// Get absolute path of .env.test
	envPath, _ := filepath.Abs("../../.env.test") // Adjust this if needed

	fmt.Println("🔹 Attempting to load .env.test from:", envPath)

	// Try loading .env.test explicitly
	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatalf("❌ Error loading .env.test file from %s: %v", envPath, err)
	}

	fmt.Println("✅ .env.test file loaded successfully!")

	// Connect to database
	fmt.Println("🔹 Connecting to test database...")
	database.ConnectDatabase()
}
