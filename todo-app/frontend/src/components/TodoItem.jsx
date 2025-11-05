import React, { useState } from 'react';
import { 
  Check, 
  Edit3, 
  Trash2, 
  Calendar, 
  AlertCircle,
  Clock,
  User,
  Tag
} from 'lucide-react';
import { useTodoActions } from '../store/todoStore';

const TodoItem = ({ todo, onEdit }) => {
  const [isDeleting, setIsDeleting] = useState(false);
  const [isToggling, setIsToggling] = useState(false);
  
  const { toggleTodo, deleteTodo } = useTodoActions();

  // Handle toggle completion
  const handleToggle = async () => {
    setIsToggling(true);
    try {
      await toggleTodo(todo.id);
    } catch (error) {
      console.error('Failed to toggle todo:', error);
    } finally {
      setIsToggling(false);
    }
  };

  // Handle delete
  const handleDelete = async () => {
    if (window.confirm('Are you sure you want to delete this todo?')) {
      setIsDeleting(true);
      try {
        await deleteTodo(todo.id);
      } catch (error) {
        console.error('Failed to delete todo:', error);
        setIsDeleting(false);
      }
    }
  };

  // Get priority styling
  const getPriorityColor = (priority) => {
    switch (priority) {
      case 'high': return 'border-l-red-500 bg-red-50';
      case 'medium': return 'border-l-yellow-500 bg-yellow-50';
      case 'low': return 'border-l-green-500 bg-green-50';
      default: return 'border-l-gray-500 bg-gray-50';
    }
  };

  const getPriorityBadgeColor = (priority) => {
    switch (priority) {
      case 'high': return 'bg-red-100 text-red-800 border-red-200';
      case 'medium': return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'low': return 'bg-green-100 text-green-800 border-green-200';
      default: return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  // Format date
  const formatDate = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  // Check if due date is overdue
  const isOverdue = (dueDate) => {
    if (!dueDate) return false;
    return new Date(dueDate) < new Date() && !todo.completed;
  };

  return (
    <div
      className={`todo-item border-l-4 transition-all duration-200 ${
        todo.completed 
          ? 'completed opacity-75' 
          : getPriorityColor(todo.priority)
      } ${isDeleting ? 'opacity-50 scale-95' : ''}`}
    >
      <div className="flex items-start gap-4">
        {/* Checkbox */}
        <button
          onClick={handleToggle}
          disabled={isToggling}
          className={`flex-shrink-0 w-5 h-5 rounded border-2 flex items-center justify-center transition-all duration-200 ${
            todo.completed
              ? 'bg-green-500 border-green-500 text-white'
              : 'border-gray-300 hover:border-green-500'
          } ${isToggling ? 'animate-pulse' : ''}`}
        >
          {todo.completed && <Check size={12} />}
        </button>

        {/* Main content */}
        <div className="flex-grow min-w-0">
          {/* Title and Priority */}
          <div className="flex items-start justify-between gap-2 mb-2">
            <h3
              className={`font-medium text-gray-900 leading-tight ${
                todo.completed ? 'line-through text-gray-500' : ''
              }`}
            >
              {todo.title}
            </h3>
            
            <span
              className={`px-2 py-1 text-xs font-medium rounded-full border capitalize flex-shrink-0 ${
                getPriorityBadgeColor(todo.priority)
              }`}
            >
              {todo.priority}
            </span>
          </div>

          {/* Description */}
          {todo.description && (
            <p
              className={`text-sm mb-3 ${
                todo.completed ? 'text-gray-400' : 'text-gray-600'
              }`}
            >
              {todo.description}
            </p>
          )}

          {/* Tags */}
          {todo.tags && todo.tags.length > 0 && (
            <div className="flex flex-wrap gap-2 mb-3">
              {todo.tags.map((tag, index) => (
                <span
                  key={index}
                  className={`inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${
                    todo.completed 
                      ? 'bg-gray-200 text-gray-600' 
                      : 'bg-blue-100 text-blue-800'
                  }`}
                >
                  <Tag size={10} />
                  {tag}
                </span>
              ))}
            </div>
          )}

          {/* Meta information */}
          <div className="flex flex-wrap items-center gap-4 text-xs text-gray-500">
            {/* Created date */}
            <div className="flex items-center gap-1">
              <Clock size={12} />
              Created {formatDate(todo.createdAt)}
            </div>

            {/* Due date */}
            {todo.dueDate && (
              <div className={`flex items-center gap-1 ${
                isOverdue(todo.dueDate) ? 'text-red-600 font-medium' : ''
              }`}>
                <Calendar size={12} />
                Due {formatDate(todo.dueDate)}
                {isOverdue(todo.dueDate) && (
                  <AlertCircle size={12} className="text-red-600" />
                )}
              </div>
            )}

            {/* Updated date (if different from created) */}
            {todo.updatedAt !== todo.createdAt && (
              <div className="flex items-center gap-1">
                <User size={12} />
                Updated {formatDate(todo.updatedAt)}
              </div>
            )}
          </div>
        </div>

        {/* Action buttons */}
        <div className="flex items-center gap-2 flex-shrink-0">
          <button
            onClick={() => onEdit(todo)}
            className="p-2 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
            title="Edit todo"
          >
            <Edit3 size={16} />
          </button>
          
          <button
            onClick={handleDelete}
            disabled={isDeleting}
            className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors disabled:opacity-50"
            title="Delete todo"
          >
            <Trash2 size={16} />
          </button>
        </div>
      </div>

      {/* Completion indicator */}
      {todo.completed && (
        <div className="mt-3 pt-3 border-t border-gray-200">
          <div className="flex items-center gap-2 text-xs text-green-600">
            <Check size={12} />
            Completed on {formatDate(todo.updatedAt)}
          </div>
        </div>
      )}
    </div>
  );
};

export default TodoItem;