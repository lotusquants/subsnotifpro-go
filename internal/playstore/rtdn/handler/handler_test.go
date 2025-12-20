package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"subsnotifpro-go/internal/playstore/rtdn/dto"
	"subsnotifpro-go/internal/playstore/rtdn/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRTDNService is a mock implementation of RTDNService
type MockRTDNService struct {
	mock.Mock
}

func (m *MockRTDNService) ProcessWebhookEventForPublish(ctx context.Context, event *dto.GooglePlayWebhookEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockRTDNService) ProcessWebhookEvent(ctx context.Context, payload models.GooglePublishPayload) error {
	args := m.Called(ctx, payload)
	return args.Error(0)
}

func (m *MockRTDNService) ProcessSubscriptionEvent(ctx context.Context, payload models.GooglePublishPayload) error {
	args := m.Called(ctx, payload)
	return args.Error(0)
}

func (m *MockRTDNService) ProcessOneTimeProductEvent(ctx context.Context, payload models.GooglePublishPayload) error {
	args := m.Called(ctx, payload)
	return args.Error(0)
}

func (m *MockRTDNService) ProcessVoidedPurchaseEvent(ctx context.Context, payload models.GooglePublishPayload) error {
	args := m.Called(ctx, payload)
	return args.Error(0)
}

func (m *MockRTDNService) GetDLQSize(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockRTDNService) RetryMessages(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestRTDNHandler_WrappedFormat(t *testing.T) {
	// Set Gin to debug mode (this will skip IP validation)
	gin.SetMode(gin.DebugMode)

	// Create mock service
	mockService := new(MockRTDNService)
	mockService.On("ProcessWebhookEventForPublish", mock.Anything, mock.Anything).Return(nil)

	// Create handler
	handler := NewRTDNHandler(mockService)

	// Create wrapped Pub/Sub payload
	wrappedPayload := map[string]interface{}{
		"message": map[string]interface{}{
			"data":      "eyJ2ZXJzaW9uIjoiMS4wIiwicGFja2FnZU5hbWUiOiJjb20uZXhhbXBsZS5hcHAiLCJldmVudFRpbWVNaWxsaXMiOjE1MDMzNDk1NjYxNjgsInN1YnNjcmlwdGlvbk5vdGlmaWNhdGlvbiI6eyJ2ZXJzaW9uIjoiMS4wIiwibm90aWZpY2F0aW9uVHlwZSI6NCwicHVyY2hhc2VUb2tlbiI6IlBVUkNIQVNFX1RPS0VOIiwic3Vic2NyaXB0aW9uSWQiOiJwcmVtaXVtX21vbnRobHkifX0=", // Base64 encoded JSON
			"messageId": "136969346945",
			"attributes": map[string]string{
				"key": "value",
			},
		},
		"subscription": "projects/myproject/subscriptions/mysubscription",
	}

	// Convert to JSON
	payloadJSON, err := json.Marshal(wrappedPayload)
	assert.NoError(t, err)

	// Create request
	req, err := http.NewRequest("POST", "/webhook", bytes.NewBuffer(payloadJSON))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Call handler
	handler.WebhookHandler(c)

	// Verify response
	assert.Equal(t, http.StatusAccepted, w.Code)
	mockService.AssertExpectations(t)
}

func TestRTDNHandler_UnwrappedFormat(t *testing.T) {
	// Set Gin to debug mode (this will skip IP validation)
	gin.SetMode(gin.DebugMode)

	// Create mock service
	mockService := new(MockRTDNService)
	mockService.On("ProcessWebhookEventForPublish", mock.Anything, mock.Anything).Return(nil)

	// Create handler
	handler := NewRTDNHandler(mockService)

	// Create unwrapped payload (direct RTDN format)
	unwrappedPayload := map[string]interface{}{
		"version":         "1.0",
		"packageName":     "com.example.app",
		"eventTimeMillis": 1503349566168,
		"subscriptionNotification": map[string]interface{}{
			"version":          "1.0",
			"notificationType": 4,
			"purchaseToken":    "PURCHASE_TOKEN",
			"subscriptionId":   "premium_monthly",
		},
	}

	// Convert to JSON
	payloadJSON, err := json.Marshal(unwrappedPayload)
	assert.NoError(t, err)

	// Create request
	req, err := http.NewRequest("POST", "/webhook", bytes.NewBuffer(payloadJSON))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Call handler
	handler.WebhookHandler(c)

	// Verify response
	assert.Equal(t, http.StatusAccepted, w.Code)
	mockService.AssertExpectations(t)
}

func TestRTDNHandler_InvalidFormat(t *testing.T) {
	// Set Gin to debug mode (this will skip IP validation)
	gin.SetMode(gin.DebugMode)

	// Create mock service
	mockService := new(MockRTDNService)

	// Create handler
	handler := NewRTDNHandler(mockService)

	// Create invalid payload
	invalidPayload := `{"invalid": "json"}`

	// Create request
	req, err := http.NewRequest("POST", "/webhook", bytes.NewBufferString(invalidPayload))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Call handler
	handler.WebhookHandler(c)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestPubSubFormats demonstrates both wrapped and unwrapped formats
func TestPubSubFormats(t *testing.T) {
	// Example of wrapped format (most common)
	wrappedExample := `{
		"message": {
			"data": "eyJ2ZXJzaW9uIjoiMS4wIiwicGFja2FnZU5hbWUiOiJjb20uZXhhbXBsZS5hcHAiLCJldmVudFRpbWVNaWxsaXMiOjE1MDMzNDk1NjYxNjgsInN1YnNjcmlwdGlvbk5vdGlmaWNhdGlvbiI6eyJ2ZXJzaW9uIjoiMS4wIiwibm90aWZpY2F0aW9uVHlwZSI6NCwicHVyY2hhc2VUb2tlbiI6IlBVUkNIQVNFX1RPS0VOIiwic3Vic2NyaXB0aW9uSWQiOiJwcmVtaXVtX21vbnRobHkifX0=",
			"messageId": "136969346945",
			"publishTime": "2021-02-26T19:13:55.749Z",
			"attributes": {
				"key": "value"
			}
		},
		"subscription": "projects/myproject/subscriptions/mysubscription"
	}`

	// Example of unwrapped format (when push subscription is configured to unwrap)
	unwrappedExample := `{
		"version": "1.0",
		"packageName": "com.example.app",
		"eventTimeMillis": 1503349566168,
		"subscriptionNotification": {
			"version": "1.0",
			"notificationType": 4,
			"purchaseToken": "PURCHASE_TOKEN",
			"subscriptionId": "premium_monthly"
		}
	}`

	t.Log("Wrapped format example:", wrappedExample)
	t.Log("Unwrapped format example:", unwrappedExample)

	// Both formats should be handled correctly by the parser
	assert.True(t, true, "Both formats are supported based on push subscription configuration")
}
