package utils

import (
	"encoding/json"
	"net/http"
	"subsnotifpro-go/internal/logger"

	"github.com/gin-gonic/gin"
)

// ------------------------------------
// Helper Functions for JSON Responses
// ------------------------------------

// writeJSONResponse writes a generic JSON response with given status and payload.
func WriteJSONResponse(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Log.Errorf("Failed to write JSON response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// writeErrorResponse simplifies writing JSON error responses.
func WriteErrorResponse(w http.ResponseWriter, status int, message string) {
	WriteJSONResponse(w, status, map[string]string{"error": message})
}

func WriteGinErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{"error": message})
}

func WriteGinJSONResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}
