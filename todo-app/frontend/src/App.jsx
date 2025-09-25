import React, { useEffect, useState } from 'react';
import Header from './components/Header';
import TodoList from './components/TodoList';
import TodoForm from './components/TodoForm';
import Filter from './components/Filter';
import Modal from './components/Modal';
import ErrorBoundary from './components/ErrorBoundary';
import { useTodoStore, useTodoActions, useError } from './store/todoStore';

function App() {
  const [isFormModalOpen, setIsFormModalOpen] = useState(false);
  const [editingTodo, setEditingTodo] = useState(null);
  
  const error = useError();
  const { fetchTodos, fetchStats, clearError } = useTodoActions();

  // Initialize data on app load
  useEffect(() => {
    fetchTodos();
    fetchStats();
  }, [fetchTodos, fetchStats]);

  // Handle opening the todo form modal
  const handleAddTodo = () => {
    setEditingTodo(null);
    setIsFormModalOpen(true);
  };

  // Handle editing a todo
  const handleEditTodo = (todo) => {
    setEditingTodo(todo);
    setIsFormModalOpen(true);
  };

  // Handle closing the form modal
  const handleCloseModal = () => {
    setIsFormModalOpen(false);
    setEditingTodo(null);
  };

  // Handle form submission
  const handleFormSubmit = async (todoData) => {
    const { updateTodo, createTodo } = useTodoActions();
    try {
      if (editingTodo) {
        // Update existing todo
        await updateTodo(editingTodo.id, todoData);
      } else {
        // Create new todo
        await createTodo(todoData);
      }
      handleCloseModal();
    } catch (error) {
      // Error is handled in the store, form will show error state
      console.error('Failed to save todo:', error);
    }
  };

  return (
    <ErrorBoundary>
      <div className="min-h-screen bg-gray-50">
        {/* Global Error Display */}
        {error && (
          <div className="bg-red-500 text-white px-4 py-2 text-center relative">
            <span>{error}</span>
            <button
              onClick={clearError}
              className="absolute right-4 top-1/2 transform -translate-y-1/2 text-white hover:text-red-200"
            >
              ×
            </button>
          </div>
        )}

        {/* Main Layout */}
        <div className="container mx-auto px-4 py-8 max-w-4xl">
          {/* Header */}
          <Header onAddTodo={handleAddTodo} />

          {/* Filter Controls */}
          <div className="mb-6">
            <Filter />
          </div>

          {/* Todo List */}
          <TodoList onEditTodo={handleEditTodo} />

          {/* Todo Form Modal */}
          <Modal
            isOpen={isFormModalOpen}
            onClose={handleCloseModal}
            title={editingTodo ? 'Edit Todo' : 'Add New Todo'}
          >
            <TodoForm
              todo={editingTodo}
              onSubmit={handleFormSubmit}
              onCancel={handleCloseModal}
            />
          </Modal>
        </div>
      </div>
    </ErrorBoundary>
  );
}

export default App;