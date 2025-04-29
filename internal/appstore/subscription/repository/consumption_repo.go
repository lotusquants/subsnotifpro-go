// repository/consumption_request_repo.go
package repository

import (
	"context"
	"subsnotifpro-go/internal/appstore/subscription/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConsumptionRequestRepository interface {
	Create(ctx context.Context, tx *gorm.DB, req *models.ConsumptionRequest) error
	Update(ctx context.Context, tx *gorm.DB, req *models.ConsumptionRequest) error
	FindBySubscriptionID(ctx context.Context, tx *gorm.DB, subscriptionID uuid.UUID) ([]*models.ConsumptionRequest, error)
}

type consumptionRequestRepo struct {
	db *gorm.DB
}

func NewConsumptionRequestRepository(db *gorm.DB) ConsumptionRequestRepository {
	return &consumptionRequestRepo{db: db}
}

func (r *consumptionRequestRepo) Create(ctx context.Context, tx *gorm.DB, req *models.ConsumptionRequest) error {
	return tx.WithContext(ctx).Create(req).Error
}

func (r *consumptionRequestRepo) Update(ctx context.Context, tx *gorm.DB, req *models.ConsumptionRequest) error {
	return tx.WithContext(ctx).Save(req).Error
}

func (r *consumptionRequestRepo) FindBySubscriptionID(ctx context.Context, tx *gorm.DB, subscriptionID uuid.UUID) ([]*models.ConsumptionRequest, error) {
	var requests []*models.ConsumptionRequest
	err := tx.WithContext(ctx).Where("subscription_id = ?", subscriptionID).Find(&requests).Error
	return requests, err
}
