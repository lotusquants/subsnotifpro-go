package service

import (
	"context"
	"strings"

	"github.com/gosimple/slug"
	"subsnotifpro-go/internal/project/models"
	"subsnotifpro-go/internal/project/repository"
)

type Service interface {
	CreateProject(ctx context.Context, orgID, name string) (*models.Project, error)
	GetProjectsByOrg(ctx context.Context, orgID string) ([]*models.Project, error)
}

type serviceImpl struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &serviceImpl{repo: repo}
}

func (s *serviceImpl) CreateProject(ctx context.Context, orgID, name string) (*models.Project, error) {
	project := &models.Project{
		Name:           name,
		Slug:           slug.Make(strings.ToLower(name)),
		OrganizationID: orgID,
	}
	err := s.repo.Create(ctx, project)
	return project, err
}

func (s *serviceImpl) GetProjectsByOrg(ctx context.Context, orgID string) ([]*models.Project, error) {
	return s.repo.GetByOrgID(ctx, orgID)
}
