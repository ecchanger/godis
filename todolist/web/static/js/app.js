class TodoApp {
    constructor() {
        this.todos = [];
        this.currentFilter = 'all';
        this.editingTodoId = null;
        this.initializeElements();
        this.bindEvents();
        this.loadTodos();
        this.loadTheme();
    }

    initializeElements() {
        // 表单元素
        this.todoForm = document.getElementById('todoForm');
        this.todoTitle = document.getElementById('todoTitle');
        this.todoDescription = document.getElementById('todoDescription');
        
        // 筛选器
        this.filterBtns = document.querySelectorAll('.filter-btn');
        
        // 列表和计数器
        this.todoList = document.getElementById('todoList');
        this.emptyState = document.getElementById('emptyState');
        this.allCount = document.getElementById('allCount');
        this.activeCount = document.getElementById('activeCount');
        this.completedCount = document.getElementById('completedCount');
        this.totalTodos = document.getElementById('totalTodos');
        this.clearCompleted = document.getElementById('clearCompleted');
        
        // 模态框
        this.editModal = document.getElementById('editModal');
        this.editForm = document.getElementById('editForm');
        this.editTitle = document.getElementById('editTitle');
        this.editDescription = document.getElementById('editDescription');
        this.closeModal = document.getElementById('closeModal');
        this.cancelEdit = document.getElementById('cancelEdit');
        
        // 主题切换
        this.themeToggle = document.getElementById('themeToggle');
        
        // 加载和提示
        this.loadingOverlay = document.getElementById('loadingOverlay');
        this.toastContainer = document.getElementById('toastContainer');
    }

    bindEvents() {
        // 表单提交
        this.todoForm.addEventListener('submit', (e) => this.handleAddTodo(e));
        
        // 筛选器
        this.filterBtns.forEach(btn => {
            btn.addEventListener('click', (e) => this.handleFilter(e));
        });
        
        // 编辑模态框
        this.editForm.addEventListener('submit', (e) => this.handleEditSubmit(e));
        this.closeModal.addEventListener('click', () => this.closeEditModal());
        this.cancelEdit.addEventListener('click', () => this.closeEditModal());
        this.editModal.addEventListener('click', (e) => {
            if (e.target === this.editModal) this.closeEditModal();
        });
        
        // 清除已完成
        this.clearCompleted.addEventListener('click', () => this.handleClearCompleted());
        
        // 主题切换
        this.themeToggle.addEventListener('click', () => this.toggleTheme());
        
        // 键盘快捷键
        document.addEventListener('keydown', (e) => this.handleKeydown(e));
        
        // 自动调整文本域高度
        this.todoDescription.addEventListener('input', () => this.autoResize(this.todoDescription));
        this.editDescription.addEventListener('input', () => this.autoResize(this.editDescription));
    }

    // API 请求方法
    async apiRequest(url, options = {}) {
        this.showLoading();
        try {
            const response = await fetch(url, {
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                },
                ...options
            });
            
            const data = await response.json();
            
            if (!data.success) {
                throw new Error(data.message);
            }
            
            return data;
        } catch (error) {
            this.showToast(error.message || '请求失败', 'error');
            throw error;
        } finally {
            this.hideLoading();
        }
    }

    // 加载所有待办事项
    async loadTodos() {
        try {
            const response = await this.apiRequest('/api/todos');
            this.todos = response.data;
            this.renderTodos();
            this.updateCounts();
        } catch (error) {
            console.error('Failed to load todos:', error);
        }
    }

    // 添加待办事项
    async handleAddTodo(e) {
        e.preventDefault();
        
        const title = this.todoTitle.value.trim();
        const description = this.todoDescription.value.trim();
        
        if (!title) {
            this.showToast('请输入任务标题', 'warning');
            return;
        }

        try {
            const response = await this.apiRequest('/api/todos', {
                method: 'POST',
                body: JSON.stringify({
                    title,
                    description,
                    completed: false
                })
            });

            this.todos.unshift(response.data);
            this.renderTodos();
            this.updateCounts();
            this.clearForm();
            this.showToast('任务添加成功', 'success');
        } catch (error) {
            console.error('Failed to add todo:', error);
        }
    }

    // 切换完成状态
    async toggleComplete(id) {
        try {
            const response = await this.apiRequest(`/api/todos/${id}/toggle`, {
                method: 'PATCH'
            });

            const todoIndex = this.todos.findIndex(todo => todo.id === id);
            if (todoIndex !== -1) {
                this.todos[todoIndex] = response.data;
                this.renderTodos();
                this.updateCounts();
            }
        } catch (error) {
            console.error('Failed to toggle todo:', error);
        }
    }

    // 删除待办事项
    async deleteTodo(id) {
        if (!confirm('确定要删除这个任务吗？')) {
            return;
        }

        try {
            await this.apiRequest(`/api/todos/${id}`, {
                method: 'DELETE'
            });

            this.todos = this.todos.filter(todo => todo.id !== id);
            this.renderTodos();
            this.updateCounts();
            this.showToast('任务删除成功', 'success');
        } catch (error) {
            console.error('Failed to delete todo:', error);
        }
    }

    // 编辑待办事项
    editTodo(id) {
        const todo = this.todos.find(todo => todo.id === id);
        if (todo) {
            this.editingTodoId = id;
            this.editTitle.value = todo.title;
            this.editDescription.value = todo.description || '';
            this.openEditModal();
        }
    }

    // 提交编辑
    async handleEditSubmit(e) {
        e.preventDefault();
        
        const title = this.editTitle.value.trim();
        const description = this.editDescription.value.trim();
        
        if (!title) {
            this.showToast('请输入任务标题', 'warning');
            return;
        }

        if (!this.editingTodoId) return;

        try {
            const todo = this.todos.find(t => t.id === this.editingTodoId);
            const response = await this.apiRequest(`/api/todos/${this.editingTodoId}`, {
                method: 'PUT',
                body: JSON.stringify({
                    title,
                    description,
                    completed: todo.completed
                })
            });

            const todoIndex = this.todos.findIndex(todo => todo.id === this.editingTodoId);
            if (todoIndex !== -1) {
                this.todos[todoIndex] = response.data;
                this.renderTodos();
                this.updateCounts();
            }
            
            this.closeEditModal();
            this.showToast('任务更新成功', 'success');
        } catch (error) {
            console.error('Failed to update todo:', error);
        }
    }

    // 筛选器处理
    handleFilter(e) {
        const filter = e.target.getAttribute('data-filter');
        if (filter) {
            this.currentFilter = filter;
            this.filterBtns.forEach(btn => btn.classList.remove('active'));
            e.target.classList.add('active');
            this.renderTodos();
        }
    }

    // 清除已完成的任务
    async handleClearCompleted() {
        const completedTodos = this.todos.filter(todo => todo.completed);
        
        if (completedTodos.length === 0) {
            this.showToast('没有已完成的任务', 'warning');
            return;
        }

        if (!confirm(`确定要删除 ${completedTodos.length} 个已完成的任务吗？`)) {
            return;
        }

        try {
            // 并发删除所有已完成的任务
            await Promise.all(
                completedTodos.map(todo => 
                    this.apiRequest(`/api/todos/${todo.id}`, { method: 'DELETE' })
                )
            );

            this.todos = this.todos.filter(todo => !todo.completed);
            this.renderTodos();
            this.updateCounts();
            this.showToast(`成功删除 ${completedTodos.length} 个已完成任务`, 'success');
        } catch (error) {
            console.error('Failed to clear completed todos:', error);
            // 重新加载以确保数据一致性
            this.loadTodos();
        }
    }

    // 渲染待办事项列表
    renderTodos() {
        const filteredTodos = this.getFilteredTodos();
        
        if (filteredTodos.length === 0) {
            this.todoList.style.display = 'none';
            this.emptyState.style.display = 'block';
        } else {
            this.todoList.style.display = 'block';
            this.emptyState.style.display = 'none';
            
            this.todoList.innerHTML = filteredTodos.map(todo => this.createTodoElement(todo)).join('');
            
            // 绑定事件
            this.bindTodoEvents();
        }
    }

    // 创建待办事项元素
    createTodoElement(todo) {
        const createdAt = new Date(todo.created_at).toLocaleString('zh-CN');
        const updatedAt = new Date(todo.updated_at).toLocaleString('zh-CN');
        
        return `
            <li class="todo-item ${todo.completed ? 'completed' : ''}">
                <div class="todo-checkbox ${todo.completed ? 'checked' : ''}" data-id="${todo.id}"></div>
                <div class="todo-content">
                    <div class="todo-title">${this.escapeHtml(todo.title)}</div>
                    ${todo.description ? `<div class="todo-description">${this.escapeHtml(todo.description)}</div>` : ''}
                    <div class="todo-meta">
                        创建时间: ${createdAt}
                        ${todo.updated_at !== todo.created_at ? ` | 更新时间: ${updatedAt}` : ''}
                    </div>
                </div>
                <div class="todo-actions">
                    <button class="action-btn edit" data-id="${todo.id}" title="编辑">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="action-btn delete" data-id="${todo.id}" title="删除">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </li>
        `;
    }

    // 绑定待办事项事件
    bindTodoEvents() {
        // 复选框事件
        document.querySelectorAll('.todo-checkbox').forEach(checkbox => {
            checkbox.addEventListener('click', (e) => {
                const id = parseInt(e.target.getAttribute('data-id'));
                this.toggleComplete(id);
            });
        });

        // 编辑按钮事件
        document.querySelectorAll('.action-btn.edit').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const id = parseInt(e.target.closest('.action-btn').getAttribute('data-id'));
                this.editTodo(id);
            });
        });

        // 删除按钮事件
        document.querySelectorAll('.action-btn.delete').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const id = parseInt(e.target.closest('.action-btn').getAttribute('data-id'));
                this.deleteTodo(id);
            });
        });
    }

    // 获取筛选后的待办事项
    getFilteredTodos() {
        switch (this.currentFilter) {
            case 'active':
                return this.todos.filter(todo => !todo.completed);
            case 'completed':
                return this.todos.filter(todo => todo.completed);
            default:
                return this.todos;
        }
    }

    // 更新计数器
    updateCounts() {
        const allCount = this.todos.length;
        const activeCount = this.todos.filter(todo => !todo.completed).length;
        const completedCount = this.todos.filter(todo => todo.completed).length;

        this.allCount.textContent = allCount;
        this.activeCount.textContent = activeCount;
        this.completedCount.textContent = completedCount;
        this.totalTodos.textContent = allCount;

        // 显示/隐藏清除已完成按钮
        this.clearCompleted.style.display = completedCount > 0 ? 'flex' : 'none';
    }

    // 模态框操作
    openEditModal() {
        this.editModal.classList.add('active');
        this.editTitle.focus();
    }

    closeEditModal() {
        this.editModal.classList.remove('active');
        this.editingTodoId = null;
        this.editForm.reset();
    }

    // 主题切换
    toggleTheme() {
        const currentTheme = document.documentElement.getAttribute('data-theme');
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
        
        document.documentElement.setAttribute('data-theme', newTheme);
        localStorage.setItem('theme', newTheme);
        
        const icon = this.themeToggle.querySelector('i');
        icon.className = newTheme === 'dark' ? 'fas fa-sun' : 'fas fa-moon';
    }

    // 加载主题
    loadTheme() {
        const savedTheme = localStorage.getItem('theme') || 'light';
        document.documentElement.setAttribute('data-theme', savedTheme);
        
        const icon = this.themeToggle.querySelector('i');
        icon.className = savedTheme === 'dark' ? 'fas fa-sun' : 'fas fa-moon';
    }

    // 键盘快捷键
    handleKeydown(e) {
        // Ctrl+Enter 或 Cmd+Enter 提交表单
        if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
            if (this.editModal.classList.contains('active')) {
                this.editForm.dispatchEvent(new Event('submit'));
            } else {
                this.todoForm.dispatchEvent(new Event('submit'));
            }
        }
        
        // Escape 关闭模态框
        if (e.key === 'Escape' && this.editModal.classList.contains('active')) {
            this.closeEditModal();
        }
    }

    // 工具方法
    clearForm() {
        this.todoForm.reset();
        this.autoResize(this.todoDescription);
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    autoResize(textarea) {
        textarea.style.height = 'auto';
        textarea.style.height = textarea.scrollHeight + 'px';
    }

    // 显示/隐藏加载指示器
    showLoading() {
        this.loadingOverlay.style.display = 'flex';
    }

    hideLoading() {
        this.loadingOverlay.style.display = 'none';
    }

    // 显示提示消息
    showToast(message, type = 'info') {
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.textContent = message;
        
        this.toastContainer.appendChild(toast);
        
        // 3秒后自动移除
        setTimeout(() => {
            if (toast.parentNode) {
                toast.style.opacity = '0';
                toast.style.transform = 'translateX(100%)';
                setTimeout(() => {
                    if (toast.parentNode) {
                        this.toastContainer.removeChild(toast);
                    }
                }, 300);
            }
        }, 3000);
    }
}

// 初始化应用
document.addEventListener('DOMContentLoaded', () => {
    new TodoApp();
});