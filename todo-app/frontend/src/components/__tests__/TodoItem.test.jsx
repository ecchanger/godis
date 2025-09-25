import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import TodoItem from '../TodoItem';

// Mock the store
jest.mock('../../store/todoStore', () => ({
  useTodoActions: () => ({
    toggleTodo: jest.fn(),
    deleteTodo: jest.fn(),
  }),
}));

const mockTodo = {
  id: 'test-id',
  title: 'Test Todo',
  description: 'Test Description',
  completed: false,
  priority: 'high',
  createdAt: '2023-01-01T00:00:00Z',
  updatedAt: '2023-01-01T00:00:00Z',
};

const mockOnEdit = jest.fn();

describe('TodoItem', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders todo item correctly', () => {
    render(<TodoItem todo={mockTodo} onEdit={mockOnEdit} />);
    
    expect(screen.getByText('Test Todo')).toBeInTheDocument();
    expect(screen.getByText('Test Description')).toBeInTheDocument();
    expect(screen.getByText('high')).toBeInTheDocument();
  });

  test('shows completed state correctly', () => {
    const completedTodo = { ...mockTodo, completed: true };
    render(<TodoItem todo={completedTodo} onEdit={mockOnEdit} />);
    
    const titleElement = screen.getByText('Test Todo');
    expect(titleElement).toHaveClass('line-through');
  });

  test('calls onEdit when edit button is clicked', () => {
    render(<TodoItem todo={mockTodo} onEdit={mockOnEdit} />);
    
    const editButton = screen.getByTitle('Edit todo');
    fireEvent.click(editButton);
    
    expect(mockOnEdit).toHaveBeenCalledWith(mockTodo);
  });

  test('shows priority badge with correct color', () => {
    render(<TodoItem todo={mockTodo} onEdit={mockOnEdit} />);
    
    const priorityBadge = screen.getByText('high');
    expect(priorityBadge).toHaveClass('bg-red-100', 'text-red-800');
  });

  test('displays due date when present', () => {
    const todoWithDueDate = {
      ...mockTodo,
      dueDate: '2023-12-31T23:59:59Z',
    };
    
    render(<TodoItem todo={todoWithDueDate} onEdit={mockOnEdit} />);
    
    expect(screen.getByText(/Due/)).toBeInTheDocument();
  });

  test('shows overdue indicator for past due date', () => {
    const pastDate = new Date();
    pastDate.setDate(pastDate.getDate() - 1);
    
    const overdueTodo = {
      ...mockTodo,
      dueDate: pastDate.toISOString(),
      completed: false,
    };
    
    render(<TodoItem todo={overdueTodo} onEdit={mockOnEdit} />);
    
    // Should show the overdue indicator (red text and alert icon)
    const dueDateElement = screen.getByText(/Due/);
    expect(dueDateElement).toHaveClass('text-red-600');
  });

  test('displays completion indicator for completed todos', () => {
    const completedTodo = { ...mockTodo, completed: true };
    render(<TodoItem todo={completedTodo} onEdit={mockOnEdit} />);
    
    expect(screen.getByText(/Completed on/)).toBeInTheDocument();
  });
});