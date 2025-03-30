package service

import (
	"context"

	"subsnotifpro-go/internal/users/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService interface {
	CreateAppUserIfNotExists(ctx context.Context, tx *gorm.DB, obfuscatedID, platform string) (uuid.UUID, error)
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
