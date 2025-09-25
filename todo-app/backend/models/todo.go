package models

import (
	"encoding/json"
	"time"
)

// Priority represents the priority level of a todo item
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// Todo represents a todo item with all its properties
type Todo struct {
	ID          string     `json:"id"`
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	Completed   bool       `json:"completed"`
	Priority    Priority   `json:"priority"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// ToJSON converts Todo to JSON bytes
func (t *Todo) ToJSON() ([]byte, error) {
	return json.Marshal(t)
}

// FromJSON creates Todo from JSON bytes
func FromJSON(data []byte) (*Todo, error) {
	var todo Todo
	err := json.Unmarshal(data, &todo)
	if err != nil {
		return nil, err
	}
	return &todo, nil
}

// IsValidPriority checks if the priority is valid
func IsValidPriority(p Priority) bool {
	return p == PriorityLow || p == PriorityMedium || p == PriorityHigh
}

// CreateTodoRequest represents the request body for creating a new todo
type CreateTodoRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	Priority    Priority   `json:"priority"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
}

// UpdateTodoRequest represents the request body for updating a todo
type UpdateTodoRequest struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	Completed   *bool      `json:"completed,omitempty"`
	Priority    *Priority  `json:"priority,omitempty"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
}

// TodoFilter represents filtering and sorting options
type TodoFilter struct {
	Completed *bool     `form:"completed"`
	Priority  *Priority `form:"priority"`
	SortBy    string    `form:"sortBy"` // "createdAt", "priority", "dueDate", "title"
	Order     string    `form:"order"`  // "asc", "desc"
}