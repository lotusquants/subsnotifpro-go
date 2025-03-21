package database

import (
	"log"

	"subsnotifpro-go/internal/database/repository"

	"gorm.io/gorm"
)

// ✅ Run All Seeders
func SeedDatabase(db *gorm.DB) error {
	seedRepo := repository.NewSeedRepository(db)

	log.Println("🌱 Seeding database...")

	// ✅ Run Subscription States Seeder
	if err := seedRepo.SeedSubscriptionStates(); err != nil {
		return err
	}

	log.Println("✅ Seeding completed successfully!")
	return nil
}
