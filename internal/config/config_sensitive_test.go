// Package config implements tests for adapter configuration and sensitive data handling
package config

import (
	"testing"
)

// TestGetAdapterConfig tests retrieving adapter-specific configuration
func TestGetAdapterConfig(t *testing.T) {
	config := &Config{
		Adapters: map[string]map[string]interface{}{
			"test-adapter": {
				"api_key": "secret123",
				"endpoint": "https://api.example.com",
			},
		},
	}
	
	adapterConfig := config.GetAdapterConfig("test-adapter")
	if adapterConfig == nil {
		t.Fatal("Expected adapter config, got nil")
	}
	
	if apiKey, exists := adapterConfig["api_key"]; !exists || apiKey != "secret123" {
		t.Errorf("Expected api_key 'secret123', got '%v'", apiKey)
	}
	
	if endpoint, exists := adapterConfig["endpoint"]; !exists || endpoint != "https://api.example.com" {
		t.Errorf("Expected endpoint 'https://api.example.com', got '%v'", endpoint)
	}
	
	// Test with non-existent adapter
	emptyConfig := config.GetAdapterConfig("non-existent")
	if emptyConfig == nil {
		t.Fatal("Expected empty adapter config, got nil")
	}
	if len(emptyConfig) != 0 {
		t.Error("Expected empty adapter config")
	}
}

// TestRedactSensitiveFields tests redaction of sensitive fields
func TestRedactSensitiveFields(t *testing.T) {
	config := &Config{
		Adapters: map[string]map[string]interface{}{
			"test-adapter": {
				"api_key":      "secret123",
				"password":     "mypassword",
				"token":        "abc123",
				"normal_field": "normal_value",
				"url":          "https://example.com",
			},
		},
	}
	
	redacted := config.RedactSensitiveFields()
	
	adapterConfig := redacted.GetAdapterConfig("test-adapter")
	
	// Check that sensitive fields are redacted
	if apiKey, exists := adapterConfig["api_key"]; !exists || apiKey != "[REDACTED]" {
		t.Errorf("Expected redacted api_key, got '%v'", apiKey)
	}
	
	if password, exists := adapterConfig["password"]; !exists || password != "[REDACTED]" {
		t.Errorf("Expected redacted password, got '%v'", password)
	}
	
	if token, exists := adapterConfig["token"]; !exists || token != "[REDACTED]" {
		t.Errorf("Expected redacted token, got '%v'", token)
	}
	
	// Check that non-sensitive fields are preserved
	if normalField, exists := adapterConfig["normal_field"]; !exists || normalField != "normal_value" {
		t.Errorf("Expected preserved normal_field, got '%v'", normalField)
	}
	
	if url, exists := adapterConfig["url"]; !exists || url != "https://example.com" {
		t.Errorf("Expected preserved url, got '%v'", url)
	}
}

// TestValidateAdapterConfig tests validation of adapter configuration
func TestValidateAdapterConfig(t *testing.T) {
	config := &Config{
		Adapters: map[string]map[string]interface{}{
			"test-adapter": {
				"api_key":  "secret123",
				"endpoint": "https://api.example.com",
			},
		},
	}
	
	// Test with valid required fields
	requiredFields := []string{"api_key", "endpoint"}
	err := config.ValidateAdapterConfig("test-adapter", requiredFields)
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}
	
	// Test with missing required field
	requiredFields = []string{"api_key", "missing_field"}
	err = config.ValidateAdapterConfig("test-adapter", requiredFields)
	if err == nil {
		t.Error("Expected error for missing required field")
	}
	
	// Test with non-existent adapter
	err = config.ValidateAdapterConfig("non-existent", []string{"field"})
	if err == nil {
		t.Error("Expected error for non-existent adapter")
	}
}

// TestIsSensitiveField tests the sensitive field detection
func TestIsSensitiveField(t *testing.T) {
	sensitiveFields := []string{
		"password", "secret", "token", "key", "credential", "auth", "api_key",
		"access_token", "refresh_token", "client_secret", "private", "cert", "jwt",
		"PASSWORD", "Secret", "API_KEY", // Test case insensitivity
	}
	
	for _, field := range sensitiveFields {
		if !isSensitiveField(field) {
			t.Errorf("Expected field '%s' to be detected as sensitive", field)
		}
	}
	
	nonSensitiveFields := []string{
		"url", "endpoint", "name", "description", "version", "path",
	}

	for _, field := range nonSensitiveFields {
		if isSensitiveField(field) {
			t.Errorf("Expected field '%s' to not be detected as sensitive", field)
		}
	}
}

// TestIsSensitiveFieldCaseInsensitive tests the sensitive field detection with case insensitive matching
func TestIsSensitiveFieldCaseInsensitive(t *testing.T) {
	sensitiveFields := []string{
		"PASSWORD", "Secret", "API_KEY", // Test case insensitivity
	}

	for _, field := range sensitiveFields {
		if !isSensitiveField(field) {
			t.Errorf("Expected field '%s' to be detected as sensitive", field)
		}
	}
}

// TestSecureCompare tests the secure comparison function
func TestSecureCompare(t *testing.T) {
	// Test equal strings
	if !SecureCompare("same", "same") {
		t.Error("Expected equal strings to compare as equal")
	}
	
	// Test different strings
	if SecureCompare("different1", "different2") {
		t.Error("Expected different strings to compare as not equal")
	}
	
	// Test empty strings
	if !SecureCompare("", "") {
		t.Error("Expected empty strings to compare as equal")
	}
	
	if SecureCompare("value", "") {
		t.Error("Expected non-empty and empty string to compare as not equal")
	}
}

// TestEncodeDecodeSensitiveData tests encoding and decoding of sensitive data
func TestEncodeDecodeSensitiveData(t *testing.T) {
	originalData := []byte("sensitive data")
	
	encoded := EncodeSensitiveData(originalData)
	if encoded == "" {
		t.Error("Expected encoded data, got empty string")
	}
	
	decoded, err := DecodeSensitiveData(encoded)
	if err != nil {
		t.Errorf("Unexpected error decoding data: %v", err)
	}
	
	if string(decoded) != string(originalData) {
		t.Errorf("Expected decoded data '%s', got '%s'", string(originalData), string(decoded))
	}
	
	// Test with invalid encoded data
	_, err = DecodeSensitiveData("invalid_base64!")
	if err == nil {
		t.Error("Expected error for invalid base64 data")
	}
}