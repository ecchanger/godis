/**
 * Utility functions for the Godis Todo application
 */

// DOM utility functions
const $ = (selector, parent = document) => parent.querySelector(selector);
const $$ = (selector, parent = document) => Array.from(parent.querySelectorAll(selector));

// Event delegation helper
const on = (parent, eventType, selector, handler) => {
  parent.addEventListener(eventType, (event) => {
    if (event.target.matches(selector)) {
      handler(event);
    }
  });
};

// Debounce function for search and other frequent operations
const debounce = (func, wait) => {
  let timeout;
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout);
      func(...args);
    };
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
};

// Throttle function for scroll events
const throttle = (func, limit) => {
  let inThrottle;
  return function () {
    const args = arguments;
    const context = this;
    if (!inThrottle) {
      func.apply(context, args);
      inThrottle = true;
      setTimeout(() => (inThrottle = false), limit);
    }
  };
};

// Date formatting utilities
const formatDate = (dateString) => {
  if (!dateString) return '';
  
  const date = new Date(dateString);
  const now = new Date();
  const diff = now - date;
  const seconds = Math.floor(diff / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 7) {
    return date.toLocaleDateString();
  } else if (days > 0) {
    return `${days} day${days > 1 ? 's' : ''} ago`;
  } else if (hours > 0) {
    return `${hours} hour${hours > 1 ? 's' : ''} ago`;
  } else if (minutes > 0) {
    return `${minutes} minute${minutes > 1 ? 's' : ''} ago`;
  } else {
    return 'Just now';
  }
};

const formatDueDate = (dueDateString) => {
  if (!dueDateString) return null;
  
  const dueDate = new Date(dueDateString);
  const now = new Date();
  const diff = dueDate - now;
  const hours = Math.floor(diff / (1000 * 60 * 60));
  const days = Math.floor(hours / 24);

  let status = 'normal';
  let text = dueDate.toLocaleDateString();

  if (diff < 0) {
    // Overdue
    const overdueDays = Math.abs(days);
    status = 'overdue';
    text = `Overdue by ${overdueDays} day${overdueDays > 1 ? 's' : ''}`;
  } else if (hours < 24) {
    // Due today
    status = 'due-soon';
    text = `Due in ${hours} hour${hours > 1 ? 's' : ''}`;
  } else if (days <= 3) {
    // Due soon
    status = 'due-soon';
    text = `Due in ${days} day${days > 1 ? 's' : ''}`;
  }

  return { text, status };
};

// Local storage utilities
const storage = {
  set: (key, value) => {
    try {
      localStorage.setItem(key, JSON.stringify(value));
    } catch (error) {
      console.warn('Failed to save to localStorage:', error);
    }
  },
  
  get: (key, defaultValue = null) => {
    try {
      const item = localStorage.getItem(key);
      return item ? JSON.parse(item) : defaultValue;
    } catch (error) {
      console.warn('Failed to read from localStorage:', error);
      return defaultValue;
    }
  },
  
  remove: (key) => {
    try {
      localStorage.removeItem(key);
    } catch (error) {
      console.warn('Failed to remove from localStorage:', error);
    }
  }
};

// Theme management
const theme = {
  get: () => storage.get('theme', 'light'),
  
  set: (themeName) => {
    storage.set('theme', themeName);
    document.body.className = `theme-${themeName}`;
    
    // Update theme toggle icon
    const themeToggle = $('#themeToggle');
    if (themeToggle) {
      const icon = themeToggle.querySelector('i');
      if (icon) {
        icon.className = themeName === 'dark' ? 'fas fa-sun' : 'fas fa-moon';
      }
    }
  },
  
  toggle: () => {
    const current = theme.get();
    const next = current === 'light' ? 'dark' : 'light';
    theme.set(next);
    return next;
  },
  
  init: () => {
    const savedTheme = theme.get();
    theme.set(savedTheme);
  }
};

// Form validation utilities
const validation = {
  required: (value) => ({
    isValid: value && value.trim().length > 0,
    message: 'This field is required'
  }),
  
  maxLength: (value, max) => ({
    isValid: !value || value.length <= max,
    message: `Maximum ${max} characters allowed`
  }),
  
  minLength: (value, min) => ({
    isValid: !value || value.length >= min,
    message: `Minimum ${min} characters required`
  }),
  
  date: (value) => {
    if (!value) return { isValid: true, message: '' };
    const date = new Date(value);
    return {
      isValid: !isNaN(date.getTime()),
      message: 'Please enter a valid date'
    };
  },
  
  priority: (value) => {
    const validPriorities = ['low', 'medium', 'high'];
    return {
      isValid: !value || validPriorities.includes(value.toLowerCase()),
      message: 'Priority must be low, medium, or high'
    };
  }
};

// Form data extraction utility
const getFormData = (form) => {
  const formData = new FormData(form);
  const data = {};
  
  for (const [key, value] of formData.entries()) {
    // Handle checkboxes
    const input = form.querySelector(`[name="${key}"]`);
    if (input && input.type === 'checkbox') {
      data[key] = input.checked;
    } else {
      data[key] = value || undefined;
    }
  }
  
  // Remove empty strings and undefined values
  Object.keys(data).forEach(key => {
    if (data[key] === '' || data[key] === undefined) {
      delete data[key];
    }
  });
  
  return data;
};

// HTML escaping for XSS prevention
const escapeHtml = (text) => {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
};

// URL utilities
const url = {
  params: () => new URLSearchParams(window.location.search),
  
  setParam: (key, value) => {
    const params = url.params();
    params.set(key, value);
    history.replaceState(null, '', `${location.pathname}?${params}`);
  },
  
  removeParam: (key) => {
    const params = url.params();
    params.delete(key);
    const queryString = params.toString();
    history.replaceState(null, '', `${location.pathname}${queryString ? '?' + queryString : ''}`);
  }
};

// Animation utilities
const animate = {
  fadeIn: (element, duration = 300) => {
    element.style.opacity = '0';
    element.style.display = 'block';
    
    let start = null;
    const fade = (timestamp) => {
      if (!start) start = timestamp;
      const progress = timestamp - start;
      const opacity = Math.min(progress / duration, 1);
      
      element.style.opacity = opacity;
      
      if (progress < duration) {
        requestAnimationFrame(fade);
      }
    };
    
    requestAnimationFrame(fade);
  },
  
  fadeOut: (element, duration = 300) => {
    let start = null;
    const fade = (timestamp) => {
      if (!start) start = timestamp;
      const progress = timestamp - start;
      const opacity = Math.max(1 - progress / duration, 0);
      
      element.style.opacity = opacity;
      
      if (progress < duration) {
        requestAnimationFrame(fade);
      } else {
        element.style.display = 'none';
      }
    };
    
    requestAnimationFrame(fade);
  },
  
  slideUp: (element, duration = 300) => {
    element.style.height = element.offsetHeight + 'px';
    element.style.overflow = 'hidden';
    element.style.transition = `height ${duration}ms ease-out`;
    
    requestAnimationFrame(() => {
      element.style.height = '0px';
    });
    
    setTimeout(() => {
      element.style.display = 'none';
      element.style.height = '';
      element.style.overflow = '';
      element.style.transition = '';
    }, duration);
  }
};

// Error handling utilities
const handleError = (error, context = 'Unknown') => {
  console.error(`Error in ${context}:`, error);
  
  // Show user-friendly error message
  const message = error.message || 'An unexpected error occurred';
  showToast('error', 'Error', message);
};

// Copy to clipboard utility
const copyToClipboard = async (text) => {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text);
      return true;
    } else {
      // Fallback for older browsers
      const textArea = document.createElement('textarea');
      textArea.value = text;
      textArea.style.position = 'fixed';
      textArea.style.left = '-999999px';
      textArea.style.top = '-999999px';
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();
      const result = document.execCommand('copy');
      textArea.remove();
      return result;
    }
  } catch (error) {
    console.error('Failed to copy to clipboard:', error);
    return false;
  }
};

// Random ID generator
const generateId = () => {
  return Math.random().toString(36).substr(2, 9);
};

// Object deep clone utility
const deepClone = (obj) => {
  if (obj === null || typeof obj !== 'object') return obj;
  if (obj instanceof Date) return new Date(obj.getTime());
  if (obj instanceof Array) return obj.map(item => deepClone(item));
  if (typeof obj === 'object') {
    const clonedObj = {};
    for (const key in obj) {
      if (obj.hasOwnProperty(key)) {
        clonedObj[key] = deepClone(obj[key]);
      }
    }
    return clonedObj;
  }
};

// Focus management for accessibility
const focusManagement = {
  trap: (element) => {
    const focusableElements = element.querySelectorAll(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    );
    const firstElement = focusableElements[0];
    const lastElement = focusableElements[focusableElements.length - 1];
    
    const handleTab = (e) => {
      if (e.key === 'Tab') {
        if (e.shiftKey) {
          if (document.activeElement === firstElement) {
            lastElement.focus();
            e.preventDefault();
          }
        } else {
          if (document.activeElement === lastElement) {
            firstElement.focus();
            e.preventDefault();
          }
        }
      }
    };
    
    element.addEventListener('keydown', handleTab);
    firstElement?.focus();
    
    return () => element.removeEventListener('keydown', handleTab);
  }
};

// Export utilities for use in other modules
window.utils = {
  $,
  $$,
  on,
  debounce,
  throttle,
  formatDate,
  formatDueDate,
  storage,
  theme,
  validation,
  getFormData,
  escapeHtml,
  url,
  animate,
  handleError,
  copyToClipboard,
  generateId,
  deepClone,
  focusManagement
};