import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import '@testing-library/jest-dom';
import TodoForm from '../TodoForm';

// Mock the store
jest.mock('../../store/todoStore', () => ({
  useLoading: () => false,
}));

const mockOnSubmit = jest.fn();
const mockOnCancel = jest.fn();

describe('TodoForm', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders form fields correctly', () => {
    render(<TodoForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);
    
    expect(screen.getByLabelText(/title/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/description/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/priority/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/due date/i)).toBeInTheDocument();
  });

  test('shows validation error for empty title', async () => {
    const user = userEvent.setup();
    render(<TodoForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);
    
    const submitButton = screen.getByText(/create todo/i);
    await user.click(submitButton);
    
    expect(screen.getByText(/title cannot be empty/i)).toBeInTheDocument();
  });

  test('submits form with valid data', async () => {
    const user = userEvent.setup();
    render(<TodoForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);
    
    const titleInput = screen.getByLabelText(/title/i);
    const descriptionInput = screen.getByLabelText(/description/i);
    const submitButton = screen.getByText(/create todo/i);
    
    await user.type(titleInput, 'Test Todo');
    await user.type(descriptionInput, 'Test Description');
    await user.click(submitButton);
    
    expect(mockOnSubmit).toHaveBeenCalledWith({
      title: 'Test Todo',
      description: 'Test Description',
      priority: 'medium', // default value
    });
  });

  test('populates form when editing existing todo', () => {
    const existingTodo = {
      id: 'test-id',
      title: 'Existing Todo',
      description: 'Existing Description',
      priority: 'high',
      dueDate: '2023-12-31',
    };
    
    render(
      <TodoForm 
        todo={existingTodo} 
        onSubmit={mockOnSubmit} 
        onCancel={mockOnCancel} 
      />
    );
    
    expect(screen.getByDisplayValue('Existing Todo')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Existing Description')).toBeInTheDocument();
    expect(screen.getByDisplayValue('high')).toBeInTheDocument();
  });

  test('shows character count for title and description', async () => {
    const user = userEvent.setup();
    render(<TodoForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);
    
    const titleInput = screen.getByLabelText(/title/i);
    await user.type(titleInput, 'Test');
    
    expect(screen.getByText('4/200 characters')).toBeInTheDocument();
  });

  test('calls onCancel when cancel button is clicked', async () => {
    const user = userEvent.setup();
    render(<TodoForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);
    
    const cancelButton = screen.getByText(/cancel/i);
    await user.click(cancelButton);
    
    expect(mockOnCancel).toHaveBeenCalled();
  });

  test('validates due date is not in the past', async () => {
    const user = userEvent.setup();
    render(<TodoForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);
    
    const titleInput = screen.getByLabelText(/title/i);
    const dueDateInput = screen.getByLabelText(/due date/i);
    const submitButton = screen.getByText(/create todo/i);
    
    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);
    const pastDate = yesterday.toISOString().split('T')[0];
    
    await user.type(titleInput, 'Test Todo');
    await user.type(dueDateInput, pastDate);
    await user.click(submitButton);
    
    expect(screen.getByText(/due date cannot be in the past/i)).toBeInTheDocument();
  });

  test('shows priority preview', async () => {
    const user = userEvent.setup();
    render(<TodoForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);
    
    const prioritySelect = screen.getByLabelText(/priority/i);
    await user.selectOptions(prioritySelect, 'high');
    
    expect(screen.getByText(/high priority/i)).toBeInTheDocument();
  });
});