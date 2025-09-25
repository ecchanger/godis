package service

import (
	"fmt"
	"strings"

	"todo-backend/models"
	"todo-backend/repository"
)

// TodoService handles business logic for todo operations
type TodoService struct {
	repo *repository.TodoRepository
}

// NewTodoService creates a new TodoService
func NewTodoService(repo *repository.TodoRepository) *TodoService {
	return &TodoService{
		repo: repo,
	}
}

// CreateTodo creates a new todo item with validation
func (s *TodoService) CreateTodo(req *models.CreateTodoRequest) (*models.Todo, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Create todo from request
	todo := &models.Todo{
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Priority:    req.Priority,
		DueDate:     req.DueDate,
		Completed:   false, // New todos are always incomplete
	}

	// Set default priority if not specified
	if todo.Priority == "" {
		todo.Priority = models.PriorityMedium
	}

	// Create todo in repository
	err := s.repo.Create(todo)
	if err != nil {
		return nil, fmt.Errorf("failed to create todo: %w", err)
	}

	return todo, nil
}

// GetTodoByID retrieves a todo by its ID
func (s *TodoService) GetTodoByID(id string) (*models.Todo, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("todo ID cannot be empty")
	}

	todo, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	return todo, nil
}

// GetAllTodos retrieves all todos with optional filtering
func (s *TodoService) GetAllTodos(filter *models.TodoFilter) ([]*models.Todo, error) {
	// Validate filter if provided
	if filter != nil {
		if err := s.validateFilter(filter); err != nil {
			return nil, err
		}
	}

	todos, err := s.repo.GetAll(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get todos: %w", err)
	}

	return todos, nil
}

// UpdateTodo updates an existing todo
func (s *TodoService) UpdateTodo(id string, req *models.UpdateTodoRequest) (*models.Todo, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("todo ID cannot be empty")
	}

	// Validate update request
	if err := s.validateUpdateRequest(req); err != nil {
		return nil, err
	}

	// Update todo in repository
	todo, err := s.repo.Update(id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}

	return todo, nil
}

// DeleteTodo deletes a todo by its ID
func (s *TodoService) DeleteTodo(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("todo ID cannot be empty")
	}

	err := s.repo.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	return nil
}

// ToggleTodoComplete toggles the completion status of a todo
func (s *TodoService) ToggleTodoComplete(id string) (*models.Todo, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("todo ID cannot be empty")
	}

	todo, err := s.repo.ToggleComplete(id)
	if err != nil {
		return nil, fmt.Errorf("failed to toggle todo completion: %w", err)
	}

	return todo, nil
}

// GetTodoStats returns statistics about todos
func (s *TodoService) GetTodoStats() (*models.TodoStatsResponse, error) {
	stats, err := s.repo.GetStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get todo stats: %w", err)
	}

	return stats, nil
}

// validateCreateRequest validates a create todo request
func (s *TodoService) validateCreateRequest(req *models.CreateTodoRequest) error {
	if req == nil {
		return fmt.Errorf("create request cannot be nil")
	}

	if strings.TrimSpace(req.Title) == "" {
		return fmt.Errorf("title cannot be empty")
	}

	if len(req.Title) > 200 {
		return fmt.Errorf("title cannot exceed 200 characters")
	}

	if len(req.Description) > 1000 {
		return fmt.Errorf("description cannot exceed 1000 characters")
	}

	if req.Priority != "" && !models.IsValidPriority(req.Priority) {
		return fmt.Errorf("invalid priority: %s", req.Priority)
	}

	return nil
}

// validateUpdateRequest validates an update todo request
func (s *TodoService) validateUpdateRequest(req *models.UpdateTodoRequest) error {
	if req == nil {
		return fmt.Errorf("update request cannot be nil")
	}

	// Check if at least one field is being updated
	if req.Title == nil && req.Description == nil && req.Completed == nil && 
	   req.Priority == nil && req.DueDate == nil {
		return fmt.Errorf("at least one field must be provided for update")
	}

	if req.Title != nil {
		if strings.TrimSpace(*req.Title) == "" {
			return fmt.Errorf("title cannot be empty")
		}
		if len(*req.Title) > 200 {
			return fmt.Errorf("title cannot exceed 200 characters")
		}
	}

	if req.Description != nil && len(*req.Description) > 1000 {
		return fmt.Errorf("description cannot exceed 1000 characters")
	}

	if req.Priority != nil && !models.IsValidPriority(*req.Priority) {
		return fmt.Errorf("invalid priority: %s", *req.Priority)
	}

	return nil
}

// validateFilter validates a todo filter
func (s *TodoService) validateFilter(filter *models.TodoFilter) error {
	if filter.Priority != nil && !models.IsValidPriority(*filter.Priority) {
		return fmt.Errorf("invalid priority filter: %s", *filter.Priority)
	}

	if filter.SortBy != "" {
		validSortFields := map[string]bool{
			"createdAt": true,
			"title":     true,
			"priority":  true,
			"dueDate":   true,
		}
		if !validSortFields[filter.SortBy] {
			return fmt.Errorf("invalid sort field: %s", filter.SortBy)
		}
	}

	if filter.Order != "" && filter.Order != "asc" && filter.Order != "desc" {
		return fmt.Errorf("invalid sort order: %s (must be 'asc' or 'desc')", filter.Order)
	}

	return nil
}