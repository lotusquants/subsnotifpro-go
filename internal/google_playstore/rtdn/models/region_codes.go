package models

import (
	"github.com/google/uuid"
)

type RegionCode struct {
	ID   uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Code string    `gorm:"type:varchar(3);not null;unique"`
}
