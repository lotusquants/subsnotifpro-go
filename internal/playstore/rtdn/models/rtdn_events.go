package models

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// GooglePlayWebhookEvent represents the top-level RTDN webhook payload
type GooglePlayWebhookEvent struct {
	ID                         string                      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Version                    string                      `gorm:"type:varchar(10);not null"`
	PackageName                string                      `gorm:"type:varchar(255);not null"`
	EventTimeMillis            int64                       `gorm:"not null"`
	SubscriptionNotification   *SubscriptionNotification   `gorm:"embedded;embeddedPrefix:sub_"`
	OneTimeProductNotification *OneTimeProductNotification `gorm:"embedded;embeddedPrefix:one_"`
	VoidedPurchaseNotification *VoidedPurchaseNotification `gorm:"embedded;embeddedPrefix:void_"`
	TestNotification           *TestNotification           `gorm:"embedded;embeddedPrefix:test_"`
	RawPayload                 string                      `gorm:"type:jsonb;not null"`
	Status                     string                      `gorm:"type:varchar(50);default:'pending';index"`
	RetryCount                 int                         `gorm:"default:0"`
	CreatedAt                  time.Time                   `gorm:"autoCreateTime;index"`
	UpdatedAt                  time.Time                   `gorm:"autoUpdateTime"`
	DeletedAt                  gorm.DeletedAt              `gorm:"index"`
}

// TestNotification represents a test notification from Google Play Console
type TestNotification struct {
	Version string `json:"version,omitempty" gorm:"default:null"`
}

// ConvertFields handles type conversions after unmarshalling
func (e *GooglePlayWebhookEvent) ConvertFields() error {
	if e.EventTimeMillis == 0 {
		return nil // Already converted, no need to process
	}

	// If event time is received as a string, convert it
	var eventTimeStr string
	if err := json.Unmarshal([]byte(e.RawPayload), &eventTimeStr); err == nil {
		eventTimeInt, err := strconv.ParseInt(eventTimeStr, 10, 64)
		if err != nil {
			return errors.New("❌ Failed to convert eventTimeMillis to int64")
		}
		e.EventTimeMillis = eventTimeInt
	}
	return nil
}
