import axios, { AxiosResponse } from 'axios';
import {
  Todo,
  CreateTodoRequest,
  UpdateTodoRequest,
  TodoFilter,
  APIResponse,
  TodoListResponse,
  TodoStatsResponse,
} from '../types';

// Create axios instance with base configuration
const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Response interceptor to handle API responses
api.interceptors.response.use(
  (response: AxiosResponse<APIResponse>) => {
    // If the response is successful, return the data
    if (response.data.success) {
      return response;
    }
    // If the API returns success: false, throw an error
    throw new Error(response.data.error || 'API request failed');
  },
  (error) => {
    // Handle network errors or HTTP error status codes
    const message = error.response?.data?.error || error.message || 'Network error';
    throw new Error(message);
  }
);

// Todo API service class
export class TodoService {
  // Get all todos with optional filtering
  static async getAllTodos(filter?: TodoFilter): Promise<TodoListResponse> {
    const params = new URLSearchParams();
    
    if (filter) {
      if (filter.completed !== undefined) {
        params.append('completed', filter.completed.toString());
      }
      if (filter.priority) {
        params.append('priority', filter.priority);
      }
      if (filter.tag) {
        params.append('tag', filter.tag);
      }
      if (filter.sortBy) {
        params.append('sortBy', filter.sortBy);
      }
      if (filter.order) {
        params.append('order', filter.order);
      }
    }
    
    const queryString = params.toString();
    const url = queryString ? `/todos?${queryString}` : '/todos';
    
    const response = await api.get<APIResponse<TodoListResponse>>(url);
    return response.data.data!;
  }

  // Get a specific todo by ID
  static async getTodoById(id: string): Promise<Todo> {
    const response = await api.get<APIResponse<Todo>>(`/todos/${id}`);
    return response.data.data!;
  }

  // Create a new todo
  static async createTodo(todo: CreateTodoRequest): Promise<Todo> {
    const response = await api.post<APIResponse<Todo>>('/todos', todo);
    return response.data.data!;
  }

  // Update an existing todo
  static async updateTodo(id: string, updates: UpdateTodoRequest): Promise<Todo> {
    const response = await api.put<APIResponse<Todo>>(`/todos/${id}`, updates);
    return response.data.data!;
  }

  // Delete a todo
  static async deleteTodo(id: string): Promise<void> {
    await api.delete(`/todos/${id}`);
  }

  // Toggle todo completion status
  static async toggleTodo(id: string): Promise<Todo> {
    const response = await api.patch<APIResponse<Todo>>(`/todos/${id}/toggle`);
    return response.data.data!;
  }

  // Get todo statistics
  static async getStats(): Promise<TodoStatsResponse> {
    const response = await api.get<APIResponse<TodoStatsResponse>>('/stats');
    return response.data.data!;
  }

  // Health check
  static async healthCheck(): Promise<boolean> {
    try {
      await api.get('/health');
      return true;
    } catch (error) {
      return false;
    }
  }
}

// Export utility functions for error handling
export const handleApiError = (error: any): string => {
  if (error.response?.data?.error) {
    return error.response.data.error;
  }
  if (error.message) {
    return error.message;
  }
  return 'An unexpected error occurred';
};

// Export the axios instance for custom requests if needed
export { api };

// Default export
export default TodoService;