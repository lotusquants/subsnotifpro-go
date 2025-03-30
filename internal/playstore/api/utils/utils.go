package utils

import (
	"errors"

	"google.golang.org/api/googleapi"
)

// **IsInvalidTokenError checks if an error message indicates an invalid purchase token.**
func IsInvalidTokenError(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		return apiErr.Code == 400 // Example for invalid token error
	}
	return false
}
