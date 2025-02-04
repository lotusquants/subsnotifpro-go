package models

// TestModel1 is a sample test model
type TestModel struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"not null"`
}
