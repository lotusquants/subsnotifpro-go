package repository

import (
	"subsnotifpro-go/internal/tenant/models"

	"gorm.io/gorm"
)

// ITenantRepository defines the interface for tenant repository
type ITenantRepository interface {
	Create(tx *gorm.DB, tenant *models.Tenant) error
	GetByID(tx *gorm.DB, id string) (*models.Tenant, error)
}

// tenantRepository is the concrete implementation of ITenantRepository
type tenantRepository struct{}

// NewTenantRepository returns an instance of ITenantRepository
func NewTenantRepository() ITenantRepository {
	return &tenantRepository{}
}

// Create inserts a new tenant record
func (r *tenantRepository) Create(tx *gorm.DB, tenant *models.Tenant) error {
	return tx.Create(tenant).Error
}

// GetByID fetches a tenant by its ID
func (r *tenantRepository) GetByID(tx *gorm.DB, id string) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := tx.First(&tenant, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}
