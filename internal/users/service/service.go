package service

import (
	"context"
	"fmt"

	"subsnotifpro-go/internal/users/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService interface {
	CreateAppUserIfNotExists(ctx context.Context, tx *gorm.DB, obfuscatedID, platform string) (uuid.UUID, error)
	LinkGoogleAccount(
		ctx context.Context,
		tx *gorm.DB,
		userID uuid.UUID,
		googleAccountID uuid.UUID,
	) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

// CreateAppUserIfNotExists creates a new AppUser if one with the given obfuscatedID doesn't exist.
func (s *userService) CreateAppUserIfNotExists(ctx context.Context, tx *gorm.DB, obfuscatedID, platform string) (uuid.UUID, error) {
	return s.repo.CreateAppUserIfNotExists(ctx, tx, obfuscatedID, platform)
}

// In internal/users/service/service.go
func (s *userService) LinkGoogleAccount(
	ctx context.Context,
	tx *gorm.DB,
	userID uuid.UUID,
	googleAccountID uuid.UUID,
) error {
	if userID == uuid.Nil || googleAccountID == uuid.Nil {
		return fmt.Errorf("invalid IDs provided")
	}
	return s.repo.LinkGoogleAccount(ctx, tx, userID, googleAccountID)
}
