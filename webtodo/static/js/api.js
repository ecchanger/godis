/**
 * API client for the Godis Todo application
 * Handles all communication with the backend REST API
 */

class APIClient {
  constructor(baseURL = '/api') {
    this.baseURL = baseURL;
    this.defaultHeaders = {
      'Content-Type': 'application/json',
    };
  }

  /**
   * Generic HTTP request method
   */
  async request(endpoint, options = {}) {
    const url = `${this.baseURL}${endpoint}`;
    const config = {
      headers: { ...this.defaultHeaders, ...options.headers },
      ...options
    };

    try {
      const response = await fetch(url, config);
      const data = await response.json();

      if (!response.ok) {
        throw new APIError(data.error || 'Request failed', response.status, data);
      }

      return data;
    } catch (error) {
      if (error instanceof APIError) {
        throw error;
      }
      
      // Network or other errors
      console.error('API request failed:', error);
      throw new APIError(
        'Network error occurred. Please check your connection.',
        0,
        { originalError: error.message }
      );
    }
  }

  /**
   * GET request
   */
  async get(endpoint, params = {}) {
    const queryString = new URLSearchParams(params).toString();
    const url = queryString ? `${endpoint}?${queryString}` : endpoint;
    
    return this.request(url, {
      method: 'GET'
    });
  }

  /**
   * POST request
   */
  async post(endpoint, data = {}) {
    return this.request(endpoint, {
      method: 'POST',
      body: JSON.stringify(data)
    });
  }

  /**
   * PUT request
   */
  async put(endpoint, data = {}) {
    return this.request(endpoint, {
      method: 'PUT',
      body: JSON.stringify(data)
    });
  }

  /**
   * DELETE request
   */
  async delete(endpoint) {
    return this.request(endpoint, {
      method: 'DELETE'
    });
  }

  // Todo-specific API methods

  /**
   * Get all todos with optional filtering
   */
  async getTodos(filters = {}) {
    return this.get('/todos', filters);
  }

  /**
   * Get a specific todo by ID
   */
  async getTodo(id) {
    return this.get(`/todos/${id}`);
  }

  /**
   * Create a new todo
   */
  async createTodo(todoData) {
    return this.post('/todos', todoData);
  }

  /**
   * Update an existing todo
   */
  async updateTodo(id, todoData) {
    return this.put(`/todos/${id}`, todoData);
  }

  /**
   * Delete a todo
   */
  async deleteTodo(id) {
    return this.delete(`/todos/${id}`);
  }

  /**
   * Toggle todo completion status
   */
  async toggleTodo(id, completed) {
    return this.updateTodo(id, { completed });
  }

  /**
   * Batch operations
   */
  async batchUpdateTodos(operations) {
    // For now, handle batch operations sequentially
    // In a real implementation, this could be optimized with a batch API endpoint
    const results = [];
    
    for (const operation of operations) {
      try {
        let result;
        switch (operation.type) {
          case 'update':
            result = await this.updateTodo(operation.id, operation.data);
            break;
          case 'delete':
            result = await this.deleteTodo(operation.id);
            break;
          case 'toggle':
            result = await this.toggleTodo(operation.id, operation.completed);
            break;
          default:
            throw new Error(`Unknown operation type: ${operation.type}`);
        }
        results.push({ success: true, id: operation.id, result });
      } catch (error) {
        results.push({ success: false, id: operation.id, error: error.message });
      }
    }
    
    return results;
  }

  /**
   * Clear completed todos
   */
  async clearCompleted() {
    // Get all completed todos first
    const response = await this.getTodos({ completed: true });
    const completedTodos = response.data || [];
    
    // Delete them in batch
    const operations = completedTodos.map(todo => ({
      type: 'delete',
      id: todo.id
    }));
    
    return this.batchUpdateTodos(operations);
  }
}

/**
 * Custom API Error class
 */
class APIError extends Error {
  constructor(message, status, data = {}) {
    super(message);
    this.name = 'APIError';
    this.status = status;
    this.data = data;
  }

  get isNetworkError() {
    return this.status === 0;
  }

  get isClientError() {
    return this.status >= 400 && this.status < 500;
  }

  get isServerError() {
    return this.status >= 500;
  }

  get userMessage() {
    // Return user-friendly error messages
    switch (this.status) {
      case 0:
        return 'Unable to connect to the server. Please check your internet connection.';
      case 400:
        return this.data.message || 'Invalid request. Please check your input.';
      case 404:
        return 'The requested item was not found.';
      case 429:
        return 'Too many requests. Please wait a moment and try again.';
      case 500:
        return 'Server error. Please try again later.';
      default:
        return this.message || 'An unexpected error occurred.';
    }
  }
}

/**
 * Request cache for optimizing API calls
 */
class RequestCache {
  constructor(ttl = 30000) { // 30 seconds default TTL
    this.cache = new Map();
    this.ttl = ttl;
  }

  generateKey(method, endpoint, params = {}) {
    const paramsString = JSON.stringify(params);
    return `${method}:${endpoint}:${paramsString}`;
  }

  set(method, endpoint, params, data) {
    const key = this.generateKey(method, endpoint, params);
    this.cache.set(key, {
      data,
      timestamp: Date.now()
    });
  }

  get(method, endpoint, params) {
    const key = this.generateKey(method, endpoint, params);
    const cached = this.cache.get(key);
    
    if (!cached) return null;
    
    if (Date.now() - cached.timestamp > this.ttl) {
      this.cache.delete(key);
      return null;
    }
    
    return cached.data;
  }

  invalidate(pattern) {
    for (const key of this.cache.keys()) {
      if (key.includes(pattern)) {
        this.cache.delete(key);
      }
    }
  }

  clear() {
    this.cache.clear();
  }
}

/**
 * Enhanced API client with caching and retry logic
 */
class CachedAPIClient extends APIClient {
  constructor(baseURL = '/api', options = {}) {
    super(baseURL);
    this.cache = new RequestCache(options.cacheTTL);
    this.retryAttempts = options.retryAttempts || 3;
    this.retryDelay = options.retryDelay || 1000;
  }

  async requestWithRetry(endpoint, options = {}, attempt = 1) {
    try {
      return await this.request(endpoint, options);
    } catch (error) {
      if (attempt < this.retryAttempts && (error.isNetworkError || error.isServerError)) {
        console.warn(`Request failed, retrying (${attempt}/${this.retryAttempts})...`);
        await this.delay(this.retryDelay * attempt);
        return this.requestWithRetry(endpoint, options, attempt + 1);
      }
      throw error;
    }
  }

  async get(endpoint, params = {}) {
    // Check cache for GET requests
    const cached = this.cache.get('GET', endpoint, params);
    if (cached) {
      return cached;
    }

    const result = await super.get(endpoint, params);
    this.cache.set('GET', endpoint, params, result);
    return result;
  }

  async post(endpoint, data = {}) {
    const result = await this.requestWithRetry(endpoint, {
      method: 'POST',
      body: JSON.stringify(data)
    });
    
    // Invalidate related cache entries
    this.cache.invalidate(endpoint.split('/')[1]); // Invalidate by resource type
    return result;
  }

  async put(endpoint, data = {}) {
    const result = await this.requestWithRetry(endpoint, {
      method: 'PUT',
      body: JSON.stringify(data)
    });
    
    this.cache.invalidate(endpoint.split('/')[1]);
    return result;
  }

  async delete(endpoint) {
    const result = await this.requestWithRetry(endpoint, {
      method: 'DELETE'
    });
    
    this.cache.invalidate(endpoint.split('/')[1]);
    return result;
  }

  delay(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

/**
 * API client with offline support
 */
class OfflineAPIClient extends CachedAPIClient {
  constructor(baseURL = '/api', options = {}) {
    super(baseURL, options);
    this.isOnline = navigator.onLine;
    this.pendingRequests = new Map();
    
    // Listen for online/offline events
    window.addEventListener('online', () => {
      this.isOnline = true;
      this.processPendingRequests();
    });
    
    window.addEventListener('offline', () => {
      this.isOnline = false;
    });
  }

  async request(endpoint, options = {}) {
    if (!this.isOnline && options.method !== 'GET') {
      // Queue write operations for when we're back online
      return this.queueRequest(endpoint, options);
    }

    try {
      return await super.requestWithRetry(endpoint, options);
    } catch (error) {
      if (error.isNetworkError && options.method !== 'GET') {
        return this.queueRequest(endpoint, options);
      }
      throw error;
    }
  }

  queueRequest(endpoint, options) {
    const id = Date.now() + Math.random();
    const request = { id, endpoint, options, timestamp: Date.now() };
    
    this.pendingRequests.set(id, request);
    
    // Save to localStorage for persistence across page reloads
    this.savePendingRequests();
    
    // Return a promise that will resolve when the request is processed
    return new Promise((resolve, reject) => {
      request.resolve = resolve;
      request.reject = reject;
    });
  }

  async processPendingRequests() {
    const requests = Array.from(this.pendingRequests.values());
    
    for (const request of requests) {
      try {
        const result = await super.requestWithRetry(request.endpoint, request.options);
        if (request.resolve) {
          request.resolve(result);
        }
        this.pendingRequests.delete(request.id);
      } catch (error) {
        if (request.reject) {
          request.reject(error);
        }
        this.pendingRequests.delete(request.id);
      }
    }
    
    this.savePendingRequests();
  }

  savePendingRequests() {
    const requests = Array.from(this.pendingRequests.entries()).map(([id, request]) => ({
      id,
      endpoint: request.endpoint,
      options: request.options,
      timestamp: request.timestamp
    }));
    
    utils.storage.set('pendingRequests', requests);
  }

  loadPendingRequests() {
    const requests = utils.storage.get('pendingRequests', []);
    
    for (const request of requests) {
      this.pendingRequests.set(request.id, request);
    }
  }
}

// Create and export the API client instance
const apiClient = new CachedAPIClient('/api', {
  cacheTTL: 30000, // 30 seconds
  retryAttempts: 3,
  retryDelay: 1000
});

// Make API client globally available
window.api = apiClient;
window.APIError = APIError;