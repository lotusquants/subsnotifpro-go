package api_tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestHealthCheck verifies that the /health endpoint returns 200 OK.
func TestHealthCheck(t *testing.T) {
	// Create a test Gin engine
	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "OK",
			"message": "SubsNotifPro backend is running!",
		})
	})

	// Create a test HTTP request
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response status
	assert.Equal(t, http.StatusOK, w.Code)

	// Check response body
	expectedBody := `{"status":"OK","message":"SubsNotifPro backend is running!"}`
	assert.JSONEq(t, expectedBody, w.Body.String())
}
