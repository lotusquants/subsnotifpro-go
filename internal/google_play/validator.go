package google_play

import (
	"encoding/json"
	"errors"
	"os"
)

// ValidateServiceAccountJSONStructure ensures the uploaded service account JSON file has valid fields
func ValidateServiceAccountJSONStructure(filePath string) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return errors.New("❌ Error reading service account file: " + err.Error())
	}

	var data map[string]interface{}
	if err := json.Unmarshal(file, &data); err != nil {
		return errors.New("❌ Invalid JSON format: " + err.Error())
	}

	// Check for required keys in the JSON file
	requiredKeys := []string{"type", "project_id", "private_key_id", "private_key", "client_email"}
	for _, key := range requiredKeys {
		if _, exists := data[key]; !exists {
			return errors.New("❌ Missing required field: " + key)
		}
	}

	return nil
}
