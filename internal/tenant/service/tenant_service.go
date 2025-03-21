package service

import (
	"context"
	"fmt"

	"subsnotifpro-go/internal/tenant/models"
	"subsnotifpro-go/internal/tenant/repository"

	"gorm.io/gorm"
)

// TenantService defines the interface for tenant service
type ITenantService interface {
	CreateTenant(ctx context.Context, name string) (*models.Tenant, error)
	GetTenantByID(ctx context.Context, id string) (*models.Tenant, error)
}

// tenantService implements TenantService interface
type tenantService struct {
	db   *gorm.DB
	repo repository.ITenantRepository
}

// NewTenantService creates a new instance of tenantService
func NewTenantService(db *gorm.DB, repo repository.ITenantRepository) ITenantService {
	return &tenantService{db: db, repo: repo}
}

// CreateTenant creates a new tenant with transaction
func (s *tenantService) CreateTenant(ctx context.Context, name string) (*models.Tenant, error) {
	tenant := &models.Tenant{Name: name}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.repo.Create(tx, tenant)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}
	return tenant, nil
}

// GetTenantByID retrieves a tenant by ID using a transaction
func (s *tenantService) GetTenantByID(ctx context.Context, id string) (*models.Tenant, error) {
	var tenant *models.Tenant
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		tenant, err = s.repo.GetByID(tx, id)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tenant: %w", err)
	}
	return tenant, nil
}
