package validator

import (
	"encoding/json"
	"errors"
	"os"
)

// ValidateServiceAccountJSONStructure ensures the uploaded JSON file is valid
func ValidateServiceAccountJSONStructure(filePath string) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return errors.New("❌ Error reading service account file: " + err.Error())
	}

	var data map[string]interface{}
	if err := json.Unmarshal(file, &data); err != nil {
		return errors.New("❌ Invalid JSON format: " + err.Error())
	}

	return nil
}
