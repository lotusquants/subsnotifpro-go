package api_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"subsnotifpro-go/database"
	playstoresettings "subsnotifpro-go/internal/google_playstore/google_playstore_settings"
	"subsnotifpro-go/internal/messaging"
	"subsnotifpro-go/routes"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var testRouter *gin.Engine

// Setup before running tests
func TestMain(m *testing.M) {

	// Ensure database is connected
	database.SetupTestDatabase()
	// ✅ Get a **single** RabbitMQ Channel
	ch, err := messaging.GetChannel(context.Background())
	if err != nil {
		log.Fatal("❌ Failed to connect to RabbitMQ:", err)
		return
	}
	defer func() {
		log.Println("🚦 Closing RabbitMQ connection...")
		_ = ch.Close()
	}()

	// ✅ Initialize RabbitMQ (Queues, Exchanges, Bindings)
	messaging.InitializeRabbitMQ(ch)

	gin.SetMode(gin.TestMode)
	testRouter = routes.SetupRouter(ch) // Ensure you call your setup router function
	os.Exit(m.Run())
}

// 🟢 1️⃣ Test Upload Service Account API
func TestUploadServiceAccount(t *testing.T) {
	// Prepare a valid JSON file
	filePath := "test_service_account.json"
	fileContent := `{"type":"service_account","project_id":"test-project","private_key_id":"test-key","private_key":"test-key-content","client_email":"test@test.com"}`
	err := os.WriteFile(filePath, []byte(fileContent), 0644)
	assert.NoError(t, err)
	defer os.Remove(filePath) // Clean up test file

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filePath)
	io.Copy(part, bytes.NewReader([]byte(fileContent)))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/google-play/upload-service-account", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// 🔴 2️⃣ Test Upload Service Account Without File
func TestUploadServiceAccountWithoutFile(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/google-play/upload-service-account", nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// 🟢 3️⃣ Test Validate Service Account API
func TestValidateServiceAccount(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/google-play/validate-service-account", nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// 🔴 4️⃣ Test Validate Without Service Account
func TestValidateWithoutServiceAccount(t *testing.T) {
	playstoresettings.DeleteExistingServiceAccount() // Ensure no service account exists

	req := httptest.NewRequest("GET", "/api/google-play/validate-service-account", nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// 🟢 6️⃣ Test Get Service Account Status API

// 🔴 7️⃣ Test Get Service Account Status Without Any Account
func TestGetServiceAccountStatusWithoutAccount(t *testing.T) {
	playstoresettings.DeleteExistingServiceAccount()

	req := httptest.NewRequest("GET", "/api/google-play/service-account-status", nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// 🔴 9️⃣ Test Delete When No Service Account Exists
func TestDeleteServiceAccountWithoutExisting(t *testing.T) {
	playstoresettings.DeleteExistingServiceAccount()

	req := httptest.NewRequest("DELETE", "/api/google-play/delete-service-account", nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// 🟢 🔟 Test Set Package Name API
func TestSetPackageName(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"package_name": "com.test.app"})
	req := httptest.NewRequest("POST", "/api/google-play/set-package-name", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// 🟢 1️⃣2️⃣ Test Get Package Name API
func TestGetPackageName(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/google-play/get-package-name", nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
