import React from 'react';
import { Loader2, AlertCircle, CheckCircle2 } from 'lucide-react';
import TodoItem from './TodoItem';
import { useFilteredTodos, useLoading, useError } from '../store/todoStore';

const TodoList = ({ onEditTodo }) => {
  const todos = useFilteredTodos();
  const loading = useLoading();
  const error = useError();

  // Loading state
  if (loading && todos.length === 0) {
    return (
      <div className="card">
        <div className="flex items-center justify-center py-12">
          <div className="text-center">
            <Loader2 className="animate-spin mx-auto mb-4 text-blue-500" size={48} />
            <p className="text-gray-600">Loading todos...</p>
          </div>
        </div>
      </div>
    );
  }

  // Error state (when no todos are loaded)
  if (error && todos.length === 0) {
    return (
      <div className="card">
        <div className="flex items-center justify-center py-12">
          <div className="text-center">
            <AlertCircle className="mx-auto mb-4 text-red-500" size={48} />
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
              Failed to load todos
            </h3>
            <p className="text-gray-600 mb-4">{error}</p>
            <button
              onClick={() => window.location.reload()}
              className="btn-primary"
            >
              Try Again
            </button>
          </div>
        </div>
      </div>
    );
  }

  // Empty state
  if (todos.length === 0) {
    return (
      <div className="card">
        <div className="flex items-center justify-center py-12">
          <div className="text-center">
            <CheckCircle2 className="mx-auto mb-4 text-gray-400" size={48} />
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
              No todos found
            </h3>
            <p className="text-gray-600 mb-4">
              Start by creating your first todo item, or adjust your filters.
            </p>
          </div>
        </div>
      </div>
    );
  }

  // Group todos by completion status for better UX
  const completedTodos = todos.filter(todo => todo.completed);
  const pendingTodos = todos.filter(todo => !todo.completed);

  return (
    <div className="space-y-6">
      {/* Pending Todos */}
      {pendingTodos.length > 0 && (
        <div>
          <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
            <div className="w-3 h-3 bg-yellow-500 rounded-full"></div>
            Pending Tasks ({pendingTodos.length})
          </h2>
          <div className="space-y-3">
            {pendingTodos.map((todo) => (
              <TodoItem
                key={todo.id}
                todo={todo}
                onEdit={onEditTodo}
              />
            ))}
          </div>
        </div>
      )}

      {/* Completed Todos */}
      {completedTodos.length > 0 && (
        <div>
          <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
            <div className="w-3 h-3 bg-green-500 rounded-full"></div>
            Completed Tasks ({completedTodos.length})
          </h2>
          <div className="space-y-3">
            {completedTodos.map((todo) => (
              <TodoItem
                key={todo.id}
                todo={todo}
                onEdit={onEditTodo}
              />
            ))}
          </div>
        </div>
      )}

      {/* Loading indicator for background operations */}
      {loading && todos.length > 0 && (
        <div className="flex items-center justify-center py-4">
          <Loader2 className="animate-spin text-blue-500" size={24} />
          <span className="ml-2 text-gray-600">Updating...</span>
        </div>
      )}
    </div>
  );
};

export default TodoList;