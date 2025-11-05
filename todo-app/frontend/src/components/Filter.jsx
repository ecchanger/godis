import React, { useMemo } from 'react';
import { Filter as FilterIcon, RotateCcw, ArrowUpDown, Tag } from 'lucide-react';
import { useFilter, useTodos, useTodoActions } from '../store/todoStore';

const Filter = () => {
  const filter = useFilter();
  const todos = useTodos();
  const { setFilter, clearFilter } = useTodoActions();

  // Extract all unique tags from todos
  const allTags = useMemo(() => {
    const tagSet = new Set();
    todos.forEach(todo => {
      if (todo.tags && todo.tags.length > 0) {
        todo.tags.forEach(tag => tagSet.add(tag));
      }
    });
    return Array.from(tagSet).sort();
  }, [todos]);

  const handleCompletionFilter = (completed) => {
    setFilter({ 
      completed: filter.completed === completed ? undefined : completed 
    });
  };

  const handlePriorityFilter = (priority) => {
    setFilter({ 
      priority: filter.priority === priority ? undefined : priority 
    });
  };

  const handleTagFilter = (tag) => {
    setFilter({ 
      tag: filter.tag === tag ? undefined : tag 
    });
  };

  const handleSortChange = (sortBy) => {
    const newOrder = filter.sortBy === sortBy && filter.order === 'asc' ? 'desc' : 'asc';
    setFilter({ sortBy, order: newOrder });
  };

  const getPriorityColor = (priority) => {
    switch (priority) {
      case 'high': return 'text-red-600 bg-red-100 border-red-200';
      case 'medium': return 'text-yellow-600 bg-yellow-100 border-yellow-200';
      case 'low': return 'text-green-600 bg-green-100 border-green-200';
      default: return 'text-gray-600 bg-gray-100 border-gray-200';
    }
  };

  const isFilterActive = filter.completed !== undefined || filter.priority !== undefined || filter.tag !== undefined;

  return (
    <div className="card">
      <div className="flex flex-col lg:flex-row lg:items-center gap-4">
        {/* Filter Label */}
        <div className="flex items-center gap-2 text-gray-700 font-medium">
          <FilterIcon size={20} />
          Filters & Sorting
        </div>

        {/* Completion Status Filter */}
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-600">Status:</span>
          <div className="flex gap-1">
            <button
              onClick={() => handleCompletionFilter(false)}
              className={`px-3 py-1 rounded-full text-sm font-medium transition-colors ${
                filter.completed === false
                  ? 'bg-yellow-500 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              Pending
            </button>
            <button
              onClick={() => handleCompletionFilter(true)}
              className={`px-3 py-1 rounded-full text-sm font-medium transition-colors ${
                filter.completed === true
                  ? 'bg-green-500 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              Completed
            </button>
          </div>
        </div>

        {/* Priority Filter */}
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-600">Priority:</span>
          <div className="flex gap-1">
            {['high', 'medium', 'low'].map((priority) => (
              <button
                key={priority}
                onClick={() => handlePriorityFilter(priority)}
                className={`px-3 py-1 rounded-full text-sm font-medium border transition-colors capitalize ${
                  filter.priority === priority
                    ? getPriorityColor(priority)
                    : 'bg-gray-100 text-gray-700 border-gray-200 hover:bg-gray-200'
                }`}
              >
                {priority}
              </button>
            ))}
          </div>
        </div>

        {/* Tag Filter */}
        {allTags.length > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-600">Tag:</span>
            <div className="flex flex-wrap gap-1 max-w-md">
              {allTags.map((tag) => (
                <button
                  key={tag}
                  onClick={() => handleTagFilter(tag)}
                  className={`inline-flex items-center gap-1 px-3 py-1 rounded-full text-sm font-medium border transition-colors ${
                    filter.tag === tag
                      ? 'bg-blue-500 text-white border-blue-500'
                      : 'bg-gray-100 text-gray-700 border-gray-200 hover:bg-gray-200'
                  }`}
                >
                  <Tag size={12} />
                  {tag}
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Sort Options */}
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-600">Sort by:</span>
          <div className="flex gap-1">
            {[
              { key: 'createdAt', label: 'Date' },
              { key: 'title', label: 'Title' },
              { key: 'priority', label: 'Priority' },
              { key: 'dueDate', label: 'Due Date' },
            ].map(({ key, label }) => (
              <button
                key={key}
                onClick={() => handleSortChange(key)}
                className={`flex items-center gap-1 px-3 py-1 rounded-full text-sm font-medium transition-colors ${
                  filter.sortBy === key
                    ? 'bg-blue-500 text-white'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                {label}
                {filter.sortBy === key && (
                  <ArrowUpDown 
                    size={12} 
                    className={`transform transition-transform ${
                      filter.order === 'desc' ? 'rotate-180' : ''
                    }`}
                  />
                )}
              </button>
            ))}
          </div>
        </div>

        {/* Clear Filters Button */}
        {isFilterActive && (
          <button
            onClick={clearFilter}
            className="flex items-center gap-2 px-3 py-1 text-sm text-gray-600 hover:text-gray-800 transition-colors"
          >
            <RotateCcw size={16} />
            Clear Filters
          </button>
        )}
      </div>

      {/* Active Filters Summary */}
      {isFilterActive && (
        <div className="mt-3 pt-3 border-t border-gray-200">
          <div className="flex flex-wrap gap-2">
            <span className="text-sm text-gray-600">Active filters:</span>
            {filter.completed !== undefined && (
              <span className={`px-2 py-1 rounded text-xs font-medium ${
                filter.completed 
                  ? 'bg-green-100 text-green-800' 
                  : 'bg-yellow-100 text-yellow-800'
              }`}>
                {filter.completed ? 'Completed' : 'Pending'}
              </span>
            )}
            {filter.priority && (
              <span className={`px-2 py-1 rounded text-xs font-medium capitalize ${
                getPriorityColor(filter.priority)
              }`}>
                {filter.priority} Priority
              </span>
            )}
            {filter.tag && (
              <span className="inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-medium bg-blue-100 text-blue-800">
                <Tag size={12} />
                {filter.tag}
              </span>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default Filter;