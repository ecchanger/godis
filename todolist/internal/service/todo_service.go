package service

import (
	"errors"
	"sort"
	"sync"
	"time"

	"todolist/internal/model"
)

var (
	ErrTodoNotFound = errors.New("todo not found")
)

// TodoService 待办事项服务
type TodoService struct {
	todos  map[int]*model.Todo
	nextID int
	mutex  sync.RWMutex
}

// NewTodoService 创建新的待办事项服务
func NewTodoService() *TodoService {
	return &TodoService{
		todos:  make(map[int]*model.Todo),
		nextID: 1,
	}
}

// Create 创建新的待办事项
func (s *TodoService) Create(req *model.TodoRequest) *model.Todo {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	todo := &model.Todo{
		ID:          s.nextID,
		Title:       req.Title,
		Description: req.Description,
		Completed:   req.Completed,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.todos[s.nextID] = todo
	s.nextID++

	return todo
}

// GetAll 获取所有待办事项
func (s *TodoService) GetAll() []*model.Todo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	todos := make([]*model.Todo, 0, len(s.todos))
	for _, todo := range s.todos {
		todos = append(todos, todo)
	}

	// 按创建时间排序，最新的在前面
	sort.Slice(todos, func(i, j int) bool {
		return todos[i].CreatedAt.After(todos[j].CreatedAt)
	})

	return todos
}

// GetByID 根据ID获取待办事项
func (s *TodoService) GetByID(id int) (*model.Todo, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	todo, exists := s.todos[id]
	if !exists {
		return nil, ErrTodoNotFound
	}

	return todo, nil
}

// Update 更新待办事项
func (s *TodoService) Update(id int, req *model.TodoRequest) (*model.Todo, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	todo, exists := s.todos[id]
	if !exists {
		return nil, ErrTodoNotFound
	}

	todo.Title = req.Title
	todo.Description = req.Description
	todo.Completed = req.Completed
	todo.UpdatedAt = time.Now()

	return todo, nil
}

// Delete 删除待办事项
func (s *TodoService) Delete(id int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.todos[id]; !exists {
		return ErrTodoNotFound
	}

	delete(s.todos, id)
	return nil
}

// ToggleComplete 切换完成状态
func (s *TodoService) ToggleComplete(id int) (*model.Todo, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	todo, exists := s.todos[id]
	if !exists {
		return nil, ErrTodoNotFound
	}

	todo.Completed = !todo.Completed
	todo.UpdatedAt = time.Now()

	return todo, nil
}