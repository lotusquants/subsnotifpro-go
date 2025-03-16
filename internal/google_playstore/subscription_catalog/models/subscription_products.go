package models

// -------------------------
// ✅ ENUMS for Better Type Safety
// -------------------------

type EeaWithdrawalRightType string

const (
	WithdrawalRightUnspecified    EeaWithdrawalRightType = "WITHDRAWAL_RIGHT_TYPE_UNSPECIFIED"
	WithdrawalRightDigitalContent EeaWithdrawalRightType = "WITHDRAWAL_RIGHT_DIGITAL_CONTENT"
	WithdrawalRightService        EeaWithdrawalRightType = "WITHDRAWAL_RIGHT_SERVICE"
)

type TaxTier string

const (
	TaxTierUnspecified      TaxTier = "TAX_TIER_UNSPECIFIED"
	TaxTierBooks1           TaxTier = "TAX_TIER_BOOKS_1"
	TaxTierNews1            TaxTier = "TAX_TIER_NEWS_1"
	TaxTierNews2            TaxTier = "TAX_TIER_NEWS_2"
	TaxTierMusicOrAudio1    TaxTier = "TAX_TIER_MUSIC_OR_AUDIO_1"
	TaxTierLiveOrBroadcast1 TaxTier = "TAX_TIER_LIVE_OR_BROADCAST_1"
)

type StreamingTaxType string

const (
	StreamingTaxTypeUnspecified            StreamingTaxType = "STREAMING_TAX_TYPE_UNSPECIFIED"
	StreamingTaxTypeTelcoVideoRental       StreamingTaxType = "STREAMING_TAX_TYPE_TELCO_VIDEO_RENTAL"
	StreamingTaxTypeTelcoVideoSales        StreamingTaxType = "STREAMING_TAX_TYPE_TELCO_VIDEO_SALES"
	StreamingTaxTypeTelcoVideoMultiChannel StreamingTaxType = "STREAMING_TAX_TYPE_TELCO_VIDEO_MULTI_CHANNEL"
	StreamingTaxTypeTelcoAudioRental       StreamingTaxType = "STREAMING_TAX_TYPE_TELCO_AUDIO_RENTAL"
	StreamingTaxTypeTelcoAudioSales        StreamingTaxType = "STREAMING_TAX_TYPE_TELCO_AUDIO_SALES"
	StreamingTaxTypeTelcoAudioMultiChannel StreamingTaxType = "STREAMING_TAX_TYPE_TELCO_AUDIO_MULTI_CHANNEL"
)

// -------------------------
// ✅ Product Subscription Model
// -------------------------

type ProductSubscription struct {
	PackageName string `gorm:"primaryKey;size:40;not null" json:"packageName"`
	ProductID   string `gorm:"primaryKey;size:40;not null" json:"productId"`

	// ✅ Relations
	BasePlans                  []ProductBasePlan                    `gorm:"foreignKey:PackageName,ProductID;references:PackageName,ProductID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE;" json:"basePlans"`
	Listings                   []ProductSubscriptionListing         `gorm:"foreignKey:PackageName,ProductID;references:PackageName,ProductID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE;" json:"listings"`
	TaxAndComplianceSettings   SubscriptionTaxAndComplianceSettings `gorm:"foreignKey:PackageName,ProductID;references:PackageName,ProductID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE;" json:"taxAndComplianceSettings"`
	RestrictedPaymentCountries RestrictedPaymentCountries           `gorm:"foreignKey:PackageName,ProductID;references:PackageName,ProductID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE;" json:"restrictedPaymentCountries"`
}

// -------------------------
// ✅ Subscription Listing (Localized Metadata)
// -------------------------

type ProductSubscriptionListing struct {
	PackageName  string   `gorm:"primaryKey;size:40;not null" json:"packageName"`
	ProductID    string   `gorm:"primaryKey;size:40;not null" json:"productId"`
	LanguageCode string   `gorm:"primaryKey;size:10;not null" json:"languageCode"`
	Title        string   `gorm:"size:255;not null" json:"title"`
	Description  string   `gorm:"size:80;not null" json:"description"`
	Benefits     []string `gorm:"serializer:json" json:"benefits"` // ✅ JSON serialization
}

// -------------------------
// ✅ Restricted Payment Countries
// -------------------------

type RestrictedPaymentCountries struct {
	PackageName string   `gorm:"primaryKey;size:40;not null" json:"packageName"`
	ProductID   string   `gorm:"primaryKey;size:40;not null" json:"productId"`
	RegionCodes []string `gorm:"serializer:json" json:"regionCodes"` // ✅ JSON serialization
}

// -------------------------
// ✅ Taxation & Compliance Settings
// -------------------------

type SubscriptionTaxAndComplianceSettings struct {
	PackageName             string                 `gorm:"primaryKey;size:40;not null" json:"packageName"`
	ProductID               string                 `gorm:"primaryKey;size:40;not null" json:"productId"`
	EeaWithdrawalRightType  EeaWithdrawalRightType `gorm:"type:varchar(50);not null" json:"eeaWithdrawalRightType"`
	IsTokenizedDigitalAsset bool                   `gorm:"not null" json:"isTokenizedDigitalAsset"`

	// ✅ Use table instead of JSON map
	TaxRateInfoByRegion []RegionalTaxRateInfo `gorm:"foreignKey:PackageName,ProductID;references:PackageName,ProductID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"taxRateInfoByRegion"`
}

// -------------------------
// ✅ Regional Tax Rate Info (Separate Table instead of Map)
// -------------------------

type RegionalTaxRateInfo struct {
	ID                             uint             `gorm:"primaryKey;autoIncrement" json:"id"`
	PackageName                    string           `gorm:"not null;size:40;index" json:"packageName"`
	ProductID                      string           `gorm:"not null;size:40;index" json:"productId"`
	RegionCode                     string           `gorm:"not null;size:3;index" json:"regionCode"` // ✅ Enforce ISO country code format
	TaxTier                        TaxTier          `gorm:"type:varchar(50);not null" json:"taxTier"`
	EligibleForStreamingServiceTax bool             `gorm:"not null" json:"eligibleForStreamingServiceTax"`
	StreamingTaxType               StreamingTaxType `gorm:"type:varchar(50);not null" json:"streamingTaxType"`
}
