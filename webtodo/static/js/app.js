/**
 * Main application file for Godis Todo
 * Coordinates all components and handles the application state
 */

class TodoApp {
  constructor() {
    this.todos = [];
    this.filteredTodos = [];
    this.currentTodo = null;
    this.isLoading = false;
    this.hasMore = true;
    this.offset = 0;
    this.limit = 20;

    // Initialize components
    this.stats = new components.StatsComponent();
    this.searchFilter = new components.SearchFilter(this.handleFilter.bind(this));
    
    // Initialize elements
    this.initElements();
    
    // Bind events
    this.bindEvents();
    
    // Initialize theme
    utils.theme.init();
    
    // Load initial data
    this.loadTodos();
  }

  initElements() {
    this.todoForm = utils.$('#todoForm');
    this.editTodoForm = utils.$('#editTodoForm');
    this.todoList = utils.$('#todoList');
    this.loadingState = utils.$('#loadingState');
    this.emptyState = utils.$('#emptyState');
    this.loadMoreContainer = utils.$('#loadMoreContainer');
    this.loadMoreBtn = utils.$('#loadMoreBtn');
    this.themeToggle = utils.$('#themeToggle');
    this.refreshBtn = utils.$('#refreshBtn');
    this.clearCompletedBtn = utils.$('#clearCompletedBtn');
  }

  bindEvents() {
    // Form submissions
    if (this.todoForm) {
      this.todoForm.addEventListener('submit', this.handleCreateTodo.bind(this));
    }

    if (this.editTodoForm) {
      this.editTodoForm.addEventListener('submit', this.handleUpdateTodo.bind(this));
    }

    // Modal close events
    const closeButtons = utils.$$('.modal-close, #cancelEditBtn, #cancelDeleteBtn');
    closeButtons.forEach(btn => {
      btn.addEventListener('click', () => components.modalManager.close());
    });

    // Confirm delete
    const confirmDeleteBtn = utils.$('#confirmDeleteBtn');
    if (confirmDeleteBtn) {
      confirmDeleteBtn.addEventListener('click', this.handleConfirmDelete.bind(this));
    }

    // Header actions
    if (this.themeToggle) {
      this.themeToggle.addEventListener('click', () => {
        utils.theme.toggle();
      });
    }

    if (this.refreshBtn) {
      this.refreshBtn.addEventListener('click', () => {
        this.refreshTodos();
      });
    }

    if (this.clearCompletedBtn) {
      this.clearCompletedBtn.addEventListener('click', this.handleClearCompleted.bind(this));
    }

    // Load more
    if (this.loadMoreBtn) {
      this.loadMoreBtn.addEventListener('click', this.handleLoadMore.bind(this));
    }

    // Todo item events (using event delegation)
    if (this.todoList) {
      this.todoList.addEventListener('todo:toggle', this.handleToggleTodo.bind(this));
      this.todoList.addEventListener('todo:edit', this.handleEditTodo.bind(this));
      this.todoList.addEventListener('todo:delete', this.handleDeleteTodo.bind(this));
    }

    // Modal setup
    components.modalManager.onBeforeOpen((modalId, data) => {
      if (modalId === 'editModal') {
        this.populateEditModal(data);
      } else if (modalId === 'deleteModal') {
        this.populateDeleteModal(data);
      }
    });

    // Keyboard shortcuts
    document.addEventListener('keydown', this.handleKeyboard.bind(this));
  }

  // Data management methods
  async loadTodos(reset = false) {
    if (this.isLoading) return;

    this.setLoading(true);
    
    if (reset) {
      this.offset = 0;
      this.hasMore = true;
    }

    try {
      const filters = this.searchFilter.getFilters();
      const params = {
        limit: this.limit,
        offset: this.offset
      };

      // Add active filters
      if (filters.status !== 'all') {
        params.completed = filters.status === 'completed';
      }
      if (filters.priority !== 'all') {
        params.priority = filters.priority;
      }

      const response = await api.getTodos(params);
      const newTodos = response.data || [];

      if (reset) {
        this.todos = newTodos;
      } else {
        this.todos = [...this.todos, ...newTodos];
      }

      this.hasMore = newTodos.length === this.limit;
      this.offset += newTodos.length;

      this.applyClientSideFilters();
      this.renderTodos();
      this.updateStats();

    } catch (error) {
      utils.handleError(error, 'Loading todos');
      showToast('error', 'Error', 'Failed to load todos');
    } finally {
      this.setLoading(false);
    }
  }

  async refreshTodos() {
    // Add visual feedback
    if (this.refreshBtn) {
      this.refreshBtn.querySelector('i').classList.add('fa-spin');
    }

    try {
      // Clear cache to force fresh data
      api.cache.clear();
      await this.loadTodos(true);
      showToast('success', 'Success', 'Todos refreshed');
    } finally {
      if (this.refreshBtn) {
        setTimeout(() => {
          this.refreshBtn.querySelector('i').classList.remove('fa-spin');
        }, 1000);
      }
    }
  }

  applyClientSideFilters() {
    const filters = this.searchFilter.getFilters();
    
    this.filteredTodos = this.todos.filter(todo => {
      // Status filter
      if (filters.status === 'completed' && !todo.completed) return false;
      if (filters.status === 'pending' && todo.completed) return false;

      // Priority filter
      if (filters.priority !== 'all' && todo.priority !== filters.priority) return false;

      // Search filter
      if (filters.search) {
        const searchTerm = filters.search.toLowerCase();
        const titleMatch = todo.title.toLowerCase().includes(searchTerm);
        const descMatch = todo.description && todo.description.toLowerCase().includes(searchTerm);
        if (!titleMatch && !descMatch) return false;
      }

      return true;
    });
  }

  // Rendering methods
  renderTodos() {
    if (!this.todoList) return;

    // Show/hide states
    const hasAnyTodos = this.todos.length > 0;
    const hasFilteredTodos = this.filteredTodos.length > 0;

    this.loadingState.style.display = this.isLoading && !hasAnyTodos ? 'block' : 'none';
    this.emptyState.style.display = !this.isLoading && !hasFilteredTodos ? 'block' : 'none';
    this.todoList.style.display = hasFilteredTodos ? 'block' : 'none';

    // Clear and render todos
    if (hasFilteredTodos) {
      this.todoList.innerHTML = '';
      
      this.filteredTodos.forEach(todo => {
        new components.TodoItemComponent(todo, this.todoList);
      });
    }

    // Show/hide load more button
    if (this.loadMoreContainer) {
      this.loadMoreContainer.style.display = 
        this.hasMore && hasFilteredTodos && !this.isLoading ? 'block' : 'none';
    }
  }

  updateStats() {
    this.stats.update(this.todos);
  }

  setLoading(loading) {
    this.isLoading = loading;
    
    if (this.loadMoreBtn) {
      this.loadMoreBtn.disabled = loading;
      const icon = this.loadMoreBtn.querySelector('i');
      if (icon) {
        icon.className = loading ? 'fas fa-spinner fa-spin' : 'fas fa-chevron-down';
      }
    }
  }

  // Event handlers
  async handleCreateTodo(event) {
    event.preventDefault();
    
    const formData = utils.getFormData(this.todoForm);
    
    // Validate form
    const validation = this.validateTodoForm(formData);
    if (!validation.isValid) {
      showToast('error', 'Validation Error', validation.message);
      return;
    }

    try {
      const response = await api.createTodo(formData);
      const newTodo = response.data;
      
      // Add to local state
      this.todos.unshift(newTodo);
      this.applyClientSideFilters();
      this.renderTodos();
      this.updateStats();
      
      // Reset form
      this.todoForm.reset();
      
      showToast('success', 'Success', 'Todo created successfully');
      
    } catch (error) {
      utils.handleError(error, 'Creating todo');
      showToast('error', 'Error', error.userMessage || 'Failed to create todo');
    }
  }

  async handleUpdateTodo(event) {
    event.preventDefault();
    
    if (!this.currentTodo) return;

    const formData = utils.getFormData(this.editTodoForm);
    
    try {
      const response = await api.updateTodo(this.currentTodo.id, formData);
      const updatedTodo = response.data;
      
      // Update local state
      const index = this.todos.findIndex(t => t.id === updatedTodo.id);
      if (index !== -1) {
        this.todos[index] = updatedTodo;
        this.applyClientSideFilters();
        this.renderTodos();
        this.updateStats();
      }
      
      components.modalManager.close();
      showToast('success', 'Success', 'Todo updated successfully');
      
    } catch (error) {
      utils.handleError(error, 'Updating todo');
      showToast('error', 'Error', error.userMessage || 'Failed to update todo');
    }
  }

  async handleToggleTodo(event) {
    const { todo, completed } = event.detail;
    
    try {
      const response = await api.toggleTodo(todo.id, completed);
      const updatedTodo = response.data;
      
      // Update local state
      const index = this.todos.findIndex(t => t.id === todo.id);
      if (index !== -1) {
        this.todos[index] = updatedTodo;
        this.applyClientSideFilters();
        this.renderTodos();
        this.updateStats();
      }
      
    } catch (error) {
      utils.handleError(error, 'Toggling todo');
      showToast('error', 'Error', error.userMessage || 'Failed to update todo');
      
      // Revert checkbox state
      const todoElement = utils.$(`[data-todo-id="${todo.id}"]`);
      if (todoElement) {
        const checkbox = todoElement.querySelector('.checkbox');
        if (checkbox) {
          checkbox.checked = todo.completed;
        }
      }
    }
  }

  handleEditTodo(event) {
    const { todo } = event.detail;
    this.currentTodo = todo;
    components.modalManager.open('editModal', todo);
  }

  handleDeleteTodo(event) {
    const { todo } = event.detail;
    this.currentTodo = todo;
    components.modalManager.open('deleteModal', todo);
  }

  async handleConfirmDelete() {
    if (!this.currentTodo) return;

    try {
      await api.deleteTodo(this.currentTodo.id);
      
      // Remove from local state
      this.todos = this.todos.filter(t => t.id !== this.currentTodo.id);
      this.applyClientSideFilters();
      this.renderTodos();
      this.updateStats();
      
      components.modalManager.close();
      showToast('success', 'Success', 'Todo deleted successfully');
      
    } catch (error) {
      utils.handleError(error, 'Deleting todo');
      showToast('error', 'Error', error.userMessage || 'Failed to delete todo');
    }
  }

  async handleClearCompleted() {
    const completedTodos = this.todos.filter(t => t.completed);
    
    if (completedTodos.length === 0) {
      showToast('info', 'Info', 'No completed todos to clear');
      return;
    }

    const confirmMessage = `Are you sure you want to delete ${completedTodos.length} completed todo${completedTodos.length > 1 ? 's' : ''}?`;
    
    if (!confirm(confirmMessage)) {
      return;
    }

    try {
      await api.clearCompleted();
      
      // Remove from local state
      this.todos = this.todos.filter(t => !t.completed);
      this.applyClientSideFilters();
      this.renderTodos();
      this.updateStats();
      
      showToast('success', 'Success', `${completedTodos.length} completed todo${completedTodos.length > 1 ? 's' : ''} deleted`);
      
    } catch (error) {
      utils.handleError(error, 'Clearing completed todos');
      showToast('error', 'Error', error.userMessage || 'Failed to clear completed todos');
    }
  }

  async handleLoadMore() {
    await this.loadTodos(false);
  }

  handleFilter(filters) {
    // Apply client-side filters immediately
    this.applyClientSideFilters();
    this.renderTodos();
    
    // For server-side filtering, we would reload with new parameters
    // this.loadTodos(true);
  }

  handleKeyboard(event) {
    // Ctrl/Cmd + Enter: Submit current form
    if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
      const activeForm = document.activeElement.closest('form');
      if (activeForm) {
        activeForm.dispatchEvent(new Event('submit', { cancelable: true }));
      }
    }
    
    // Ctrl/Cmd + R: Refresh todos
    if ((event.ctrlKey || event.metaKey) && event.key === 'r') {
      event.preventDefault();
      this.refreshTodos();
    }
    
    // Ctrl/Cmd + N: Focus new todo input
    if ((event.ctrlKey || event.metaKey) && event.key === 'n') {
      event.preventDefault();
      const titleInput = utils.$('#todoTitle');
      if (titleInput) {
        titleInput.focus();
      }
    }
  }

  // Modal population methods
  populateEditModal(todo) {
    if (!todo) return;

    // Populate form fields
    const fields = ['title', 'description', 'priority', 'completed'];
    fields.forEach(field => {
      const element = utils.$(`#editTodo${field.charAt(0).toUpperCase() + field.slice(1)}`);
      if (element) {
        if (element.type === 'checkbox') {
          element.checked = todo[field] || false;
        } else {
          element.value = todo[field] || '';
        }
      }
    });

    // Handle due date
    const dueDateElement = utils.$('#editTodoDueDate');
    if (dueDateElement && todo.due_date) {
      // Convert to local datetime format
      const date = new Date(todo.due_date);
      const localDate = new Date(date.getTime() - date.getTimezoneOffset() * 60000);
      dueDateElement.value = localDate.toISOString().slice(0, 16);
    }
  }

  populateDeleteModal(todo) {
    const preview = utils.$('#deleteTaskPreview');
    if (preview && todo) {
      preview.innerHTML = `
        <div class="preview-title">${utils.escapeHtml(todo.title)}</div>
        <div class="preview-meta">
          <span class="todo-priority ${todo.priority}">
            <i class="fas fa-flag"></i>
            ${todo.priority}
          </span>
          <span>Created ${utils.formatDate(todo.created_at)}</span>
        </div>
      `;
    }
  }

  // Validation methods
  validateTodoForm(data) {
    const validations = [
      utils.validation.required(data.title),
      utils.validation.maxLength(data.title, 255),
      utils.validation.maxLength(data.description, 1000),
      utils.validation.priority(data.priority),
      utils.validation.date(data.due_date)
    ];

    const failed = validations.find(v => !v.isValid);
    
    return {
      isValid: !failed,
      message: failed ? failed.message : ''
    };
  }
}

// Initialize application when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
  try {
    window.todoApp = new TodoApp();
    console.log('Godis Todo application initialized successfully');
  } catch (error) {
    console.error('Failed to initialize application:', error);
    showToast('error', 'Initialization Error', 'Failed to start the application');
  }
});

// Handle online/offline status
window.addEventListener('online', () => {
  showToast('success', 'Connection Restored', 'You are back online');
});

window.addEventListener('offline', () => {
  showToast('warning', 'Connection Lost', 'You are currently offline');
});

// Global error handler
window.addEventListener('error', (event) => {
  console.error('Global error:', event.error);
  showToast('error', 'Unexpected Error', 'Something went wrong');
});

// Unhandled promise rejection handler
window.addEventListener('unhandledrejection', (event) => {
  console.error('Unhandled promise rejection:', event.reason);
  showToast('error', 'Error', 'An unexpected error occurred');
  event.preventDefault();
});