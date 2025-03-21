package repository

import (
	"log"

	rtdnModels "subsnotifpro-go/internal/google_playstore/rtdn/models"

	"gorm.io/gorm"
)

// SeedRepository interface for seeding data
type SeedRepository interface {
	SeedSubscriptionStates() error
}

// ✅ Struct with injected DB
type seedRepository struct {
	db *gorm.DB
}

// ✅ Constructor
func NewSeedRepository(db *gorm.DB) SeedRepository {
	return &seedRepository{db: db}
}

// ✅ Seed Subscription States
func (r *seedRepository) SeedSubscriptionStates() error {
	states := []rtdnModels.SubscriptionState{
		rtdnModels.SubscriptionStateUnspecified,
		rtdnModels.SubscriptionStatePending,
		rtdnModels.SubscriptionStateActive,
		rtdnModels.SubscriptionStatePaused,
		rtdnModels.SubscriptionStateInGracePeriod,
		rtdnModels.SubscriptionStateOnHold,
		rtdnModels.SubscriptionStateCanceled,
		rtdnModels.SubscriptionStateExpired,
		rtdnModels.SubscriptionStatePendingPurchaseCanceled,
	}

	for _, state := range states {
		var existing rtdnModels.SubscriptionStateModel
		err := r.db.Where("state = ?", state).First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			if err := r.db.Create(&rtdnModels.SubscriptionStateModel{State: state}).Error; err != nil {
				log.Printf("❌ Failed to insert state %s: %v", state, err)
				return err
			}
			log.Printf("✅ Inserted subscription state: %s", state)
		}
	}
	return nil
}
