package repository

import (
	"context"
	"subsnotifpro-go/internal/organization/models"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, org *models.Organization) error
	GetByID(ctx context.Context, id string) (*models.Organization, error)
	GetByOwnerID(ctx context.Context, ownerID string) (*models.Organization, error)
}

type repoImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repoImpl{db}
}

func (r *repoImpl) Create(ctx context.Context, org *models.Organization) error {
	return r.db.WithContext(ctx).Create(org).Error
}

func (r *repoImpl) GetByID(ctx context.Context, id string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.WithContext(ctx).First(&org, "id = ?", id).Error
	return &org, err
}

func (r *repoImpl) GetByOwnerID(ctx context.Context, ownerID string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.WithContext(ctx).First(&org, "owner_id = ?", ownerID).Error
	return &org, err
}
