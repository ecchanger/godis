/**
 * UI Components for the Godis Todo application
 * Handles modals, toasts, todo items, and other interactive components
 */

/**
 * Toast notification system
 */
class ToastManager {
  constructor() {
    this.container = utils.$('#toastContainer');
    this.toasts = new Map();
    this.defaultDuration = 5000;
  }

  show(type, title, message, duration = this.defaultDuration) {
    const id = utils.generateId();
    const toast = this.createToast(id, type, title, message, duration);
    
    this.container.appendChild(toast);
    this.toasts.set(id, toast);
    
    // Auto-remove after duration
    if (duration > 0) {
      setTimeout(() => this.remove(id), duration);
    }
    
    return id;
  }

  createToast(id, type, title, message, duration) {
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.dataset.toastId = id;
    
    const icons = {
      success: 'fas fa-check',
      error: 'fas fa-exclamation-triangle',
      warning: 'fas fa-exclamation-circle',
      info: 'fas fa-info-circle'
    };
    
    toast.innerHTML = `
      <div class="toast-icon">
        <i class="${icons[type] || icons.info}"></i>
      </div>
      <div class="toast-content">
        <div class="toast-title">${utils.escapeHtml(title)}</div>
        ${message ? `<div class="toast-message">${utils.escapeHtml(message)}</div>` : ''}
      </div>
      <button class="toast-close" type="button">
        <i class="fas fa-times"></i>
      </button>
      ${duration > 0 ? '<div class="toast-progress"></div>' : ''}
    `;
    
    // Add click handler for close button
    const closeBtn = toast.querySelector('.toast-close');
    closeBtn.addEventListener('click', () => this.remove(id));
    
    // Add progress bar animation
    if (duration > 0) {
      const progress = toast.querySelector('.toast-progress');
      if (progress) {
        progress.style.width = '100%';
        setTimeout(() => {
          progress.style.width = '0%';
          progress.style.transition = `width ${duration}ms linear`;
        }, 50);
      }
    }
    
    return toast;
  }

  remove(id) {
    const toast = this.toasts.get(id);
    if (toast) {
      toast.classList.add('toast-exit');
      setTimeout(() => {
        if (toast.parentNode) {
          toast.parentNode.removeChild(toast);
        }
        this.toasts.delete(id);
      }, 300);
    }
  }

  clear() {
    for (const id of this.toasts.keys()) {
      this.remove(id);
    }
  }
}

/**
 * Modal management system
 */
class ModalManager {
  constructor() {
    this.activeModal = null;
    this.focusTrap = null;
    this.beforeOpenCallback = null;
    this.afterCloseCallback = null;
    
    this.initEventListeners();
  }

  initEventListeners() {
    // Close modal on escape key
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && this.activeModal) {
        this.close();
      }
    });

    // Close modal on overlay click
    document.addEventListener('click', (e) => {
      if (e.target.classList.contains('modal-overlay') && this.activeModal) {
        this.close();
      }
    });
  }

  open(modalId, data = {}) {
    const modal = utils.$(`#${modalId}`);
    if (!modal) {
      console.error(`Modal with id "${modalId}" not found`);
      return;
    }

    // Call before open callback
    if (this.beforeOpenCallback) {
      this.beforeOpenCallback(modalId, data);
    }

    // Populate modal with data if provided
    this.populateModal(modal, data);

    // Show modal
    modal.classList.add('active');
    this.activeModal = modal;

    // Set up focus trap
    this.focusTrap = utils.focusManagement.trap(modal);

    // Prevent body scroll
    document.body.style.overflow = 'hidden';
  }

  close() {
    if (!this.activeModal) return;

    const modal = this.activeModal;
    
    // Remove focus trap
    if (this.focusTrap) {
      this.focusTrap();
      this.focusTrap = null;
    }

    // Hide modal
    modal.classList.remove('active');
    
    // Restore body scroll
    document.body.style.overflow = '';

    // Call after close callback
    if (this.afterCloseCallback) {
      this.afterCloseCallback(modal.id);
    }

    this.activeModal = null;
  }

  populateModal(modal, data) {
    // Generic modal population - override for specific modals
    Object.keys(data).forEach(key => {
      const element = modal.querySelector(`[name="${key}"], #${key}`);
      if (element) {
        if (element.type === 'checkbox') {
          element.checked = data[key];
        } else {
          element.value = data[key] || '';
        }
      }
    });
  }

  onBeforeOpen(callback) {
    this.beforeOpenCallback = callback;
  }

  onAfterClose(callback) {
    this.afterCloseCallback = callback;
  }
}

/**
 * Todo item component
 */
class TodoItemComponent {
  constructor(todo, container) {
    this.todo = todo;
    this.container = container;
    this.element = null;
    
    this.render();
    this.bindEvents();
  }

  render() {
    const dueDate = this.todo.due_date ? utils.formatDueDate(this.todo.due_date) : null;
    
    this.element = document.createElement('div');
    this.element.className = `todo-item ${this.todo.completed ? 'completed' : ''}`;
    this.element.dataset.todoId = this.todo.id;
    
    this.element.innerHTML = `
      <label class="todo-checkbox checkbox-label">
        <input type="checkbox" class="checkbox" ${this.todo.completed ? 'checked' : ''}>
        <span class="checkbox-custom"></span>
      </label>
      
      <div class="todo-content">
        <h3 class="todo-title">${utils.escapeHtml(this.todo.title)}</h3>
        ${this.todo.description ? `<p class="todo-description">${utils.escapeHtml(this.todo.description)}</p>` : ''}
        
        <div class="todo-meta">
          <span class="todo-priority ${this.todo.priority}">
            <i class="fas fa-flag"></i>
            ${this.todo.priority}
          </span>
          
          <span class="todo-date">
            Created ${utils.formatDate(this.todo.created_at)}
          </span>
          
          ${dueDate ? `
            <span class="todo-due-date ${dueDate.status}">
              <i class="fas fa-calendar-alt"></i>
              ${dueDate.text}
            </span>
          ` : ''}
        </div>
      </div>
      
      <div class="todo-actions">
        <button class="todo-action-btn edit" type="button" title="Edit task">
          <i class="fas fa-edit"></i>
        </button>
        <button class="todo-action-btn delete" type="button" title="Delete task">
          <i class="fas fa-trash"></i>
        </button>
      </div>
    `;
    
    if (this.container) {
      this.container.appendChild(this.element);
    }
  }

  bindEvents() {
    if (!this.element) return;

    // Checkbox toggle
    const checkbox = this.element.querySelector('.checkbox');
    checkbox.addEventListener('change', (e) => {
      this.onToggle(e.target.checked);
    });

    // Edit button
    const editBtn = this.element.querySelector('.edit');
    editBtn.addEventListener('click', () => {
      this.onEdit();
    });

    // Delete button
    const deleteBtn = this.element.querySelector('.delete');
    deleteBtn.addEventListener('click', () => {
      this.onDelete();
    });
  }

  onToggle(completed) {
    // Emit custom event
    this.element.dispatchEvent(new CustomEvent('todo:toggle', {
      detail: { todo: this.todo, completed },
      bubbles: true
    }));
  }

  onEdit() {
    this.element.dispatchEvent(new CustomEvent('todo:edit', {
      detail: { todo: this.todo },
      bubbles: true
    }));
  }

  onDelete() {
    this.element.dispatchEvent(new CustomEvent('todo:delete', {
      detail: { todo: this.todo },
      bubbles: true
    }));
  }

  update(newTodo) {
    this.todo = { ...this.todo, ...newTodo };
    
    // Re-render the component
    const parent = this.element.parentNode;
    const nextSibling = this.element.nextSibling;
    
    this.element.remove();
    this.render();
    this.bindEvents();
    
    if (parent) {
      parent.insertBefore(this.element, nextSibling);
    }
  }

  remove() {
    if (this.element && this.element.parentNode) {
      utils.animate.slideUp(this.element, 300);
      setTimeout(() => {
        if (this.element && this.element.parentNode) {
          this.element.parentNode.removeChild(this.element);
        }
      }, 300);
    }
  }
}

/**
 * Loading states manager
 */
class LoadingManager {
  constructor() {
    this.loadingElements = new Set();
  }

  show(element, type = 'spinner') {
    if (typeof element === 'string') {
      element = utils.$(element);
    }
    
    if (!element) return;

    element.classList.add('loading');
    this.loadingElements.add(element);

    if (type === 'skeleton') {
      this.showSkeleton(element);
    } else {
      this.showSpinner(element);
    }
  }

  hide(element) {
    if (typeof element === 'string') {
      element = utils.$(element);
    }
    
    if (!element) return;

    element.classList.remove('loading');
    this.loadingElements.delete(element);

    // Remove loading content
    const loadingContent = element.querySelector('.loading-content');
    if (loadingContent) {
      loadingContent.remove();
    }
  }

  showSpinner(element) {
    const spinner = document.createElement('div');
    spinner.className = 'loading-content loading-spinner';
    spinner.innerHTML = '<i class="fas fa-spinner fa-spin"></i>';
    element.appendChild(spinner);
  }

  showSkeleton(element) {
    const skeleton = document.createElement('div');
    skeleton.className = 'loading-content todo-skeleton';
    skeleton.innerHTML = `
      <div class="skeleton-line medium"></div>
      <div class="skeleton-line long"></div>
      <div class="skeleton-line short"></div>
    `;
    element.appendChild(skeleton);
  }

  hideAll() {
    for (const element of this.loadingElements) {
      this.hide(element);
    }
  }
}

/**
 * Search and filter component
 */
class SearchFilter {
  constructor(onFilter) {
    this.onFilter = onFilter;
    this.currentFilters = {
      status: 'all',
      priority: 'all',
      search: ''
    };
    
    this.initElements();
    this.bindEvents();
  }

  initElements() {
    this.searchInput = utils.$('#searchInput');
    this.statusButtons = utils.$$('[data-filter]');
    this.priorityButtons = utils.$$('[data-priority]');
  }

  bindEvents() {
    // Search input with debounce
    if (this.searchInput) {
      this.searchInput.addEventListener('input', utils.debounce((e) => {
        this.updateFilter('search', e.target.value);
      }, 300));
    }

    // Filter buttons
    this.statusButtons.forEach(button => {
      button.addEventListener('click', () => {
        this.setActiveButton(this.statusButtons, button);
        this.updateFilter('status', button.dataset.filter);
      });
    });

    this.priorityButtons.forEach(button => {
      button.addEventListener('click', () => {
        this.setActiveButton(this.priorityButtons, button);
        this.updateFilter('priority', button.dataset.priority);
      });
    });
  }

  setActiveButton(buttons, activeButton) {
    buttons.forEach(btn => btn.classList.remove('active'));
    activeButton.classList.add('active');
  }

  updateFilter(key, value) {
    this.currentFilters[key] = value;
    this.onFilter(this.currentFilters);
  }

  getFilters() {
    return { ...this.currentFilters };
  }

  reset() {
    this.currentFilters = {
      status: 'all',
      priority: 'all',
      search: ''
    };

    // Reset UI
    if (this.searchInput) {
      this.searchInput.value = '';
    }

    // Reset active buttons
    this.statusButtons.forEach((btn, index) => {
      btn.classList.toggle('active', index === 0);
    });

    this.priorityButtons.forEach((btn, index) => {
      btn.classList.toggle('active', index === 0);
    });

    this.onFilter(this.currentFilters);
  }
}

/**
 * Stats component
 */
class StatsComponent {
  constructor() {
    this.elements = {
      total: utils.$('#totalCount'),
      completed: utils.$('#completedCount'),
      pending: utils.$('#pendingCount'),
      highPriority: utils.$('#highPriorityCount')
    };
  }

  update(todos) {
    const stats = this.calculateStats(todos);
    
    Object.keys(stats).forEach(key => {
      if (this.elements[key]) {
        this.animateNumber(this.elements[key], stats[key]);
      }
    });
  }

  calculateStats(todos) {
    return {
      total: todos.length,
      completed: todos.filter(t => t.completed).length,
      pending: todos.filter(t => !t.completed).length,
      highPriority: todos.filter(t => t.priority === 'high' && !t.completed).length
    };
  }

  animateNumber(element, newValue) {
    const currentValue = parseInt(element.textContent) || 0;
    const duration = 500;
    const startTime = Date.now();

    const animate = () => {
      const elapsed = Date.now() - startTime;
      const progress = Math.min(elapsed / duration, 1);
      
      const value = Math.round(currentValue + (newValue - currentValue) * progress);
      element.textContent = value;

      if (progress < 1) {
        requestAnimationFrame(animate);
      }
    };

    requestAnimationFrame(animate);
  }
}

// Initialize and export components
const toastManager = new ToastManager();
const modalManager = new ModalManager();
const loadingManager = new LoadingManager();

// Global toast function for easy access
window.showToast = (type, title, message, duration) => {
  return toastManager.show(type, title, message, duration);
};

// Export components
window.components = {
  ToastManager,
  ModalManager,
  TodoItemComponent,
  LoadingManager,
  SearchFilter,
  StatsComponent,
  toastManager,
  modalManager,
  loadingManager
};