package mocks

import (
	"context"

	"subsnotifpro-go/internal/playstore/settings/models"

	"github.com/stretchr/testify/mock"
)

// ✅ MockPlaystoreSettingsRepository is a mock implementation of PlaystoreSettingsRepository
type MockPlaystoreSettingsRepository struct {
	mock.Mock
}

// ✅ Mock SaveServiceAccount
func (m *MockPlaystoreSettingsRepository) SaveServiceAccount(ctx context.Context, filePath string, fileName string) error {
	args := m.Called(ctx, filePath, fileName)
	return args.Error(0)
}

// ✅ Mock GetLatestServiceAccount
func (m *MockPlaystoreSettingsRepository) GetLatestServiceAccount(ctx context.Context) (models.GooglePlayServiceAccount, error) {
	args := m.Called(ctx)
	account, _ := args.Get(0).(models.GooglePlayServiceAccount) // ✅ Ensure type safety
	return account, args.Error(1)
}

// ✅ Mock UpdateServiceAccountValidationStatus
func (m *MockPlaystoreSettingsRepository) UpdateServiceAccountValidationStatus(ctx context.Context, isValid bool) error {
	args := m.Called(ctx, isValid)
	return args.Error(0)
}

// ✅ Mock DeleteExistingServiceAccount
func (m *MockPlaystoreSettingsRepository) DeleteExistingServiceAccount(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// ✅ Mock UpdatePackageName
func (m *MockPlaystoreSettingsRepository) UpdatePackageName(ctx context.Context, packageName string) error {
	args := m.Called(ctx, packageName)
	return args.Error(0)
}

// ✅ Mock FetchPackageName
func (m *MockPlaystoreSettingsRepository) FetchPackageName(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

// ✅ Mock GetSettings
func (m *MockPlaystoreSettingsRepository) GetSettings(ctx context.Context) (models.GooglePlaySettings, error) {
	args := m.Called(ctx)
	settings, _ := args.Get(0).(models.GooglePlaySettings) // ✅ Ensure type safety
	return settings, args.Error(1)
}

// ✅ Mock SaveSettings
func (m *MockPlaystoreSettingsRepository) SaveSettings(ctx context.Context, settings models.GooglePlaySettings) error {
	args := m.Called(ctx, settings)
	return args.Error(0)
}

// ✅ Mock DeletePackageName
func (m *MockPlaystoreSettingsRepository) DeletePackageName(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
