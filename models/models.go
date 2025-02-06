package models

import (
	"subsnotifpro-go/internal/google_playstore/models" // Import Google Play models
)

// Collect all models in a list for AutoMigrate
var AllModels = []interface{}{
	&TestModel{},
	&models.GooglePlayServiceAccount{}, // Include models from internal/google_play/models
	&models.GooglePlaySettings{},
}
