package service

import (
	"fmt"
	"testing"
	"time"

	"todo-backend/models"
)

// MockTodoRepository implements the repository interface for testing
type MockTodoRepository struct {
	todos  map[string]*models.Todo
	nextID int
}

func NewMockTodoRepository() *MockTodoRepository {
	return &MockTodoRepository{
		todos:  make(map[string]*models.Todo),
		nextID: 1,
	}
}

func (m *MockTodoRepository) Create(todo *models.Todo) error {
	todo.ID = fmt.Sprintf("mock-id-%d", m.nextID)
	m.nextID++
	now := time.Now()
	todo.CreatedAt = now
	todo.UpdatedAt = now
	m.todos[todo.ID] = todo
	return nil
}

func (m *MockTodoRepository) GetByID(id string) (*models.Todo, error) {
	todo, exists := m.todos[id]
	if !exists {
		return nil, fmt.Errorf("todo not found")
	}
	return todo, nil
}

func (m *MockTodoRepository) GetAll(filter *models.TodoFilter) ([]*models.Todo, error) {
	var result []*models.Todo
	for _, todo := range m.todos {
		// Apply filter if provided
		if filter != nil && filter.Completed != nil && todo.Completed != *filter.Completed {
			continue
		}
		if filter != nil && filter.Priority != nil && todo.Priority != *filter.Priority {
			continue
		}
		result = append(result, todo)
	}
	return result, nil
}

func (m *MockTodoRepository) Update(id string, updates *models.UpdateTodoRequest) (*models.Todo, error) {
	todo, exists := m.todos[id]
	if !exists {
		return nil, fmt.Errorf("todo not found")
	}
	
	// Apply updates
	if updates.Title != nil {
		todo.Title = *updates.Title
	}
	if updates.Description != nil {
		todo.Description = *updates.Description
	}
	if updates.Completed != nil {
		todo.Completed = *updates.Completed
	}
	if updates.Priority != nil {
		todo.Priority = *updates.Priority
	}
	todo.UpdatedAt = time.Now()
	
	return todo, nil
}

func (m *MockTodoRepository) Delete(id string) error {
	_, exists := m.todos[id]
	if !exists {
		return fmt.Errorf("todo not found")
	}
	delete(m.todos, id)
	return nil
}

func (m *MockTodoRepository) ToggleComplete(id string) (*models.Todo, error) {
	todo, exists := m.todos[id]
	if !exists {
		return nil, fmt.Errorf("todo not found")
	}
	todo.Completed = !todo.Completed
	todo.UpdatedAt = time.Now()
	return todo, nil
}

func (m *MockTodoRepository) GetStats() (*models.TodoStatsResponse, error) {
	total := len(m.todos)
	completed := 0
	highPriority := 0
	
	for _, todo := range m.todos {
		if todo.Completed {
			completed++
		}
		if todo.Priority == models.PriorityHigh {
			highPriority++
		}
	}
	
	return &models.TodoStatsResponse{
		Total:        total,
		Completed:    completed,
		Pending:      total - completed,
		HighPriority: highPriority,
	}, nil
}

func TestCreateTodo(t *testing.T) {
	repo := NewMockTodoRepository()
	service := NewTodoService(repo)

	req := &models.CreateTodoRequest{
		Title:       "Test Todo",
		Description: "Test Description",
		Priority:    models.PriorityHigh,
	}

	todo, err := service.CreateTodo(req)
	if err != nil {
		t.Fatalf("CreateTodo failed: %v", err)
	}

	if todo.Title != req.Title {
		t.Errorf("Expected title '%s', got '%s'", req.Title, todo.Title)
	}
	if todo.Priority != req.Priority {
		t.Errorf("Expected priority '%s', got '%s'", req.Priority, todo.Priority)
	}
	if todo.Completed {
		t.Errorf("Expected completed false, got %v", todo.Completed)
	}
}

func TestCreateTodoValidation(t *testing.T) {
	repo := NewMockTodoRepository()
	service := NewTodoService(repo)

	// Test empty title
	req := &models.CreateTodoRequest{
		Title: "",
	}

	_, err := service.CreateTodo(req)
	if err == nil {
		t.Errorf("Expected validation error for empty title")
	}

	// Test title too long
	req = &models.CreateTodoRequest{
		Title: string(make([]byte, 201)), // 201 characters
	}

	_, err = service.CreateTodo(req)
	if err == nil {
		t.Errorf("Expected validation error for title too long")
	}
}

func TestGetTodoByID(t *testing.T) {
	repo := NewMockTodoRepository()
	service := NewTodoService(repo)

	// Create a todo first
	createReq := &models.CreateTodoRequest{
		Title: "Test Todo",
	}
	created, err := service.CreateTodo(createReq)
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Get the todo
	retrieved, err := service.GetTodoByID(created.ID)
	if err != nil {
		t.Fatalf("GetTodoByID failed: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID '%s', got '%s'", created.ID, retrieved.ID)
	}
	if retrieved.Title != created.Title {
		t.Errorf("Expected title '%s', got '%s'", created.Title, retrieved.Title)
	}
}

func TestUpdateTodo(t *testing.T) {
	repo := NewMockTodoRepository()
	service := NewTodoService(repo)

	// Create a todo first
	createReq := &models.CreateTodoRequest{
		Title: "Original Title",
	}
	created, err := service.CreateTodo(createReq)
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Update the todo
	newTitle := "Updated Title"
	completed := true
	updateReq := &models.UpdateTodoRequest{
		Title:     &newTitle,
		Completed: &completed,
	}

	updated, err := service.UpdateTodo(created.ID, updateReq)
	if err != nil {
		t.Fatalf("UpdateTodo failed: %v", err)
	}

	if updated.Title != newTitle {
		t.Errorf("Expected title '%s', got '%s'", newTitle, updated.Title)
	}
	if !updated.Completed {
		t.Errorf("Expected completed true, got %v", updated.Completed)
	}
}

func TestToggleTodoComplete(t *testing.T) {
	repo := NewMockTodoRepository()
	service := NewTodoService(repo)

	// Create a todo first
	createReq := &models.CreateTodoRequest{
		Title: "Test Todo",
	}
	created, err := service.CreateTodo(createReq)
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	originalCompleted := created.Completed

	// Toggle completion
	toggled, err := service.ToggleTodoComplete(created.ID)
	if err != nil {
		t.Fatalf("ToggleTodoComplete failed: %v", err)
	}

	if toggled.Completed == originalCompleted {
		t.Errorf("Expected completion status to change from %v", originalCompleted)
	}
}

func TestDeleteTodo(t *testing.T) {
	repo := NewMockTodoRepository()
	service := NewTodoService(repo)

	// Create a todo first
	createReq := &models.CreateTodoRequest{
		Title: "Test Todo",
	}
	created, err := service.CreateTodo(createReq)
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Delete the todo
	err = service.DeleteTodo(created.ID)
	if err != nil {
		t.Fatalf("DeleteTodo failed: %v", err)
	}

	// Try to get the deleted todo
	_, err = service.GetTodoByID(created.ID)
	if err == nil {
		t.Errorf("Expected error when getting deleted todo")
	}
}

func TestGetTodoStats(t *testing.T) {
	repo := NewMockTodoRepository()
	service := NewTodoService(repo)

	// Create some test todos
	todos := []*models.CreateTodoRequest{
		{Title: "Todo 1", Priority: models.PriorityHigh},
		{Title: "Todo 2", Priority: models.PriorityMedium},
		{Title: "Todo 3", Priority: models.PriorityHigh},
	}

	var createdIDs []string
	for _, req := range todos {
		created, err := service.CreateTodo(req)
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}
		createdIDs = append(createdIDs, created.ID)
	}

	// Complete one todo
	service.ToggleTodoComplete(createdIDs[0])

	// Get stats
	stats, err := service.GetTodoStats()
	if err != nil {
		t.Fatalf("GetTodoStats failed: %v", err)
	}

	if stats.Total != 3 {
		t.Errorf("Expected total 3, got %d", stats.Total)
	}
	if stats.Completed != 1 {
		t.Errorf("Expected completed 1, got %d", stats.Completed)
	}
	if stats.Pending != 2 {
		t.Errorf("Expected pending 2, got %d", stats.Pending)
	}
	if stats.HighPriority != 2 {
		t.Errorf("Expected high priority 2, got %d", stats.HighPriority)
	}
}