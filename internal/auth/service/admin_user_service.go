package service

import (
	"context"
	"fmt"
	"subsnotifpro-go/internal/auth/models"
	"subsnotifpro-go/internal/auth/repository"
	"subsnotifpro-go/internal/auth/utils"

	"gorm.io/gorm"
)

type IAdminService interface {
	Register(ctx context.Context, email, password, tenantID string) (*models.AdminUser, error)
	Login(ctx context.Context, email, password string) (string, error)
}

type adminService struct {
	db   *gorm.DB
	repo repository.IAdminRepository
}

func NewAdminService(db *gorm.DB, repo repository.IAdminRepository) IAdminService {
	return &adminService{db: db, repo: repo}
}

func (s *adminService) Register(ctx context.Context, email, password, tenantID string) (*models.AdminUser, error) {
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}
	admin := &models.AdminUser{
		Email:    email,
		Password: hashedPassword,
		TenantID: tenantID,
		Role:     "admin",
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.repo.Create(tx, admin)
	})
	return admin, err
}

func (s *adminService) Login(ctx context.Context, email, password string) (string, error) {
	var admin *models.AdminUser
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		admin, err = s.repo.FindByEmail(tx, email)
		return err
	})
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	if !utils.CheckPasswordHash(password, admin.Password) {
		return "", fmt.Errorf("invalid credentials")
	}

	return utils.GenerateJWT(admin.ID, admin.TenantID, admin.Role)
}
