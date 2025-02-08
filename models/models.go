package models

import (
	"subsnotifpro-go/internal/google_playstore/models" // Import Google Play models
)

// Collect all models in a list for AutoMigrate
var AllModels = []interface{}{
	&TestModel{},

	// Include models from internal/google_playstore/models
	&models.GooglePlayServiceAccount{},
	&models.GooglePlaySettings{},
	&models.GooglePlayWebhookEvent{},
}
