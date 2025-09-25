package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTodoToJSON(t *testing.T) {
	now := time.Now()
	todo := &Todo{
		ID:          "test-id",
		Title:       "Test Todo",
		Description: "Test Description",
		Completed:   false,
		Priority:    PriorityHigh,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	jsonData, err := todo.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Verify JSON contains expected fields
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result["id"] != "test-id" {
		t.Errorf("Expected id 'test-id', got %v", result["id"])
	}
	if result["title"] != "Test Todo" {
		t.Errorf("Expected title 'Test Todo', got %v", result["title"])
	}
	if result["priority"] != string(PriorityHigh) {
		t.Errorf("Expected priority 'high', got %v", result["priority"])
	}
}

func TestFromJSON(t *testing.T) {
	jsonData := `{
		"id": "test-id",
		"title": "Test Todo",
		"description": "Test Description",
		"completed": false,
		"priority": "high",
		"createdAt": "2023-01-01T00:00:00Z",
		"updatedAt": "2023-01-01T00:00:00Z"
	}`

	todo, err := FromJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if todo.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got %s", todo.ID)
	}
	if todo.Title != "Test Todo" {
		t.Errorf("Expected title 'Test Todo', got %s", todo.Title)
	}
	if todo.Priority != PriorityHigh {
		t.Errorf("Expected priority 'high', got %s", todo.Priority)
	}
	if todo.Completed {
		t.Errorf("Expected completed false, got %v", todo.Completed)
	}
}

func TestIsValidPriority(t *testing.T) {
	tests := []struct {
		priority Priority
		expected bool
	}{
		{PriorityLow, true},
		{PriorityMedium, true},
		{PriorityHigh, true},
		{Priority("invalid"), false},
		{Priority(""), false},
	}

	for _, tt := range tests {
		result := IsValidPriority(tt.priority)
		if result != tt.expected {
			t.Errorf("IsValidPriority(%s) = %v, expected %v", tt.priority, result, tt.expected)
		}
	}
}

func TestSuccessResponse(t *testing.T) {
	data := map[string]string{"key": "value"}
	message := "Success"

	response := SuccessResponse(data, message)

	if !response.Success {
		t.Errorf("Expected Success true, got %v", response.Success)
	}
	if response.Message != message {
		t.Errorf("Expected message '%s', got %s", message, response.Message)
	}
	if response.Data == nil {
		t.Errorf("Expected data to be set, got nil")
	}
	if response.Error != "" {
		t.Errorf("Expected empty error, got %s", response.Error)
	}
}

func TestErrorResponse(t *testing.T) {
	errorMsg := "Something went wrong"
	code := "ERROR_CODE"

	response := ErrorResponse(errorMsg, code)

	if response.Success {
		t.Errorf("Expected Success false, got %v", response.Success)
	}
	if response.Error != errorMsg {
		t.Errorf("Expected error '%s', got %s", errorMsg, response.Error)
	}
	if response.Code != code {
		t.Errorf("Expected code '%s', got %s", code, response.Code)
	}
	if response.Data != nil {
		t.Errorf("Expected data to be nil, got %v", response.Data)
	}
}