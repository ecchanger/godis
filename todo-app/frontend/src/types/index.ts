// Todo types
export type Priority = 'low' | 'medium' | 'high';

export interface Todo {
  id: string;
  title: string;
  description: string;
  completed: boolean;
  priority: Priority;
  dueDate?: string;
  createdAt: string;
  updatedAt: string;
}

// API Request types
export interface CreateTodoRequest {
  title: string;
  description?: string;
  priority?: Priority;
  dueDate?: string;
}

export interface UpdateTodoRequest {
  title?: string;
  description?: string;
  completed?: boolean;
  priority?: Priority;
  dueDate?: string;
}

// API Response types
export interface APIResponse<T = any> {
  success: boolean;
  data?: T;
  message?: string;
  error?: string;
  code?: string;
}

export interface TodoListResponse {
  todos: Todo[];
  total: number;
}

export interface TodoStatsResponse {
  total: number;
  completed: number;
  pending: number;
  highPriority: number;
}

// Filter and sorting types
export interface TodoFilter {
  completed?: boolean;
  priority?: Priority;
  sortBy?: 'createdAt' | 'title' | 'priority' | 'dueDate';
  order?: 'asc' | 'desc';
}

// Store types
export interface TodoState {
  todos: Todo[];
  loading: boolean;
  error: string | null;
  filter: TodoFilter;
  stats: TodoStatsResponse | null;
}

export interface TodoActions {
  // Data operations
  fetchTodos: () => Promise<void>;
  createTodo: (todo: CreateTodoRequest) => Promise<void>;
  updateTodo: (id: string, updates: UpdateTodoRequest) => Promise<void>;
  deleteTodo: (id: string) => Promise<void>;
  toggleTodo: (id: string) => Promise<void>;
  fetchStats: () => Promise<void>;
  
  // Filter operations
  setFilter: (filter: Partial<TodoFilter>) => void;
  clearFilter: () => void;
  
  // UI operations
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  clearError: () => void;
}

export type TodoStore = TodoState & TodoActions;

// Component props types
export interface TodoItemProps {
  todo: Todo;
  onToggle: (id: string) => void;
  onEdit: (todo: Todo) => void;
  onDelete: (id: string) => void;
}

export interface TodoFormProps {
  todo?: Todo;
  onSubmit: (todo: CreateTodoRequest | UpdateTodoRequest) => void;
  onCancel: () => void;
  isLoading?: boolean;
}

export interface FilterProps {
  filter: TodoFilter;
  onFilterChange: (filter: Partial<TodoFilter>) => void;
  onClearFilter: () => void;
}

export interface HeaderProps {
  stats: TodoStatsResponse | null;
  onAddTodo: () => void;
}

// Utility types
export type SortDirection = 'asc' | 'desc';
export type SortField = 'createdAt' | 'title' | 'priority' | 'dueDate';

// Modal types
export interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
}