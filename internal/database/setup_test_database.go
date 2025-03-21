package database

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"subsnotifpro-go/internal/models" // Import models for migration

	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDatabase initializes an in-memory SQLite test database with migrations.
func SetupTestDatabase() (*gorm.DB, error) {
	// Load environment variables from .env.test
	envPath, _ := filepath.Abs("../../.env.test")
	fmt.Println("🔹 Attempting to load .env.test from:", envPath)

	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("❌ Error loading .env.test file: %v", err)
	}

	fmt.Println("✅ .env.test file loaded successfully!")

	// Use SQLite in-memory database for tests
	testDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("❌ failed to connect to test database: %w", err)
	}

	// Run migrations
	fmt.Println("🚀 Running test database migrations...")
	if err := testDB.AutoMigrate(models.AllModels...); err != nil {
		return nil, fmt.Errorf("❌ failed to migrate test database: %w", err)
	}

	// Configure connection pooling for tests
	sqlDB, err := testDB.DB()
	if err != nil {
		return nil, fmt.Errorf("❌ failed to get test DB instance: %w", err)
	}
	sqlDB.SetMaxOpenConns(1) // Single connection for in-memory SQLite
	sqlDB.SetConnMaxLifetime(time.Minute * 5)

	fmt.Println("✅ Test database ready!")
	return testDB, nil
}
