package database

import (
	"fmt"
	"testing"

	"subsnotifpro-go/database"
	"subsnotifpro-go/internal/models"
)

// TestAutoMigration checks if tables are created after auto-migration
func TestAutoMigration(t *testing.T) {
	// Ensure database is connected
	database.SetupTestDatabase()

	// Apply auto-migrations
	fmt.Println("🔹 Running auto-migrations...")
	database.AutoMigrateTables()

	// Verify that tables exist using GORM's HasTable method
	for _, model := range models.AllModels {
		if !database.DB.Migrator().HasTable(model) {
			t.Errorf("❌ Table for model %T was not created!", model)
		} else {
			fmt.Printf("✅ Table for model %T exists!\n", model)
		}
	}
}
