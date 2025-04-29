package service

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"subsnotifpro-go/internal/appstore/settings/models"
	"subsnotifpro-go/internal/appstore/settings/repository"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type AppStoreSettingsService interface {
	GetSettings(bundleID string) (*models.AppStoreSettingsResponse, error)
	GetAllSettings() ([]models.AppStoreSettingsResponse, error)
	UpdateSettings(request *models.AppStoreSettingsRequest) (*models.AppStoreSettingsResponse, error)
	DeleteSettings(bundleID string) error

	GenerateAppStoreJWT(bundleID string) (string, error)

	SendTestNotification(ctx context.Context, bundleID string, sandbox bool) (*models.TestNotificationResponse, error)
}

type appStoreSettingsService struct {
	repo           repository.AppStoreSettingsRepository
	httpClient     *http.Client
	prodBaseURL    string
	sandboxBaseURL string
}

func NewAppStoreSettingsService(repo repository.AppStoreSettingsRepository) AppStoreSettingsService {
	return &appStoreSettingsService{
		repo:           repo,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
		prodBaseURL:    "https://api.storekit.itunes.apple.com",
		sandboxBaseURL: "https://api.storekit-sandbox.itunes.apple.com",
	}
}

func (s *appStoreSettingsService) GetSettings(bundleID string) (*models.AppStoreSettingsResponse, error) {
	settings, err := s.repo.GetSettings(bundleID)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, nil
	}

	return &models.AppStoreSettingsResponse{
		ID:       settings.ID,
		BundleID: settings.BundleID,
		IssuerID: settings.IssuerID,
		KeyID:    settings.KeyID,
		P8Key:    settings.P8Key,
	}, nil
}

func (s *appStoreSettingsService) GetAllSettings() ([]models.AppStoreSettingsResponse, error) {
	settingsList, err := s.repo.GetAllSettings()
	if err != nil {
		return nil, err
	}

	var responses []models.AppStoreSettingsResponse
	for _, settings := range settingsList {
		responses = append(responses, models.AppStoreSettingsResponse{
			ID:       settings.ID,
			BundleID: settings.BundleID,
			IssuerID: settings.IssuerID,
			KeyID:    settings.KeyID,
		})
	}

	return responses, nil
}

func (s *appStoreSettingsService) UpdateSettings(request *models.AppStoreSettingsRequest) (*models.AppStoreSettingsResponse, error) {
	settings := &models.AppStoreSettings{
		BundleID: request.BundleID,
		IssuerID: request.IssuerID,
		KeyID:    request.KeyID,
		P8Key:    request.P8Key,
	}

	updated, err := s.repo.CreateOrUpdateSettings(settings)
	if err != nil {
		return nil, err
	}

	return &models.AppStoreSettingsResponse{
		ID:       updated.ID,
		BundleID: updated.BundleID,
		IssuerID: updated.IssuerID,
		KeyID:    updated.KeyID,
	}, nil
}

func (s *appStoreSettingsService) DeleteSettings(bundleID string) error {
	return s.repo.DeleteSettings(bundleID)
}

func (s *appStoreSettingsService) GenerateAppStoreJWT(bundleID string) (string, error) {
	// Get settings for the bundle ID
	settings, err := s.GetSettings(bundleID)
	if err != nil {
		return "", fmt.Errorf("failed to get settings: %w", err)
	}
	if settings == nil {
		return "", errors.New("no settings found for this bundle ID")
	}

	// Parse the PKCS#8 private key
	block, _ := pem.Decode([]byte(settings.P8Key))
	if block == nil {
		return "", errors.New("failed to parse PEM block containing the private key")
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	ecdsaKey, ok := privKey.(*ecdsa.PrivateKey)
	if !ok {
		return "", errors.New("private key is not of type ECDSA")
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": settings.IssuerID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(20 * time.Minute).Unix(),
		"aud": "appstoreconnect-v1",
		"bid": bundleID,
	})

	// Set the key ID header
	token.Header["kid"] = settings.KeyID

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(ecdsaKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (s *appStoreSettingsService) SendTestNotification(ctx context.Context, bundleID string, sandbox bool) (*models.TestNotificationResponse, error) {
	// Generate JWT token
	token, err := s.GenerateAppStoreJWT(bundleID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	// Determine the base URL
	baseURL := s.prodBaseURL
	if sandbox {
		baseURL = s.sandboxBaseURL
	}

	// Create request
	url := fmt.Sprintf("%s/inApps/v1/notifications/test", baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle response
	if resp.StatusCode != http.StatusOK {
		var errorResponse struct {
			ErrorMessage string `json:"errorMessage"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, errorResponse.ErrorMessage)
		}
		return nil, fmt.Errorf("API returned unexpected status: %d", resp.StatusCode)
	}

	// Parse successful response
	var result models.TestNotificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
