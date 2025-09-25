import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import TodoService, { handleApiError } from '../services/todoService';
import {
  TodoStore,
  TodoFilter,
  CreateTodoRequest,
  UpdateTodoRequest,
} from '../types';

// Initial state
const initialState = {
  todos: [],
  loading: false,
  error: null,
  filter: {
    sortBy: 'createdAt',
    order: 'desc',
  },
  stats: null,
};

// Create the store
export const useTodoStore = create<TodoStore>()(
  devtools(
    (set, get) => ({
      ...initialState,

      // Data operations
      fetchTodos: async () => {
        const { filter } = get();
        set({ loading: true, error: null });
        
        try {
          const response = await TodoService.getAllTodos(filter);
          set({ 
            todos: response.todos,
            loading: false 
          });
        } catch (error) {
          set({ 
            error: handleApiError(error),
            loading: false 
          });
        }
      },

      createTodo: async (todoData) => {
        set({ loading: true, error: null });
        
        try {
          const newTodo = await TodoService.createTodo(todoData);
          const { todos } = get();
          set({ 
            todos: [newTodo, ...todos],
            loading: false 
          });
          
          // Refresh stats after creating
          get().fetchStats();
        } catch (error) {
          set({ 
            error: handleApiError(error),
            loading: false 
          });
          throw error; // Re-throw to handle in UI
        }
      },

      updateTodo: async (id, updates) => {
        set({ loading: true, error: null });
        
        try {
          const updatedTodo = await TodoService.updateTodo(id, updates);
          const { todos } = get();
          const newTodos = todos.map(todo => 
            todo.id === id ? updatedTodo : todo
          );
          
          set({ 
            todos: newTodos,
            loading: false 
          });
          
          // Refresh stats if completion status changed
          if (updates.completed !== undefined) {
            get().fetchStats();
          }
        } catch (error) {
          set({ 
            error: handleApiError(error),
            loading: false 
          });
          throw error;
        }
      },

      deleteTodo: async (id) => {
        set({ loading: true, error: null });
        
        try {
          await TodoService.deleteTodo(id);
          const { todos } = get();
          const newTodos = todos.filter(todo => todo.id !== id);
          
          set({ 
            todos: newTodos,
            loading: false 
          });
          
          // Refresh stats after deleting
          get().fetchStats();
        } catch (error) {
          set({ 
            error: handleApiError(error),
            loading: false 
          });
          throw error;
        }
      },

      toggleTodo: async (id) => {
        set({ error: null });
        
        try {
          const updatedTodo = await TodoService.toggleTodo(id);
          const { todos } = get();
          const newTodos = todos.map(todo => 
            todo.id === id ? updatedTodo : todo
          );
          
          set({ todos: newTodos });
          
          // Refresh stats after toggling
          get().fetchStats();
        } catch (error) {
          set({ error: handleApiError(error) });
          throw error;
        }
      },

      fetchStats: async () => {
        try {
          const stats = await TodoService.getStats();
          set({ stats });
        } catch (error) {
          console.error('Failed to fetch stats:', handleApiError(error));
          // Don't set error for stats failure as it's not critical
        }
      },

      // Filter operations
      setFilter: (newFilter) => {
        const { filter } = get();
        const updatedFilter = { ...filter, ...newFilter };
        set({ filter: updatedFilter });
        
        // Automatically fetch todos with new filter
        get().fetchTodos();
      },

      clearFilter: () => {
        const clearedFilter = {
          sortBy: 'createdAt',
          order: 'desc',
        };
        set({ filter: clearedFilter });
        
        // Automatically fetch todos with cleared filter
        get().fetchTodos();
      },

      // UI operations
      setLoading: (loading) => set({ loading }),
      
      setError: (error) => set({ error }),
      
      clearError: () => set({ error: null }),
    }),
    {
      name: 'todo-store', // Store name for devtools
    }
  )
);

// Selector hooks for better performance
export const useTodos = () => useTodoStore((state) => state.todos);
export const useLoading = () => useTodoStore((state) => state.loading);
export const useError = () => useTodoStore((state) => state.error);
export const useFilter = () => useTodoStore((state) => state.filter);
export const useStats = () => useTodoStore((state) => state.stats);

// Action hooks
export const useTodoActions = () => useTodoStore((state) => ({
  fetchTodos: state.fetchTodos,
  createTodo: state.createTodo,
  updateTodo: state.updateTodo,
  deleteTodo: state.deleteTodo,
  toggleTodo: state.toggleTodo,
  fetchStats: state.fetchStats,
  setFilter: state.setFilter,
  clearFilter: state.clearFilter,
  setLoading: state.setLoading,
  setError: state.setError,
  clearError: state.clearError,
}));

// Computed selectors
export const useFilteredTodos = () => {
  return useTodoStore((state) => {
    const { todos, filter } = state;
    
    let filtered = [...todos];
    
    // Apply completion filter
    if (filter.completed !== undefined) {
      filtered = filtered.filter(todo => todo.completed === filter.completed);
    }
    
    // Apply priority filter
    if (filter.priority) {
      filtered = filtered.filter(todo => todo.priority === filter.priority);
    }
    
    // Apply sorting
    if (filter.sortBy) {
      filtered.sort((a, b) => {
        let comparison = 0;
        
        switch (filter.sortBy) {
          case 'title':
            comparison = a.title.localeCompare(b.title);
            break;
          case 'priority':
            const priorityOrder = { low: 1, medium: 2, high: 3 };
            comparison = priorityOrder[a.priority] - priorityOrder[b.priority];
            break;
          case 'dueDate':
            const aDate = a.dueDate ? new Date(a.dueDate) : new Date('9999-12-31');
            const bDate = b.dueDate ? new Date(b.dueDate) : new Date('9999-12-31');
            comparison = aDate.getTime() - bDate.getTime();
            break;
          case 'createdAt':
          default:
            comparison = new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime();
            break;
        }
        
        return filter.order === 'desc' ? -comparison : comparison;
      });
    }
    
    return filtered;
  });
};

// Export the store as default
export default useTodoStore;