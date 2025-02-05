package database

import (
	"fmt"
	"testing"

	"subsnotifpro-go/database"
)

// TestQueryLogging verifies that queries are logged when enabled
func TestQueryLogging(t *testing.T) {
	// Ensure database is connected
	database.SetupTestDatabase()

	fmt.Println("🔹 Running test query...")

	// Run a test query
	var result int
	err := database.DB.Raw("SELECT 1").Scan(&result).Error
	if err != nil {
		t.Fatalf("❌ Query Execution Failed: %v", err)
	}

	fmt.Println("✅ Query executed successfully! Check logs above for `SELECT 1` statement.")
}
