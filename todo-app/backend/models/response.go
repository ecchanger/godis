package models

// APIResponse represents a standard API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    string      `json:"code,omitempty"`
}

// SuccessResponse creates a successful API response
func SuccessResponse(data interface{}, message string) *APIResponse {
	return &APIResponse{
		Success: true,
		Data:    data,
		Message: message,
	}
}

// ErrorResponse creates an error API response
func ErrorResponse(error, code string) *APIResponse {
	return &APIResponse{
		Success: false,
		Error:   error,
		Code:    code,
	}
}

// TodoListResponse represents a response containing multiple todos
type TodoListResponse struct {
	Todos []Todo `json:"todos"`
	Total int    `json:"total"`
}

// TodoStatsResponse represents todo statistics
type TodoStatsResponse struct {
	Total       int `json:"total"`
	Completed   int `json:"completed"`
	Pending     int `json:"pending"`
	HighPriority int `json:"highPriority"`
}