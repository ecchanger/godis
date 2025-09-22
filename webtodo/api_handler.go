package webtodo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hdt3213/godis/lib/logger"
	"github.com/hdt3213/godis/redis/client"
	"github.com/hdt3213/godis/redis/protocol"
)

// APIHandler handles REST API requests for todo operations
type APIHandler struct {
	redisClient *client.Client
	todoService *TodoService
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(redisClient *client.Client) *APIHandler {
	return &APIHandler{
		redisClient: redisClient,
		todoService: NewTodoService(redisClient),
	}
}

// ServeHTTP implements the http.Handler interface
func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Route the request
	path := strings.TrimPrefix(r.URL.Path, "/")
	pathParts := strings.Split(path, "/")

	if len(pathParts) == 0 || pathParts[0] != "todos" {
		h.sendErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "Endpoint not found", nil)
		return
	}

	// Handle different HTTP methods and paths
	switch r.Method {
	case "GET":
		if len(pathParts) == 1 {
			// GET /api/todos
			h.handleGetTodos(w, r)
		} else if len(pathParts) == 2 {
			// GET /api/todos/{id}
			h.handleGetTodo(w, r, pathParts[1])
		} else {
			h.sendErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "Invalid path", nil)
		}
	case "POST":
		if len(pathParts) == 1 {
			// POST /api/todos
			h.handleCreateTodo(w, r)
		} else {
			h.sendErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "Invalid path", nil)
		}
	case "PUT":
		if len(pathParts) == 2 {
			// PUT /api/todos/{id}
			h.handleUpdateTodo(w, r, pathParts[1])
		} else {
			h.sendErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "Invalid path", nil)
		}
	case "DELETE":
		if len(pathParts) == 2 {
			// DELETE /api/todos/{id}
			h.handleDeleteTodo(w, r, pathParts[1])
		} else {
			h.sendErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "Invalid path", nil)
		}
	default:
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", nil)
	}
}

// handleGetTodos retrieves all todos
func (h *APIHandler) handleGetTodos(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	
	// Parse query parameters
	completed := r.URL.Query().Get("completed")
	priority := r.URL.Query().Get("priority")
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	filter := TodoFilter{
		UserID:   userID,
		Priority: priority,
	}

	if completed != "" {
		if completedBool, err := strconv.ParseBool(completed); err == nil {
			filter.Completed = &completedBool
		}
	}

	if limit != "" {
		if limitInt, err := strconv.Atoi(limit); err == nil && limitInt > 0 {
			filter.Limit = limitInt
		}
	}

	if offset != "" {
		if offsetInt, err := strconv.Atoi(offset); err == nil && offsetInt >= 0 {
			filter.Offset = offsetInt
		}
	}

	todos, err := h.todoService.GetTodos(filter)
	if err != nil {
		logger.Errorf("Error getting todos: %v", err)
		h.sendErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve todos", nil)
		return
	}

	h.sendSuccessResponse(w, http.StatusOK, todos, "Todos retrieved successfully")
}

// handleGetTodo retrieves a specific todo
func (h *APIHandler) handleGetTodo(w http.ResponseWriter, r *http.Request, todoID string) {
	userID := h.getUserID(r)

	todo, err := h.todoService.GetTodo(userID, todoID)
	if err != nil {
		if err == ErrTodoNotFound {
			h.sendErrorResponse(w, http.StatusNotFound, "TODO_NOT_FOUND", "Todo not found", nil)
		} else {
			logger.Errorf("Error getting todo %s: %v", todoID, err)
			h.sendErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve todo", nil)
		}
		return
	}

	h.sendSuccessResponse(w, http.StatusOK, todo, "Todo retrieved successfully")
}

// handleCreateTodo creates a new todo
func (h *APIHandler) handleCreateTodo(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)

	var request CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format", nil)
		return
	}

	// Validate request
	if err := h.validateCreateTodoRequest(&request); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	todo, err := h.todoService.CreateTodo(userID, &request)
	if err != nil {
		logger.Errorf("Error creating todo: %v", err)
		h.sendErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create todo", nil)
		return
	}

	h.sendSuccessResponse(w, http.StatusCreated, todo, "Todo created successfully")
}

// handleUpdateTodo updates an existing todo
func (h *APIHandler) handleUpdateTodo(w http.ResponseWriter, r *http.Request, todoID string) {
	userID := h.getUserID(r)

	var request UpdateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format", nil)
		return
	}

	todo, err := h.todoService.UpdateTodo(userID, todoID, &request)
	if err != nil {
		if err == ErrTodoNotFound {
			h.sendErrorResponse(w, http.StatusNotFound, "TODO_NOT_FOUND", "Todo not found", nil)
		} else {
			logger.Errorf("Error updating todo %s: %v", todoID, err)
			h.sendErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update todo", nil)
		}
		return
	}

	h.sendSuccessResponse(w, http.StatusOK, todo, "Todo updated successfully")
}

// handleDeleteTodo deletes a todo
func (h *APIHandler) handleDeleteTodo(w http.ResponseWriter, r *http.Request, todoID string) {
	userID := h.getUserID(r)

	err := h.todoService.DeleteTodo(userID, todoID)
	if err != nil {
		if err == ErrTodoNotFound {
			h.sendErrorResponse(w, http.StatusNotFound, "TODO_NOT_FOUND", "Todo not found", nil)
		} else {
			logger.Errorf("Error deleting todo %s: %v", todoID, err)
			h.sendErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete todo", nil)
		}
		return
	}

	h.sendSuccessResponse(w, http.StatusOK, nil, "Todo deleted successfully")
}

// getUserID extracts user ID from request (for now, use a default user)
func (h *APIHandler) getUserID(r *http.Request) string {
	// In a real implementation, this would extract user ID from JWT token or session
	// For now, use a default user ID
	return "default_user"
}

// validateCreateTodoRequest validates the create todo request
func (h *APIHandler) validateCreateTodoRequest(request *CreateTodoRequest) error {
	if strings.TrimSpace(request.Title) == "" {
		return fmt.Errorf("title is required")
	}

	if len(request.Title) > 255 {
		return fmt.Errorf("title must be less than 255 characters")
	}

	if len(request.Description) > 1000 {
		return fmt.Errorf("description must be less than 1000 characters")
	}

	if request.Priority != "" {
		validPriorities := map[string]bool{"low": true, "medium": true, "high": true}
		if !validPriorities[strings.ToLower(request.Priority)] {
			return fmt.Errorf("priority must be one of: low, medium, high")
		}
	}

	if request.DueDate != "" {
		if _, err := time.Parse(time.RFC3339, request.DueDate); err != nil {
			return fmt.Errorf("due_date must be in RFC3339 format")
		}
	}

	return nil
}

// sendSuccessResponse sends a successful API response
func (h *APIHandler) sendSuccessResponse(w http.ResponseWriter, statusCode int, data interface{}, message string) {
	response := APIResponse{
		Success: true,
		Data:    data,
		Message: message,
	}

	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

// sendErrorResponse sends an error API response
func (h *APIHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string, details interface{}) {
	response := APIResponse{
		Success: false,
		Error:   errorCode,
		Message: message,
		Data:    details,
	}

	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf("Error encoding error response: %v", err)
	}
}

// isRedisError checks if the error is a Redis-related error
func (h *APIHandler) isRedisError(reply interface{}) bool {
	if errReply, ok := reply.(*protocol.StandardErrReply); ok {
		logger.Errorf("Redis error: %s", errReply.Error())
		return true
	}
	return false
}