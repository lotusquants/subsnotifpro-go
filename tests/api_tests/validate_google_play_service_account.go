package api_tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"subsnotifpro-go/database"
	"subsnotifpro-go/models"
	"subsnotifpro-go/routes"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestValidateServiceAccountAPI tests the validation of a service account
func TestValidateServiceAccountAPI(t *testing.T) {
	// Set up a test router
	gin.SetMode(gin.TestMode)
	router := routes.SetupRouter()

	// Create a temporary service account JSON file
	tempFilePath := "secrets/google_play/test-service-account.json"
	mockServiceAccount := `{
		"type": "service_account",
		"project_id": "test-project",
		"private_key_id": "12345",
		"private_key": "-----BEGIN PRIVATE KEY-----\\nMIIEvQIBADAN...",
		"client_email": "test@developer.gserviceaccount.com"
	}`

	// Ensure secrets directory exists
	_ = os.MkdirAll("secrets/google_play/", os.ModePerm)

	// Write the mock JSON to the test file
	err := os.WriteFile(tempFilePath, []byte(mockServiceAccount), 0644)
	assert.NoError(t, err)

	// Insert a mock entry into the database
	mockEntry := models.GooglePlayServiceAccount{
		FileName:  "test-service-account.json",
		FilePath:  tempFilePath,
		Validated: false,
	}
	database.DB.Create(&mockEntry)

	// Create a request to validate the service account
	req, _ := http.NewRequest("GET", "/api/google-play/validate-service-account", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Assert response
	assert.Equal(t, http.StatusOK, resp.Code, "Expected HTTP 200 OK")
	assert.Contains(t, resp.Body.String(), "Service account validated successfully")

	// Clean up after test
	_ = os.Remove(tempFilePath)
	database.DB.Where("file_name = ?", "test-service-account.json").Delete(&models.GooglePlayServiceAccount{})
}
