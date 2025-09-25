import React from 'react';
import { Plus, CheckCircle, Circle, AlertCircle } from 'lucide-react';
import { useStats } from '../store/todoStore';

const Header = ({ onAddTodo }) => {
  const stats = useStats();

  return (
    <div className="mb-8">
      {/* Main Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-6">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            Todo List
          </h1>
          <p className="text-gray-600">
            Manage your tasks efficiently with our modern todo application
          </p>
        </div>
        
        <button
          onClick={onAddTodo}
          className="btn-primary flex items-center gap-2 mt-4 sm:mt-0"
        >
          <Plus size={20} />
          Add Todo
        </button>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {/* Total Tasks */}
          <div className="card flex items-center gap-3 p-4">
            <div className="p-2 bg-blue-100 rounded-lg">
              <Circle className="text-blue-600" size={24} />
            </div>
            <div>
              <p className="text-2xl font-semibold text-gray-900">
                {stats.total}
              </p>
              <p className="text-sm text-gray-600">Total Tasks</p>
            </div>
          </div>

          {/* Completed Tasks */}
          <div className="card flex items-center gap-3 p-4">
            <div className="p-2 bg-green-100 rounded-lg">
              <CheckCircle className="text-green-600" size={24} />
            </div>
            <div>
              <p className="text-2xl font-semibold text-gray-900">
                {stats.completed}
              </p>
              <p className="text-sm text-gray-600">Completed</p>
            </div>
          </div>

          {/* Pending Tasks */}
          <div className="card flex items-center gap-3 p-4">
            <div className="p-2 bg-yellow-100 rounded-lg">
              <Circle className="text-yellow-600" size={24} />
            </div>
            <div>
              <p className="text-2xl font-semibold text-gray-900">
                {stats.pending}
              </p>
              <p className="text-sm text-gray-600">Pending</p>
            </div>
          </div>

          {/* High Priority Tasks */}
          <div className="card flex items-center gap-3 p-4">
            <div className="p-2 bg-red-100 rounded-lg">
              <AlertCircle className="text-red-600" size={24} />
            </div>
            <div>
              <p className="text-2xl font-semibold text-gray-900">
                {stats.highPriority}
              </p>
              <p className="text-sm text-gray-600">High Priority</p>
            </div>
          </div>
        </div>
      )}

      {/* Progress Bar */}
      {stats && stats.total > 0 && (
        <div className="mt-4">
          <div className="flex justify-between text-sm text-gray-600 mb-1">
            <span>Progress</span>
            <span>{Math.round((stats.completed / stats.total) * 100)}%</span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div
              className="bg-green-500 h-2 rounded-full transition-all duration-300"
              style={{
                width: `${(stats.completed / stats.total) * 100}%`,
              }}
            />
          </div>
        </div>
      )}
    </div>
  );
};

export default Header;