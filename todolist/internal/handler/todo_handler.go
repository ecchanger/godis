package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"todolist/internal/model"
	"todolist/internal/service"
)

// TodoHandler 待办事项处理器
type TodoHandler struct {
	service *service.TodoService
}

// NewTodoHandler 创建新的待办事项处理器
func NewTodoHandler(service *service.TodoService) *TodoHandler {
	return &TodoHandler{
		service: service,
	}
}

// CreateTodo 创建待办事项
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var req model.TodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		h.sendErrorResponse(w, "Title is required", http.StatusBadRequest)
		return
	}

	todo := h.service.Create(&req)
	h.sendSuccessResponse(w, "Todo created successfully", todo)
}

// GetAllTodos 获取所有待办事项
func (h *TodoHandler) GetAllTodos(w http.ResponseWriter, r *http.Request) {
	todos := h.service.GetAll()
	h.sendSuccessResponse(w, "Todos retrieved successfully", todos)
}

// GetTodo 获取单个待办事项
func (h *TodoHandler) GetTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.sendErrorResponse(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	todo, err := h.service.GetByID(id)
	if err != nil {
		h.sendErrorResponse(w, "Todo not found", http.StatusNotFound)
		return
	}

	h.sendSuccessResponse(w, "Todo retrieved successfully", todo)
}

// UpdateTodo 更新待办事项
func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.sendErrorResponse(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	var req model.TodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		h.sendErrorResponse(w, "Title is required", http.StatusBadRequest)
		return
	}

	todo, err := h.service.Update(id, &req)
	if err != nil {
		h.sendErrorResponse(w, "Todo not found", http.StatusNotFound)
		return
	}

	h.sendSuccessResponse(w, "Todo updated successfully", todo)
}

// DeleteTodo 删除待办事项
func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.sendErrorResponse(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	err = h.service.Delete(id)
	if err != nil {
		h.sendErrorResponse(w, "Todo not found", http.StatusNotFound)
		return
	}

	h.sendSuccessResponse(w, "Todo deleted successfully", nil)
}

// ToggleComplete 切换完成状态
func (h *TodoHandler) ToggleComplete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.sendErrorResponse(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	todo, err := h.service.ToggleComplete(id)
	if err != nil {
		h.sendErrorResponse(w, "Todo not found", http.StatusNotFound)
		return
	}

	h.sendSuccessResponse(w, "Todo status toggled successfully", todo)
}

// ServeIndex 提供首页
func (h *TodoHandler) ServeIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/templates/index.html")
}

// sendSuccessResponse 发送成功响应
func (h *TodoHandler) sendSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response := model.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(response)
}

// sendErrorResponse 发送错误响应
func (h *TodoHandler) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := model.APIResponse{
		Success: false,
		Message: message,
	}
	json.NewEncoder(w).Encode(response)
}