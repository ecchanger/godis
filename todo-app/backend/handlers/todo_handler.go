package handlers

import (
	"net/http"
	"strconv"

	"todo-backend/models"
	"todo-backend/service"

	"github.com/gin-gonic/gin"
)

// TodoHandler handles HTTP requests for todo operations
type TodoHandler struct {
	service *service.TodoService
}

// NewTodoHandler creates a new TodoHandler
func NewTodoHandler(service *service.TodoService) *TodoHandler {
	return &TodoHandler{
		service: service,
	}
}

// CreateTodo handles POST /api/todos
func (h *TodoHandler) CreateTodo(c *gin.Context) {
	var req models.CreateTodoRequest
	
	// Bind JSON request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid request body: "+err.Error(),
			"INVALID_REQUEST",
		))
		return
	}
	
	// Create todo via service
	todo, err := h.service.CreateTodo(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			err.Error(),
			"CREATE_FAILED",
		))
		return
	}
	
	// Return success response
	c.JSON(http.StatusCreated, models.SuccessResponse(
		todo,
		"Todo created successfully",
	))
}

// GetTodoByID handles GET /api/todos/:id
func (h *TodoHandler) GetTodoByID(c *gin.Context) {
	id := c.Param("id")
	
	// Get todo via service
	todo, err := h.service.GetTodoByID(id)
	if err != nil {
		status := http.StatusNotFound
		if err.Error() == "todo ID cannot be empty" {
			status = http.StatusBadRequest
		}
		
		c.JSON(status, models.ErrorResponse(
			err.Error(),
			"GET_FAILED",
		))
		return
	}
	
	// Return success response
	c.JSON(http.StatusOK, models.SuccessResponse(
		todo,
		"Todo retrieved successfully",
	))
}

// GetAllTodos handles GET /api/todos
func (h *TodoHandler) GetAllTodos(c *gin.Context) {
	// Parse query parameters for filtering
	var filter models.TodoFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid query parameters: "+err.Error(),
			"INVALID_QUERY",
		))
		return
	}
	
	// Get todos via service
	todos, err := h.service.GetAllTodos(&filter)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			err.Error(),
			"GET_ALL_FAILED",
		))
		return
	}
	
	// Create response with todos and total count
	response := models.TodoListResponse{
		Todos: make([]models.Todo, len(todos)),
		Total: len(todos),
	}
	
	// Convert pointers to values for response
	for i, todo := range todos {
		response.Todos[i] = *todo
	}
	
	// Return success response
	c.JSON(http.StatusOK, models.SuccessResponse(
		response,
		"Todos retrieved successfully",
	))
}

// UpdateTodo handles PUT /api/todos/:id
func (h *TodoHandler) UpdateTodo(c *gin.Context) {
	id := c.Param("id")
	
	var req models.UpdateTodoRequest
	
	// Bind JSON request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid request body: "+err.Error(),
			"INVALID_REQUEST",
		))
		return
	}
	
	// Update todo via service
	todo, err := h.service.UpdateTodo(id, &req)
	if err != nil {
		status := http.StatusNotFound
		if err.Error() == "todo ID cannot be empty" || 
		   err.Error() == "at least one field must be provided for update" {
			status = http.StatusBadRequest
		}
		
		c.JSON(status, models.ErrorResponse(
			err.Error(),
			"UPDATE_FAILED",
		))
		return
	}
	
	// Return success response
	c.JSON(http.StatusOK, models.SuccessResponse(
		todo,
		"Todo updated successfully",
	))
}

// DeleteTodo handles DELETE /api/todos/:id
func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	id := c.Param("id")
	
	// Delete todo via service
	err := h.service.DeleteTodo(id)
	if err != nil {
		status := http.StatusNotFound
		if err.Error() == "todo ID cannot be empty" {
			status = http.StatusBadRequest
		}
		
		c.JSON(status, models.ErrorResponse(
			err.Error(),
			"DELETE_FAILED",
		))
		return
	}
	
	// Return success response
	c.JSON(http.StatusOK, models.SuccessResponse(
		gin.H{"deleted": true},
		"Todo deleted successfully",
	))
}

// ToggleTodoComplete handles PATCH /api/todos/:id/toggle
func (h *TodoHandler) ToggleTodoComplete(c *gin.Context) {
	id := c.Param("id")
	
	// Toggle todo completion via service
	todo, err := h.service.ToggleTodoComplete(id)
	if err != nil {
		status := http.StatusNotFound
		if err.Error() == "todo ID cannot be empty" {
			status = http.StatusBadRequest
		}
		
		c.JSON(status, models.ErrorResponse(
			err.Error(),
			"TOGGLE_FAILED",
		))
		return
	}
	
	// Return success response
	c.JSON(http.StatusOK, models.SuccessResponse(
		todo,
		"Todo completion status toggled successfully",
	))
}

// GetTodoStats handles GET /api/stats
func (h *TodoHandler) GetTodoStats(c *gin.Context) {
	// Get stats via service
	stats, err := h.service.GetTodoStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			err.Error(),
			"STATS_FAILED",
		))
		return
	}
	
	// Return success response
	c.JSON(http.StatusOK, models.SuccessResponse(
		stats,
		"Stats retrieved successfully",
	))
}

// Helper function to parse boolean query parameter
func parseBoolParam(c *gin.Context, param string) *bool {
	value := c.Query(param)
	if value == "" {
		return nil
	}
	
	if boolValue, err := strconv.ParseBool(value); err == nil {
		return &boolValue
	}
	
	return nil
}