package repository

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"todo-backend/models"

	"github.com/google/uuid"
)

// GodisClient interface defines the operations we need from Godis
type GodisClient interface {
	Set(key string, value interface{}) error
	Get(key string) (interface{}, error)
	Del(keys ...string) error
	Exists(key string) (bool, error)
	SAdd(key string, members ...interface{}) error
	SRem(key string, members ...interface{}) error
	SMembers(key string) ([]string, error)
	HSet(key string, field string, value interface{}) error
	HGet(key string, field string) (interface{}, error)
	HGetAll(key string) (map[string]interface{}, error)
	HDel(key string, fields ...string) error
}

// TodoRepository handles Todo data operations using Godis
type TodoRepository struct {
	client GodisClient
}

// NewTodoRepository creates a new TodoRepository
func NewTodoRepository(client GodisClient) *TodoRepository {
	return &TodoRepository{
		client: client,
	}
}

// Constants for Redis keys
const (
	TodoHashPrefix     = "todo:"
	TodosAllSet        = "todos:all"
	TodosCompletedSet  = "todos:completed"
	TodosPendingSet    = "todos:pending"
	TodosPriorityPrefix = "todos:priority:"
	TodosTagPrefix     = "todos:tag:"
	TodoCounterKey     = "todo:counter"
)

// Create creates a new todo item
func (r *TodoRepository) Create(todo *models.Todo) error {
	// Generate ID and timestamps
	todo.ID = uuid.New().String()
	now := time.Now()
	todo.CreatedAt = now
	todo.UpdatedAt = now

	// Set default priority if not specified
	if todo.Priority == "" {
		todo.Priority = models.PriorityMedium
	}

	// Convert todo to JSON for storage
	todoJSON, err := todo.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal todo: %w", err)
	}

	// Store todo data as hash
	todoKey := TodoHashPrefix + todo.ID
	err = r.client.Set(todoKey, string(todoJSON))
	if err != nil {
		return fmt.Errorf("failed to store todo: %w", err)
	}

	// Add to all todos set
	err = r.client.SAdd(TodosAllSet, todo.ID)
	if err != nil {
		return fmt.Errorf("failed to add to all todos set: %w", err)
	}

	// Add to appropriate status set
	if todo.Completed {
		err = r.client.SAdd(TodosCompletedSet, todo.ID)
	} else {
		err = r.client.SAdd(TodosPendingSet, todo.ID)
	}
	if err != nil {
		return fmt.Errorf("failed to add to status set: %w", err)
	}

	// Add to priority set
	priorityKey := TodosPriorityPrefix + string(todo.Priority)
	err = r.client.SAdd(priorityKey, todo.ID)
	if err != nil {
		return fmt.Errorf("failed to add to priority set: %w", err)
	}

	// Add to tag sets
	for _, tag := range todo.Tags {
		tagKey := TodosTagPrefix + tag
		err = r.client.SAdd(tagKey, todo.ID)
		if err != nil {
			return fmt.Errorf("failed to add to tag set: %w", err)
		}
	}

	return nil
}

// GetByID retrieves a todo by its ID
func (r *TodoRepository) GetByID(id string) (*models.Todo, error) {
	todoKey := TodoHashPrefix + id

	// Check if todo exists
	exists, err := r.client.Exists(todoKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check todo existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("todo not found")
	}

	// Get todo data
	data, err := r.client.Get(todoKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	// Parse JSON data
	dataStr, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	todo, err := models.FromJSON([]byte(dataStr))
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal todo: %w", err)
	}

	return todo, nil
}

// GetAll retrieves all todos with optional filtering
func (r *TodoRepository) GetAll(filter *models.TodoFilter) ([]*models.Todo, error) {
	// Get todo IDs based on filter
	var candidateIDs []string
	var err error

	// Start with tag filter if specified (most restrictive)
	if filter != nil && filter.Tag != nil && *filter.Tag != "" {
		tagKey := TodosTagPrefix + *filter.Tag
		candidateIDs, err = r.client.SMembers(tagKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get todos by tag: %w", err)
		}
	} else {
		// Otherwise get from completion status filter or all todos
		if filter != nil && filter.Completed != nil {
			if *filter.Completed {
				candidateIDs, err = r.client.SMembers(TodosCompletedSet)
			} else {
				candidateIDs, err = r.client.SMembers(TodosPendingSet)
			}
		} else {
			candidateIDs, err = r.client.SMembers(TodosAllSet)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get todo IDs: %w", err)
	}

	// Retrieve all todos
	todos := make([]*models.Todo, 0, len(candidateIDs))
	for _, id := range candidateIDs {
		todo, err := r.GetByID(id)
		if err != nil {
			// Skip invalid todos but log the error
			continue
		}

		// Apply completion filter if tag filter was used
		if filter != nil && filter.Tag != nil && filter.Completed != nil {
			if todo.Completed != *filter.Completed {
				continue
			}
		}

		// Apply priority filter if specified
		if filter != nil && filter.Priority != nil && todo.Priority != *filter.Priority {
			continue
		}

		todos = append(todos, todo)
	}

	// Sort todos if requested
	if filter != nil && filter.SortBy != "" {
		r.sortTodos(todos, filter.SortBy, filter.Order)
	}

	return todos, nil
}

// Update updates an existing todo
func (r *TodoRepository) Update(id string, updates *models.UpdateTodoRequest) (*models.Todo, error) {
	// Get existing todo
	existingTodo, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Track if completion status changed
	oldCompleted := existingTodo.Completed
	
	// Track old tags for index update
	oldTags := make(map[string]bool)
	for _, tag := range existingTodo.Tags {
		oldTags[tag] = true
	}

	// Apply updates
	if updates.Title != nil {
		existingTodo.Title = *updates.Title
	}
	if updates.Description != nil {
		existingTodo.Description = *updates.Description
	}
	if updates.Completed != nil {
		existingTodo.Completed = *updates.Completed
	}
	if updates.Priority != nil && models.IsValidPriority(*updates.Priority) {
		// Remove from old priority set
		oldPriorityKey := TodosPriorityPrefix + string(existingTodo.Priority)
		r.client.SRem(oldPriorityKey, id)
		
		existingTodo.Priority = *updates.Priority
		
		// Add to new priority set
		newPriorityKey := TodosPriorityPrefix + string(existingTodo.Priority)
		r.client.SAdd(newPriorityKey, id)
	}
	if updates.DueDate != nil {
		existingTodo.DueDate = updates.DueDate
	}
	if updates.Tags != nil {
		// Remove from old tag sets
		for tag := range oldTags {
			tagKey := TodosTagPrefix + tag
			r.client.SRem(tagKey, id)
		}
		
		existingTodo.Tags = *updates.Tags
		
		// Add to new tag sets
		for _, tag := range existingTodo.Tags {
			tagKey := TodosTagPrefix + tag
			r.client.SAdd(tagKey, id)
		}
	}

	// Update timestamp
	existingTodo.UpdatedAt = time.Now()

	// Save updated todo
	todoJSON, err := existingTodo.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated todo: %w", err)
	}

	todoKey := TodoHashPrefix + id
	err = r.client.Set(todoKey, string(todoJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to save updated todo: %w", err)
	}

	// Update status sets if completion status changed
	if updates.Completed != nil && oldCompleted != *updates.Completed {
		if *updates.Completed {
			// Move from pending to completed
			r.client.SRem(TodosPendingSet, id)
			r.client.SAdd(TodosCompletedSet, id)
		} else {
			// Move from completed to pending
			r.client.SRem(TodosCompletedSet, id)
			r.client.SAdd(TodosPendingSet, id)
		}
	}

	return existingTodo, nil
}

// Delete removes a todo
func (r *TodoRepository) Delete(id string) error {
	// Get todo to determine its current state
	todo, err := r.GetByID(id)
	if err != nil {
		return err
	}

	// Remove from all sets
	todoKey := TodoHashPrefix + id
	err = r.client.Del(todoKey)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	// Remove from all sets
	r.client.SRem(TodosAllSet, id)
	
	if todo.Completed {
		r.client.SRem(TodosCompletedSet, id)
	} else {
		r.client.SRem(TodosPendingSet, id)
	}

	// Remove from priority set
	priorityKey := TodosPriorityPrefix + string(todo.Priority)
	r.client.SRem(priorityKey, id)

	// Remove from tag sets
	for _, tag := range todo.Tags {
		tagKey := TodosTagPrefix + tag
		r.client.SRem(tagKey, id)
	}

	return nil
}

// ToggleComplete toggles the completion status of a todo
func (r *TodoRepository) ToggleComplete(id string) (*models.Todo, error) {
	todo, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Toggle completion status
	updates := &models.UpdateTodoRequest{
		Completed: &[]bool{!todo.Completed}[0],
	}

	return r.Update(id, updates)
}

// GetStats returns todo statistics
func (r *TodoRepository) GetStats() (*models.TodoStatsResponse, error) {
	allIDs, err := r.client.SMembers(TodosAllSet)
	if err != nil {
		return nil, fmt.Errorf("failed to get all todos: %w", err)
	}

	completedIDs, err := r.client.SMembers(TodosCompletedSet)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed todos: %w", err)
	}

	highPriorityIDs, err := r.client.SMembers(TodosPriorityPrefix + string(models.PriorityHigh))
	if err != nil {
		return nil, fmt.Errorf("failed to get high priority todos: %w", err)
	}

	stats := &models.TodoStatsResponse{
		Total:        len(allIDs),
		Completed:    len(completedIDs),
		Pending:      len(allIDs) - len(completedIDs),
		HighPriority: len(highPriorityIDs),
	}

	return stats, nil
}

// sortTodos sorts the todos slice based on the given criteria
func (r *TodoRepository) sortTodos(todos []*models.Todo, sortBy, order string) {
	ascending := order != "desc"

	sort.Slice(todos, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "title":
			less = strings.ToLower(todos[i].Title) < strings.ToLower(todos[j].Title)
		case "priority":
			priorityOrder := map[models.Priority]int{
				models.PriorityLow:    1,
				models.PriorityMedium: 2,
				models.PriorityHigh:   3,
			}
			less = priorityOrder[todos[i].Priority] < priorityOrder[todos[j].Priority]
		case "dueDate":
			if todos[i].DueDate == nil && todos[j].DueDate == nil {
				less = false
			} else if todos[i].DueDate == nil {
				less = false
			} else if todos[j].DueDate == nil {
				less = true
			} else {
				less = todos[i].DueDate.Before(*todos[j].DueDate)
			}
		default: // "createdAt"
			less = todos[i].CreatedAt.Before(todos[j].CreatedAt)
		}

		if ascending {
			return less
		}
		return !less
	})
}