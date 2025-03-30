package repository

import (
    "context"

    "subsnotifpro-go/internal/project/models"
    "gorm.io/gorm"
)

type Repository interface {
    Create(ctx context.Context, project *models.Project) error
    GetByOrgID(ctx context.Context, orgID string) ([]*models.Project, error)
}

type repoImpl struct {
    db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
    return &repoImpl{db: db}
}

func (r *repoImpl) Create(ctx context.Context, project *models.Project) error {
    return r.db.WithContext(ctx).Create(project).Error
}

func (r *repoImpl) GetByOrgID(ctx context.Context, orgID string) ([]*models.Project, error) {
    var projects []*models.Project
    err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).Find(&projects).Error
    return projects, err
}
