package repository

import (
	"context"
	"errors"

	"subsnotifpro-go/internal/google_playstore/subscription_catalog/models"

	"gorm.io/gorm"
)

// Custom errors for better error handling
var (
	ErrNotFound      = errors.New("record not found")
	ErrInactiveOffer = errors.New("offer is inactive")
)

// SubscriptionCatalogRepository defines methods for managing subscription catalog data.
type SubscriptionCatalogRepository interface {
	UpsertSubscriptionProducts(ctx context.Context, subscription *models.ProductSubscription) error
	DeleteAllSubscriptionProducts(ctx context.Context) error
	UpsertSubscriptionBasePlans(ctx context.Context, basePlan *models.ProductBasePlan) error
	UpsertSubscriptionOffers(ctx context.Context, offers []models.SubscriptionOffer) error
	DeleteAllSubscriptionOffers(ctx context.Context, packageName, productID, basePlanID string) error

	ListSubscriptionProducts(ctx context.Context) ([]models.ProductSubscription, error)
	GetSubscriptionProduct(ctx context.Context, packageName, productID string) (*models.ProductSubscription, error)
	CheckSubscriptionProductExists(ctx context.Context, packageName, productID string) (bool, error)

	GetBasePlanDetails(ctx context.Context, packageName, productID, basePlanID string) (*models.ProductBasePlan, error)
	IsBasePlanActive(ctx context.Context, packageName, productID, basePlanID string) (bool, error)
	IsBasePlanAvailableInRegion(ctx context.Context, packageName, productID, basePlanID, regionCode string) (bool, error)
	ListBasePlanNames(ctx context.Context, packageName, productID string) ([]string, error)
	GetRegionalBasePlanPrice(ctx context.Context, packageName, productID, basePlanID, regionCode string) (*models.Money, error)
	GetOtherRegionsBasePlanPrice(ctx context.Context, packageName, productID, basePlanID, currency string) (*models.Money, error)

	GetSubscriptionOffer(ctx context.Context, packageName, productID, basePlanID, offerID string) (*models.SubscriptionOffer, error)
	ListOfferNamesForBasePlan(ctx context.Context, packageName, productID, basePlanID string) ([]string, error)
	IsSubscriptionOfferActive(ctx context.Context, packageName, productID, basePlanID, offerID string) (bool, error)
	IsSubscriptionOfferAvailableInRegion(ctx context.Context, packageName, productID, basePlanID, offerID, regionCode string) (bool, error)

	GetSubscriptionOfferPhases(ctx context.Context, packageName, productID, basePlanID, offerID string) ([]models.SubscriptionOfferPhase, error)
	IsSubscriptionOfferPhaseExists(ctx context.Context, packageName, productID, basePlanID, offerID string) (bool, error)
	GetRegionalOfferPhaseConfig(ctx context.Context, packageName, productID, basePlanID, offerID string, phaseIndex int, regionCode string) (*models.RegionalSubscriptionOfferPhaseConfig, error)
	GetOtherRegionsOfferPhaseConfig(ctx context.Context, packageName, productID, basePlanID, offerID string, phaseIndex int) (*models.OtherRegionsSubscriptionOfferPhaseConfig, error)
}

// Concrete implementation
type subscriptionCatalogRepository struct {
	db        *gorm.DB
	batchSize int
}

// ✅ Ensure struct implements the interface at compile-time
var _ SubscriptionCatalogRepository = (*subscriptionCatalogRepository)(nil)

// NewSubscriptionCatalogRepository creates a new instance.
func NewSubscriptionCatalogRepository(db *gorm.DB, batchSize int) SubscriptionCatalogRepository {
	return &subscriptionCatalogRepository{
		db:        db,
		batchSize: batchSize,
	}
}
