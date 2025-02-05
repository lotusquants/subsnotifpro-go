package api_tests

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"subsnotifpro-go/database"
	"subsnotifpro-go/internal/google_play"
	"subsnotifpro-go/models"
	"testing"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestUploadServiceAccount tests the upload functionality for Google Play service accounts.
func TestUploadServiceAccount(t *testing.T) {

	// Ensure database is connected
	database.SetupTestDatabase()

	// Setup Gin for testing
	r := gin.Default()
	r.POST("/google-play/upload-service-account", google_play.UploadServiceAccountHandler) // Use the actual handler

	// Create a mock service account file
	testFile := "./test_service_account.json"
	err := os.WriteFile(testFile, []byte(`{"client_email": "test@google.com"}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile) // Clean up after test

	// Prepare the file upload form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", testFile)
	if err != nil {
		t.Fatalf("Error creating form file: %v", err)
	}

	// Write the file data to the form
	fileData, _ := os.ReadFile(testFile)
	part.Write(fileData)
	writer.Close()

	// Make the POST request to upload the file
	req, err := http.NewRequest(http.MethodPost, "/google-play/upload-service-account", &buf)
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}

	// Set the content type for multipart form data
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Perform the request and check the response
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	// Assert the response code
	assert.Equal(t, http.StatusOK, resp.Code)

	// Verify the service account file is saved in the directory
	secretsDir := "secrets/google_play/"
	expectedFilePath := fmt.Sprintf("%s%s", secretsDir, "test_service_account.json") // No './' here
	if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) {
		t.Fatalf("File not found: %v", err)
	}

	// Verify the service account was added to the database
	var serviceAccount models.GooglePlayServiceAccount
	database.DB.First(&serviceAccount) // Assuming GORM and database is set up

	// Compare file paths, ignoring extra './'
	assert.Equal(t, "test_service_account.json", serviceAccount.FileName)        // Check file name
	assert.True(t, strings.HasSuffix(serviceAccount.FilePath, expectedFilePath)) // Ensure the file path matches, ignoring leading './'
	assert.False(t, serviceAccount.Validated)                                    // Check validated default value
	assert.NotZero(t, serviceAccount.LastChecked)                                // Check if LastChecked is populated
	assert.Empty(t, serviceAccount.DeletedAt)                                    // Ensure it's not deleted

	// Clean up the test file after validation
	os.Remove(expectedFilePath)
}
