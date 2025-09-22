package webtodo

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hdt3213/godis/lib/logger"
	"github.com/hdt3213/godis/redis/client"
	"github.com/hdt3213/godis/redis/protocol"
)

// Common errors
var (
	ErrTodoNotFound = errors.New("todo not found")
	ErrInvalidID    = errors.New("invalid todo ID")
	ErrUserNotFound = errors.New("user not found")
)

// Todo represents a todo item
type Todo struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	Priority    string    `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// CreateTodoRequest represents a request to create a new todo
type CreateTodoRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
}

// UpdateTodoRequest represents a request to update a todo
type UpdateTodoRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Completed   *bool   `json:"completed,omitempty"`
	Priority    *string `json:"priority,omitempty"`
	DueDate     *string `json:"due_date,omitempty"`
}

// TodoFilter represents filtering options for todos
type TodoFilter struct {
	UserID    string
	Completed *bool
	Priority  string
	Limit     int
	Offset    int
}

// TodoService handles todo operations with Redis
type TodoService struct {
	redisClient *client.Client
}

// NewTodoService creates a new todo service
func NewTodoService(redisClient *client.Client) *TodoService {
	return &TodoService{
		redisClient: redisClient,
	}
}

// Redis key patterns following the design document
const (
	TodoHashKeyPattern     = "todo:%s:%s"        // todo:{user_id}:{todo_id}
	UserTodoListKey        = "todos:user:%s"     // todos:user:{user_id}
	TodoCounterKey         = "todo:counter:%s"   // todo:counter:{user_id}
	PriorityIndexKey       = "todos:priority:%s:%s" // todos:priority:{priority}:{user_id}
	DueDateIndexKey        = "todos:due:%s:%s"   // todos:due:{date}:{user_id}
)

// CreateTodo creates a new todo item
func (s *TodoService) CreateTodo(userID string, request *CreateTodoRequest) (*Todo, error) {
	// Generate unique ID using Redis counter
	todoID, err := s.generateTodoID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate todo ID: %v", err)
	}

	now := time.Now()
	todo := &Todo{
		ID:          todoID,
		Title:       strings.TrimSpace(request.Title),
		Description: strings.TrimSpace(request.Description),
		Completed:   false,
		Priority:    strings.ToLower(request.Priority),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Parse due date if provided
	if request.DueDate != "" {
		dueDate, err := time.Parse(time.RFC3339, request.DueDate)
		if err != nil {
			return nil, fmt.Errorf("invalid due date format: %v", err)
		}
		todo.DueDate = &dueDate
	}

	// Set default priority if not provided
	if todo.Priority == "" {
		todo.Priority = "medium"
	}

	// Store todo in Redis using transaction-like operations
	err = s.storeTodo(userID, todo)
	if err != nil {
		return nil, fmt.Errorf("failed to store todo: %v", err)
	}

	return todo, nil
}

// GetTodo retrieves a specific todo
func (s *TodoService) GetTodo(userID, todoID string) (*Todo, error) {
	todoKey := fmt.Sprintf(TodoHashKeyPattern, userID, todoID)
	
	// Get all fields from the hash
	reply := s.redisClient.Send([][]byte{
		[]byte("HGETALL"),
		[]byte(todoKey),
	})

	if errReply, ok := reply.(*protocol.StandardErrReply); ok {
		return nil, fmt.Errorf("redis error: %s", errReply.Error())
	}

	multiBulkReply, ok := reply.(*protocol.MultiBulkReply)
	if !ok {
		return nil, fmt.Errorf("unexpected reply type")
	}

	if len(multiBulkReply.Args) == 0 {
		return nil, ErrTodoNotFound
	}

	// Parse hash fields into Todo struct
	todo, err := s.parseHashToTodo(multiBulkReply.Args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse todo: %v", err)
	}

	return todo, nil
}

// GetTodos retrieves todos based on filter criteria
func (s *TodoService) GetTodos(filter TodoFilter) ([]*Todo, error) {
	userListKey := fmt.Sprintf(UserTodoListKey, filter.UserID)
	
	// Determine range for pagination
	start := int64(filter.Offset)
	end := int64(-1) // Get all by default
	if filter.Limit > 0 {
		end = start + int64(filter.Limit) - 1
	}

	// Get todo IDs from the user's todo list
	reply := s.redisClient.Send([][]byte{
		[]byte("LRANGE"),
		[]byte(userListKey),
		[]byte(strconv.FormatInt(start, 10)),
		[]byte(strconv.FormatInt(end, 10)),
	})

	if errReply, ok := reply.(*protocol.StandardErrReply); ok {
		return nil, fmt.Errorf("redis error: %s", errReply.Error())
	}

	multiBulkReply, ok := reply.(*protocol.MultiBulkReply)
	if !ok {
		return nil, fmt.Errorf("unexpected reply type")
	}

	var todos []*Todo
	for _, arg := range multiBulkReply.Args {
		if bulkReply, ok := arg.(*protocol.BulkReply); ok {
			todoID := string(bulkReply.Arg)
			todo, err := s.GetTodo(filter.UserID, todoID)
			if err != nil {
				if err == ErrTodoNotFound {
					// Todo might have been deleted, skip it
					continue
				}
				logger.Errorf("Error getting todo %s: %v", todoID, err)
				continue
			}

			// Apply filters
			if s.matchesFilter(todo, filter) {
				todos = append(todos, todo)
			}
		}
	}

	return todos, nil
}

// UpdateTodo updates an existing todo
func (s *TodoService) UpdateTodo(userID, todoID string, request *UpdateTodoRequest) (*Todo, error) {
	// First, get the existing todo
	todo, err := s.GetTodo(userID, todoID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if request.Title != nil {
		todo.Title = strings.TrimSpace(*request.Title)
	}
	if request.Description != nil {
		todo.Description = strings.TrimSpace(*request.Description)
	}
	if request.Completed != nil {
		todo.Completed = *request.Completed
	}
	if request.Priority != nil {
		todo.Priority = strings.ToLower(*request.Priority)
	}
	if request.DueDate != nil {
		if *request.DueDate == "" {
			todo.DueDate = nil
		} else {
			dueDate, err := time.Parse(time.RFC3339, *request.DueDate)
			if err != nil {
				return nil, fmt.Errorf("invalid due date format: %v", err)
			}
			todo.DueDate = &dueDate
		}
	}

	todo.UpdatedAt = time.Now()

	// Update todo in Redis
	err = s.updateTodoInRedis(userID, todo)
	if err != nil {
		return nil, fmt.Errorf("failed to update todo: %v", err)
	}

	return todo, nil
}

// DeleteTodo deletes a todo
func (s *TodoService) DeleteTodo(userID, todoID string) error {
	// Check if todo exists
	_, err := s.GetTodo(userID, todoID)
	if err != nil {
		return err
	}

	// Delete todo hash
	todoKey := fmt.Sprintf(TodoHashKeyPattern, userID, todoID)
	reply := s.redisClient.Send([][]byte{
		[]byte("DEL"),
		[]byte(todoKey),
	})

	if errReply, ok := reply.(*protocol.StandardErrReply); ok {
		return fmt.Errorf("redis error: %s", errReply.Error())
	}

	// Remove from user's todo list
	userListKey := fmt.Sprintf(UserTodoListKey, userID)
	reply = s.redisClient.Send([][]byte{
		[]byte("LREM"),
		[]byte(userListKey),
		[]byte("0"), // Remove all occurrences
		[]byte(todoID),
	})

	if errReply, ok := reply.(*protocol.StandardErrReply); ok {
		logger.Errorf("Error removing todo from list: %s", errReply.Error())
	}

	return nil
}

// Helper methods

// generateTodoID generates a unique todo ID using Redis counter
func (s *TodoService) generateTodoID(userID string) (string, error) {
	counterKey := fmt.Sprintf(TodoCounterKey, userID)
	reply := s.redisClient.Send([][]byte{
		[]byte("INCR"),
		[]byte(counterKey),
	})

	if errReply, ok := reply.(*protocol.StandardErrReply); ok {
		return "", fmt.Errorf("redis error: %s", errReply.Error())
	}

	intReply, ok := reply.(*protocol.IntReply)
	if !ok {
		return "", fmt.Errorf("unexpected reply type for counter")
	}

	return strconv.FormatInt(intReply.Code, 10), nil
}

// storeTodo stores a todo item in Redis
func (s *TodoService) storeTodo(userID string, todo *Todo) error {
	todoKey := fmt.Sprintf(TodoHashKeyPattern, userID, todo.ID)
	userListKey := fmt.Sprintf(UserTodoListKey, userID)

	// Serialize todo fields
	fields := s.todoToHashFields(todo)

	// Store todo hash
	args := [][]byte{[]byte("HMSET"), []byte(todoKey)}
	for key, value := range fields {
		args = append(args, []byte(key), []byte(value))
	}

	reply := s.redisClient.Send(args)
	if errReply, ok := reply.(*protocol.StandardErrReply); ok {
		return fmt.Errorf("redis error storing todo: %s", errReply.Error())
	}

	// Add to user's todo list (prepend for chronological order)
	reply = s.redisClient.Send([][]byte{
		[]byte("LPUSH"),
		[]byte(userListKey),
		[]byte(todo.ID),
	})

	if errReply, ok := reply.(*protocol.StandardErrReply); ok {
		return fmt.Errorf("redis error adding to list: %s", errReply.Error())
	}

	return nil
}

// updateTodoInRedis updates a todo in Redis
func (s *TodoService) updateTodoInRedis(userID string, todo *Todo) error {
	todoKey := fmt.Sprintf(TodoHashKeyPattern, userID, todo.ID)
	fields := s.todoToHashFields(todo)

	// Update todo hash
	args := [][]byte{[]byte("HMSET"), []byte(todoKey)}
	for key, value := range fields {
		args = append(args, []byte(key), []byte(value))
	}

	reply := s.redisClient.Send(args)
	if errReply, ok := reply.(*protocol.StandardErrReply); ok {
		return fmt.Errorf("redis error updating todo: %s", errReply.Error())
	}

	return nil
}

// todoToHashFields converts a Todo struct to Redis hash fields
func (s *TodoService) todoToHashFields(todo *Todo) map[string]string {
	fields := map[string]string{
		"id":          todo.ID,
		"title":       todo.Title,
		"description": todo.Description,
		"completed":   strconv.FormatBool(todo.Completed),
		"priority":    todo.Priority,
		"created_at":  todo.CreatedAt.Format(time.RFC3339),
		"updated_at":  todo.UpdatedAt.Format(time.RFC3339),
	}

	if todo.DueDate != nil {
		fields["due_date"] = todo.DueDate.Format(time.RFC3339)
	}

	return fields
}

// parseHashToTodo converts Redis hash fields to a Todo struct
func (s *TodoService) parseHashToTodo(args []*protocol.BulkReply) (*Todo, error) {
	if len(args)%2 != 0 {
		return nil, fmt.Errorf("invalid hash format")
	}

	fields := make(map[string]string)
	for i := 0; i < len(args); i += 2 {
		key := string(args[i].Arg)
		value := string(args[i+1].Arg)
		fields[key] = value
	}

	todo := &Todo{}
	
	// Parse required fields
	todo.ID = fields["id"]
	todo.Title = fields["title"]
	todo.Description = fields["description"]
	todo.Priority = fields["priority"]

	// Parse boolean completed field
	if completedStr, exists := fields["completed"]; exists {
		completed, err := strconv.ParseBool(completedStr)
		if err != nil {
			return nil, fmt.Errorf("invalid completed value: %v", err)
		}
		todo.Completed = completed
	}

	// Parse timestamps
	if createdAtStr, exists := fields["created_at"]; exists {
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("invalid created_at value: %v", err)
		}
		todo.CreatedAt = createdAt
	}

	if updatedAtStr, exists := fields["updated_at"]; exists {
		updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, fmt.Errorf("invalid updated_at value: %v", err)
		}
		todo.UpdatedAt = updatedAt
	}

	// Parse optional due date
	if dueDateStr, exists := fields["due_date"]; exists && dueDateStr != "" {
		dueDate, err := time.Parse(time.RFC3339, dueDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid due_date value: %v", err)
		}
		todo.DueDate = &dueDate
	}

	return todo, nil
}

// matchesFilter checks if a todo matches the given filter criteria
func (s *TodoService) matchesFilter(todo *Todo, filter TodoFilter) bool {
	// Check completion status filter
	if filter.Completed != nil && todo.Completed != *filter.Completed {
		return false
	}

	// Check priority filter
	if filter.Priority != "" && todo.Priority != strings.ToLower(filter.Priority) {
		return false
	}

	return true
}