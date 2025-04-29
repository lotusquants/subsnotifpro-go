// models/consumption_request.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConsumptionRequest struct {
	gorm.Model
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SubscriptionID        uuid.UUID `gorm:"type:uuid;index"`
	OriginalTransactionID string    `gorm:"size:128;index"`

	// Consumption data
	AccountTenure            int32   `gorm:"default:0"`
	AppAccountToken          *string `gorm:"type:uuid"`
	ConsumptionStatus        int32   `gorm:"default:0"` // 0=undeclared
	CustomerConsented        bool    `gorm:"default:true"`
	DeliveryStatus           int32   `gorm:"default:0"` // 0=delivered
	LifetimeDollarsPurchased string  `gorm:"size:16;default:'0.00'"`
	LifetimeDollarsRefunded  string  `gorm:"size:16;default:'0.00'"`
	Platform                 int32   `gorm:"default:0"` // 0=undeclared
	PlayTime                 int32   `gorm:"default:0"` // minutes
	SampleContentProvided    bool    `gorm:"default:false"`
	UserStatus               int32   `gorm:"default:0"` // 0=undeclared

	// Apple response tracking
	RequestStatus    string `gorm:"size:32"` // pending, completed, failed
	AppleResponse    string `gorm:"type:jsonb"`
	ResponseReceived *time.Time

	// Relationships
	Subscription AppStoreSubscription `gorm:"foreignKey:SubscriptionID"`
}
