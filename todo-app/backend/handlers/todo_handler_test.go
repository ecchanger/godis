package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"todo-backend/models"

	"github.com/gin-gonic/gin"
)

// MockTodoService implements the service interface for testing
type MockTodoService struct {
	todos  map[string]*models.Todo
	nextID int
}

func NewMockTodoService() *MockTodoService {
	return &MockTodoService{
		todos:  make(map[string]*models.Todo),
		nextID: 1,
	}
}

func (m *MockTodoService) CreateTodo(req *models.CreateTodoRequest) (*models.Todo, error) {
	todo := &models.Todo{
		ID:          fmt.Sprintf("test-id-%d", m.nextID),
		Title:       req.Title,
		Description: req.Description,
		Completed:   false,
		Priority:    req.Priority,
		DueDate:     req.DueDate,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if todo.Priority == "" {
		todo.Priority = models.PriorityMedium
	}
	m.nextID++
	m.todos[todo.ID] = todo
	return todo, nil
}

func (m *MockTodoService) GetTodoByID(id string) (*models.Todo, error) {
	todo, exists := m.todos[id]
	if !exists {
		return nil, fmt.Errorf("todo not found")
	}
	return todo, nil
}

func (m *MockTodoService) GetAllTodos(filter *models.TodoFilter) ([]*models.Todo, error) {
	var result []*models.Todo
	for _, todo := range m.todos {
		result = append(result, todo)
	}
	return result, nil
}

func (m *MockTodoService) UpdateTodo(id string, req *models.UpdateTodoRequest) (*models.Todo, error) {
	todo, exists := m.todos[id]
	if !exists {
		return nil, fmt.Errorf("todo not found")
	}
	
	if req.Title != nil {
		todo.Title = *req.Title
	}
	if req.Completed != nil {
		todo.Completed = *req.Completed
	}
	todo.UpdatedAt = time.Now()
	
	return todo, nil
}

func (m *MockTodoService) DeleteTodo(id string) error {
	_, exists := m.todos[id]
	if !exists {
		return fmt.Errorf("todo not found")
	}
	delete(m.todos, id)
	return nil
}

func (m *MockTodoService) ToggleTodoComplete(id string) (*models.Todo, error) {
	todo, exists := m.todos[id]
	if !exists {
		return nil, fmt.Errorf("todo not found")
	}
	todo.Completed = !todo.Completed
	todo.UpdatedAt = time.Now()
	return todo, nil
}

func (m *MockTodoService) GetTodoStats() (*models.TodoStatsResponse, error) {
	total := len(m.todos)
	completed := 0
	for _, todo := range m.todos {
		if todo.Completed {
			completed++
		}
	}
	
	return &models.TodoStatsResponse{
		Total:        total,
		Completed:    completed,
		Pending:      total - completed,
		HighPriority: 0,
	}, nil
}

func setupTestRouter() (*gin.Engine, *MockTodoService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	mockService := NewMockTodoService()
	handler := &TodoHandler{service: mockService}
	
	api := router.Group("/api")
	{
		todos := api.Group("/todos")
		{
			todos.GET("", handler.GetAllTodos)
			todos.POST("", handler.CreateTodo)
			todos.GET("/:id", handler.GetTodoByID)
			todos.PUT("/:id", handler.UpdateTodo)
			todos.DELETE("/:id", handler.DeleteTodo)
			todos.PATCH("/:id/toggle", handler.ToggleTodoComplete)
		}
		api.GET("/stats", handler.GetTodoStats)
	}
	
	return router, mockService
}

func TestCreateTodo(t *testing.T) {
	router, _ := setupTestRouter()

	todo := models.CreateTodoRequest{
		Title:       "Test Todo",
		Description: "Test Description",
		Priority:    models.PriorityHigh,
	}

	jsonData, _ := json.Marshal(todo)
	req, _ := http.NewRequest("POST", "/api/todos", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success true, got %v", response.Success)
	}
}

func TestGetAllTodos(t *testing.T) {
	router, _ := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/todos", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success true, got %v", response.Success)
	}
}

func TestGetTodoByID(t *testing.T) {
	router, mockService := setupTestRouter()

	// First create a todo
	mockService := mockService
	todo, _ := mockService.CreateTodo(&models.CreateTodoRequest{
		Title: "Test Todo",
	})

	req, _ := http.NewRequest("GET", "/api/todos/"+todo.ID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success true, got %v", response.Success)
	}
}

func TestGetTodoByIDNotFound(t *testing.T) {
	router, _ := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/todos/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Success {
		t.Errorf("Expected success false, got %v", response.Success)
	}
}

func TestUpdateTodo(t *testing.T) {
	router, handler := setupTestRouter()

	// First create a todo
	mockService := mockService
	todo, _ := mockService.CreateTodo(&models.CreateTodoRequest{
		Title: "Original Title",
	})

	// Update the todo
	newTitle := "Updated Title"
	updateReq := models.UpdateTodoRequest{
		Title: &newTitle,
	}

	jsonData, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/api/todos/"+todo.ID, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success true, got %v", response.Success)
	}
}

func TestDeleteTodo(t *testing.T) {
	router, handler := setupTestRouter()

	// First create a todo
	mockService := mockService
	todo, _ := mockService.CreateTodo(&models.CreateTodoRequest{
		Title: "Test Todo",
	})

	req, _ := http.NewRequest("DELETE", "/api/todos/"+todo.ID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success true, got %v", response.Success)
	}
}

func TestToggleTodoComplete(t *testing.T) {
	router, handler := setupTestRouter()

	// First create a todo
	mockService := mockService
	todo, _ := mockService.CreateTodo(&models.CreateTodoRequest{
		Title: "Test Todo",
	})

	req, _ := http.NewRequest("PATCH", "/api/todos/"+todo.ID+"/toggle", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success true, got %v", response.Success)
	}
}

func TestGetTodoStats(t *testing.T) {
	router, _ := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success true, got %v", response.Success)
	}
}

func TestCreateTodoInvalidJSON(t *testing.T) {
	router, _ := setupTestRouter()

	req, _ := http.NewRequest("POST", "/api/todos", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Success {
		t.Errorf("Expected success false, got %v", response.Success)
	}
}