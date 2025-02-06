package api_tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"subsnotifpro-go/routes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFetchServiceAccountStatus tests the API to fetch stored service account status
func TestFetchServiceAccountStatus(t *testing.T) {
	// Initialize router
	router := routes.SetupRouter()

	// Make a request to fetch the service account status
	req, _ := http.NewRequest("GET", "/api/google-play/service-account-status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response status code
	assert.Equal(t, http.StatusOK, w.Code, "Expected HTTP 200 OK")

	// Parse response JSON
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Response should be valid JSON")

	// Ensure necessary fields exist in response

	assert.Contains(t, response, "validated", "Response must contain validated status")
	assert.Contains(t, response, "last_checked", "Response must contain last_checked timestamp")
	assert.Contains(t, response, "file_path", "Response must contain file_path")

	// Ensure the response values match expected types

	assert.IsType(t, true, response["validated"], "validated should be a boolean")
	assert.IsType(t, "", response["last_checked"], "last_checked should be a string")
	assert.IsType(t, "", response["file_path"], "file_path should be a string")
}
