package service

import (
	"context"
	"strings"
	"subsnotifpro-go/internal/organization/models"
	"subsnotifpro-go/internal/organization/repository"

	"github.com/gosimple/slug"
)

type Service interface {
	CreateOrganization(ctx context.Context, name string, ownerID string) (*models.Organization, error)
	GetOrganizationByOwner(ctx context.Context, ownerID string) (*models.Organization, error)
}

type serviceImpl struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &serviceImpl{repo}
}

func (s *serviceImpl) CreateOrganization(ctx context.Context, name string, ownerID string) (*models.Organization, error) {
	org := &models.Organization{
		Name:    name,
		Slug:    slug.Make(strings.ToLower(name)),
		OwnerID: ownerID,
		Plan:    "free",
	}
	err := s.repo.Create(ctx, org)
	return org, err
}

func (s *serviceImpl) GetOrganizationByOwner(ctx context.Context, ownerID string) (*models.Organization, error) {
	return s.repo.GetByOwnerID(ctx, ownerID)
}
