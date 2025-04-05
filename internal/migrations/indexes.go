package migrations

import (
	"gorm.io/gorm"
)

// ApplyCompositeIndexes manually adds composite indexes for performance tuning.
func ApplyCompositeIndexes(db *gorm.DB) error {
	compositeIndexes := []string{}

	for _, query := range compositeIndexes {
		if err := db.Exec(query).Error; err != nil {
			return err
		}
	}
	return nil
}
