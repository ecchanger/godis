import { renderHook, act } from '@testing-library/react';
import { useTodoStore } from '../todoStore';

// Mock the API service
jest.mock('../../services/todoService', () => ({
  __esModule: true,
  default: {
    getAllTodos: jest.fn(() => Promise.resolve({ todos: [], total: 0 })),
    createTodo: jest.fn((todo) => Promise.resolve({ 
      id: 'test-id', 
      ...todo, 
      completed: false,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    })),
    updateTodo: jest.fn((id, updates) => Promise.resolve({ 
      id, 
      ...updates,
      updatedAt: new Date().toISOString(),
    })),
    deleteTodo: jest.fn(() => Promise.resolve()),
    toggleTodo: jest.fn((id) => Promise.resolve({ 
      id, 
      completed: true,
      updatedAt: new Date().toISOString(),
    })),
    getStats: jest.fn(() => Promise.resolve({
      total: 0,
      completed: 0,
      pending: 0,
      highPriority: 0,
    })),
  },
  handleApiError: jest.fn((error) => error.message),
}));

describe('TodoStore', () => {
  beforeEach(() => {
    // Reset the store state before each test
    const { result } = renderHook(() => useTodoStore());
    act(() => {
      result.current.clearError();
      // Reset todos array (this would need to be implemented in the store)
    });
  });

  test('initial state is correct', () => {
    const { result } = renderHook(() => useTodoStore());
    
    expect(result.current.todos).toEqual([]);
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBe(null);
    expect(result.current.filter.sortBy).toBe('createdAt');
    expect(result.current.filter.order).toBe('desc');
  });

  test('setLoading updates loading state', () => {
    const { result } = renderHook(() => useTodoStore());
    
    act(() => {
      result.current.setLoading(true);
    });
    
    expect(result.current.loading).toBe(true);
  });

  test('setError updates error state', () => {
    const { result } = renderHook(() => useTodoStore());
    const errorMessage = 'Test error';
    
    act(() => {
      result.current.setError(errorMessage);
    });
    
    expect(result.current.error).toBe(errorMessage);
  });

  test('clearError clears error state', () => {
    const { result } = renderHook(() => useTodoStore());
    
    act(() => {
      result.current.setError('Test error');
    });
    
    expect(result.current.error).toBe('Test error');
    
    act(() => {
      result.current.clearError();
    });
    
    expect(result.current.error).toBe(null);
  });

  test('setFilter updates filter state', () => {
    const { result } = renderHook(() => useTodoStore());
    
    act(() => {
      result.current.setFilter({ completed: true });
    });
    
    expect(result.current.filter.completed).toBe(true);
  });

  test('clearFilter resets filter to default', () => {
    const { result } = renderHook(() => useTodoStore());
    
    // Set some filters first
    act(() => {
      result.current.setFilter({ completed: true, priority: 'high' });
    });
    
    // Clear filters
    act(() => {
      result.current.clearFilter();
    });
    
    expect(result.current.filter.completed).toBeUndefined();
    expect(result.current.filter.priority).toBeUndefined();
    expect(result.current.filter.sortBy).toBe('createdAt');
    expect(result.current.filter.order).toBe('desc');
  });

  test('fetchTodos calls API and updates state', async () => {
    const { result } = renderHook(() => useTodoStore());
    
    await act(async () => {
      await result.current.fetchTodos();
    });
    
    // The mock API returns empty array, so todos should be empty
    expect(result.current.todos).toEqual([]);
    expect(result.current.loading).toBe(false);
  });

  test('createTodo calls API and adds todo to state', async () => {
    const { result } = renderHook(() => useTodoStore());
    const newTodo = {
      title: 'Test Todo',
      description: 'Test Description',
      priority: 'high',
    };
    
    await act(async () => {
      await result.current.createTodo(newTodo);
    });
    
    // Check that the todo was added (assuming the mock returns the expected structure)
    expect(result.current.todos).toHaveLength(1);
    expect(result.current.todos[0].title).toBe('Test Todo');
  });
});