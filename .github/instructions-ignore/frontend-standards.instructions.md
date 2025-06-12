---
applyTo: '**/*.{html,css,js}'
---

# Frontend Development Standards

This document covers HTML/HTMX, CSS, and JavaScript standards for the Go Chat Widget Dashboard frontend.

## HTML & HTMX Standards

### HTMX Integration Philosophy
- Use HTMX for dynamic content updates without full page reloads
- Implement progressive enhancement - pages should work without JavaScript
- Leverage server-side rendering with Go templates
- Minimize client-side JavaScript complexity

### Template Structure
- Use Go's `html/template` package for server-side rendering
- Implement template inheritance with `{{define}}` and `{{template}}`
- Use semantic HTML5 elements (header, nav, main, section, article)
- Include proper meta tags and accessibility attributes

```html
<!-- Base template -->
{{define "base"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Chat Widget Dashboard</title>
    <link rel="stylesheet" href="/static/css/admin.css">
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body>
    {{template "content" .}}
</body>
</html>
{{end}}
```

### HTMX Patterns and Best Practices

#### Form Submissions
```html
<!-- Auto-updating forms -->
<form hx-post="/api/customers" hx-target="#customer-list" hx-swap="beforeend">
    <input type="text" name="name" required>
    <input type="email" name="email" required>
    <button type="submit">Add Customer</button>
</form>

<!-- Form with validation feedback -->
<form hx-post="/api/customers" hx-target="#form-container" hx-swap="outerHTML">
    <div id="form-container">
        <!-- Form fields -->
    </div>
</form>
```

#### Dynamic Content Loading
```html
<!-- Lazy loading with indicators -->
<div hx-get="/api/chat-logs" 
     hx-trigger="load" 
     hx-target="this" 
     hx-swap="innerHTML"
     hx-indicator="#loading">
    <div id="loading" class="htmx-indicator">Loading chat logs...</div>
</div>

<!-- Infinite scroll -->
<div id="chat-logs">
    <!-- Initial content -->
    <div hx-get="/api/chat-logs?page=2" 
         hx-trigger="revealed" 
         hx-target="#chat-logs" 
         hx-swap="beforeend">
    </div>
</div>
```

#### Real-time Updates
```html
<!-- Polling for updates -->
<div id="status" 
     hx-get="/api/status" 
     hx-trigger="every 5s" 
     hx-target="this" 
     hx-swap="innerHTML">
    Status: Unknown
</div>

<!-- WebSocket integration -->
<div id="chat-container" 
     hx-ext="ws" 
     ws-connect="/ws">
    <div id="messages"></div>
    <form ws-send>
        <input type="text" name="message" placeholder="Type a message...">
        <button type="submit">Send</button>
    </form>
</div>
```

#### Navigation and Boosting
```html
<!-- Boost navigation links -->
<nav hx-boost="true">
    <a href="/dashboard">Dashboard</a>
    <a href="/customers">Customers</a>
    <a href="/chat-logs">Chat Logs</a>
</nav>

<!-- Custom transitions -->
<div hx-get="/api/page" 
     hx-target="#main-content" 
     hx-swap="innerHTML transition:true">
</div>
```

### Accessibility Guidelines
- Use proper ARIA labels and roles for dynamic content
- Ensure keyboard navigation works with HTMX interactions
- Provide alt text for images and icons
- Use sufficient color contrast ratios (WCAG 2.1 AA)
- Include focus management for dynamic updates

```html
<!-- Accessible HTMX form -->
<form hx-post="/api/customers" 
      hx-target="#customer-list" 
      hx-swap="beforeend"
      aria-label="Add new customer">
    <label for="customer-name">Customer Name</label>
    <input type="text" 
           id="customer-name" 
           name="name" 
           required 
           aria-describedby="name-error">
    <div id="name-error" class="error-message" aria-live="polite"></div>
    
    <button type="submit" aria-describedby="submit-status">Add Customer</button>
    <div id="submit-status" aria-live="polite"></div>
</form>
```

## CSS Standards

### Methodology and Organization
- Use BEM (Block Element Modifier) naming convention
- Implement mobile-first responsive design
- Use CSS Grid and Flexbox for layouts
- Follow progressive enhancement principles

### CSS Variables and Design System
```css
:root {
    /* Color palette */
    --primary-color: #007bff;
    --primary-dark: #0056b3;
    --secondary-color: #6c757d;
    --success-color: #28a745;
    --warning-color: #ffc107;
    --danger-color: #dc3545;
    --info-color: #17a2b8;
    
    /* Typography */
    --font-family-base: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    --font-family-mono: "SF Mono", Monaco, Consolas, monospace;
    --font-size-base: 1rem;
    --line-height-base: 1.5;
    
    /* Spacing */
    --spacing-xs: 0.25rem;
    --spacing-sm: 0.5rem;
    --spacing-md: 1rem;
    --spacing-lg: 1.5rem;
    --spacing-xl: 3rem;
    
    /* Layout */
    --border-radius: 0.375rem;
    --border-width: 1px;
    --max-width: 1200px;
}
```

### BEM Component Structure
```css
/* Block: Chat Widget */
.chat-widget {
    position: fixed;
    bottom: 1rem;
    right: 1rem;
    width: 350px;
    max-height: 500px;
    background: white;
    border-radius: var(--border-radius);
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

/* Element: Chat Widget Header */
.chat-widget__header {
    padding: var(--spacing-md);
    background: var(--primary-color);
    color: white;
    border-radius: var(--border-radius) var(--border-radius) 0 0;
    cursor: pointer;
}

.chat-widget__title {
    margin: 0;
    font-size: 1.1rem;
    font-weight: 600;
}

/* Element: Chat Widget Body */
.chat-widget__body {
    padding: var(--spacing-md);
    max-height: 300px;
    overflow-y: auto;
}

/* Element: Chat Widget Messages */
.chat-widget__messages {
    display: flex;
    flex-direction: column;
    gap: var(--spacing-sm);
}

.chat-widget__message {
    padding: var(--spacing-sm) var(--spacing-md);
    border-radius: var(--border-radius);
    max-width: 80%;
    word-wrap: break-word;
}

/* Modifier: User Messages */
.chat-widget__message--user {
    background: var(--primary-color);
    color: white;
    align-self: flex-end;
}

/* Modifier: Bot Messages */
.chat-widget__message--bot {
    background: #f8f9fa;
    color: #333;
    align-self: flex-start;
}

/* Modifier: Collapsed State */
.chat-widget--collapsed .chat-widget__body {
    display: none;
}
```

### Responsive Design
```css
/* Mobile-first approach */
.dashboard-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: var(--spacing-md);
    padding: var(--spacing-md);
}

/* Tablet and up */
@media (min-width: 768px) {
    .dashboard-grid {
        grid-template-columns: 250px 1fr;
        padding: var(--spacing-lg);
    }
}

/* Desktop and up */
@media (min-width: 1024px) {
    .dashboard-grid {
        grid-template-columns: 300px 1fr 250px;
        max-width: var(--max-width);
        margin: 0 auto;
    }
}
```

### HTMX-specific CSS
```css
/* Loading indicators */
.htmx-indicator {
    opacity: 0;
    transition: opacity 0.3s ease;
}

.htmx-request .htmx-indicator {
    opacity: 1;
}

.htmx-request.htmx-indicator {
    opacity: 1;
}

/* Smooth transitions for content swapping */
.htmx-settling {
    transition: all 0.3s ease;
}

/* Error states */
.htmx-error {
    background-color: #f8d7da;
    border: 1px solid #f5c6cb;
    color: #721c24;
    padding: var(--spacing-sm) var(--spacing-md);
    border-radius: var(--border-radius);
}
```

### Performance Optimization
- Minimize CSS bundle size through critical CSS extraction
- Use efficient selectors (avoid deep nesting > 3 levels)
- Implement CSS containment for performance
- Use `will-change` sparingly for animations

```css
/* Efficient selectors */
.btn { /* Good: single class */ }
.nav .btn { /* Acceptable: 2 levels */ }
.sidebar .nav .btn { /* Limit: 3 levels max */ }

/* CSS containment for performance */
.chat-widget {
    contain: layout style paint;
}

/* Animation performance */
.modal {
    will-change: transform, opacity;
    transition: transform 0.3s ease, opacity 0.3s ease;
}

.modal.is-open {
    will-change: auto; /* Remove after animation */
}
```

## JavaScript Standards

### Modern JavaScript Practices
- Use ES6+ features (const/let, arrow functions, destructuring)
- Implement async/await for asynchronous operations
- Use modules for code organization
- Follow functional programming principles where appropriate

### HTMX Event Handling
```javascript
// Global HTMX event listeners
document.addEventListener('htmx:afterRequest', function(event) {
    if (event.detail.successful) {
        showNotification('Operation successful', 'success');
    } else {
        showNotification('Operation failed', 'error');
    }
});

document.addEventListener('htmx:beforeRequest', function(event) {
    // Add loading state
    event.target.classList.add('loading');
});

document.addEventListener('htmx:afterRequest', function(event) {
    // Remove loading state
    event.target.classList.remove('loading');
});

// Custom HTMX events
document.addEventListener('htmx:afterSwap', function(event) {
    // Re-initialize components after content swap
    initializeTooltips(event.target);
    initializeDatePickers(event.target);
});
```

### WebSocket Integration
```javascript
class ChatWebSocket {
    constructor(url) {
        this.url = url;
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000;
    }
    
    connect() {
        try {
            this.ws = new WebSocket(this.url);
            
            this.ws.onopen = () => {
                console.log('WebSocket connected');
                this.reconnectAttempts = 0;
                this.updateConnectionStatus('connected');
            };
            
            this.ws.onmessage = (event) => {
                const message = JSON.parse(event.data);
                this.handleMessage(message);
            };
            
            this.ws.onclose = () => {
                console.log('WebSocket disconnected');
                this.updateConnectionStatus('disconnected');
                this.attemptReconnect();
            };
            
            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                this.updateConnectionStatus('error');
            };
        } catch (error) {
            console.error('Failed to create WebSocket:', error);
        }
    }
    
    handleMessage(message) {
        // Trigger HTMX event for new messages
        htmx.trigger('#chat-container', 'newMessage', message);
        
        // Update UI directly
        const messagesContainer = document.getElementById('messages');
        const messageElement = this.createMessageElement(message);
        messagesContainer.appendChild(messageElement);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }
    
    sendMessage(message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
        } else {
            console.error('WebSocket is not connected');
            this.showError('Connection lost. Please refresh the page.');
        }
    }
    
    attemptReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            console.log(`Attempting to reconnect... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
            
            setTimeout(() => {
                this.connect();
            }, this.reconnectDelay * this.reconnectAttempts);
        } else {
            this.showError('Unable to connect. Please refresh the page.');
        }
    }
    
    updateConnectionStatus(status) {
        const statusElement = document.getElementById('connection-status');
        if (statusElement) {
            statusElement.className = `connection-status connection-status--${status}`;
            statusElement.textContent = status.charAt(0).toUpperCase() + status.slice(1);
        }
    }
    
    createMessageElement(message) {
        const div = document.createElement('div');
        div.className = `chat-widget__message chat-widget__message--${message.isFromUser ? 'user' : 'bot'}`;
        div.textContent = message.content;
        div.setAttribute('data-timestamp', message.timestamp);
        return div;
    }
    
    showError(message) {
        // Trigger HTMX event or show notification
        htmx.trigger(document.body, 'showNotification', {
            message: message,
            type: 'error'
        });
    }
}

// Initialize WebSocket connection
const chatWS = new ChatWebSocket('ws://localhost:8080/ws');
chatWS.connect();
```

### Form Validation and UX Enhancements
```javascript
// Client-side validation before HTMX submission
document.addEventListener('htmx:configRequest', function(event) {
    if (event.target.matches('form[data-validate]')) {
        const form = event.target;
        const errors = validateForm(form);
        
        if (errors.length > 0) {
            event.preventDefault();
            showValidationErrors(form, errors);
        }
    }
});

function validateForm(form) {
    const errors = [];
    const requiredFields = form.querySelectorAll('[required]');
    
    requiredFields.forEach(field => {
        if (!field.value.trim()) {
            errors.push({
                field: field.name,
                message: `${field.labels[0]?.textContent || field.name} is required`
            });
        }
    });
    
    // Email validation
    const emailFields = form.querySelectorAll('input[type="email"]');
    emailFields.forEach(field => {
        if (field.value && !isValidEmail(field.value)) {
            errors.push({
                field: field.name,
                message: 'Please enter a valid email address'
            });
        }
    });
    
    return errors;
}

function showValidationErrors(form, errors) {
    // Clear previous errors
    form.querySelectorAll('.error-message').forEach(el => el.remove());
    form.querySelectorAll('.field-error').forEach(el => el.classList.remove('field-error'));
    
    // Show new errors
    errors.forEach(error => {
        const field = form.querySelector(`[name="${error.field}"]`);
        if (field) {
            field.classList.add('field-error');
            
            const errorElement = document.createElement('div');
            errorElement.className = 'error-message';
            errorElement.textContent = error.message;
            errorElement.setAttribute('aria-live', 'polite');
            
            field.parentNode.appendChild(errorElement);
        }
    });
}

function isValidEmail(email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
}
```

### Notification System
```javascript
class NotificationManager {
    constructor() {
        this.container = this.createContainer();
        this.notifications = new Map();
    }
    
    createContainer() {
        const container = document.createElement('div');
        container.id = 'notification-container';
        container.className = 'notification-container';
        document.body.appendChild(container);
        return container;
    }
    
    show(message, type = 'info', duration = 5000) {
        const notification = this.createNotification(message, type);
        this.container.appendChild(notification);
        
        // Trigger entrance animation
        requestAnimationFrame(() => {
            notification.classList.add('notification--visible');
        });
        
        // Auto-dismiss
        if (duration > 0) {
            setTimeout(() => {
                this.dismiss(notification);
            }, duration);
        }
        
        return notification;
    }
    
    createNotification(message, type) {
        const notification = document.createElement('div');
        notification.className = `notification notification--${type}`;
        notification.innerHTML = `
            <div class="notification__content">
                <span class="notification__message">${message}</span>
                <button class="notification__close" aria-label="Close notification">×</button>
            </div>
        `;
        
        // Add close button functionality
        const closeBtn = notification.querySelector('.notification__close');
        closeBtn.addEventListener('click', () => this.dismiss(notification));
        
        return notification;
    }
    
    dismiss(notification) {
        notification.classList.add('notification--hiding');
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 300);
    }
}

// Global notification instance
const notifications = new NotificationManager();

// Global notification function
window.showNotification = (message, type = 'info', duration = 5000) => {
    return notifications.show(message, type, duration);
};

// HTMX notification event listener
document.addEventListener('showNotification', function(event) {
    const { message, type, duration } = event.detail;
    showNotification(message, type, duration);
});
```

### Utility Functions
```javascript
// Debounce function for search inputs
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// Throttle function for scroll events
function throttle(func, limit) {
    let inThrottle;
    return function(...args) {
        if (!inThrottle) {
            func.apply(this, args);
            inThrottle = true;
            setTimeout(() => inThrottle = false, limit);
        }
    };
}

// Format date utilities
function formatRelativeTime(date) {
    const now = new Date();
    const diff = now - date;
    const seconds = Math.floor(diff / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);
    
    if (seconds < 60) return 'just now';
    if (minutes < 60) return `${minutes}m ago`;
    if (hours < 24) return `${hours}h ago`;
    if (days < 7) return `${days}d ago`;
    return date.toLocaleDateString();
}

// Local storage helpers
const storage = {
    get(key, defaultValue = null) {
        try {
            const item = localStorage.getItem(key);
            return item ? JSON.parse(item) : defaultValue;
        } catch (error) {
            console.error('Error reading from localStorage:', error);
            return defaultValue;
        }
    },
    
    set(key, value) {
        try {
            localStorage.setItem(key, JSON.stringify(value));
        } catch (error) {
            console.error('Error writing to localStorage:', error);
        }
    },
    
    remove(key) {
        try {
            localStorage.removeItem(key);
        } catch (error) {
            console.error('Error removing from localStorage:', error);
        }
    }
};
```

### Component Initialization
```javascript
// Component initialization system
class ComponentManager {
    constructor() {
        this.components = new Map();
        this.initializeOnLoad();
    }
    
    register(selector, initFunction) {
        this.components.set(selector, initFunction);
    }
    
    initialize(container = document) {
        this.components.forEach((initFunction, selector) => {
            const elements = container.querySelectorAll(selector);
            elements.forEach(element => {
                if (!element.hasAttribute('data-initialized')) {
                    initFunction(element);
                    element.setAttribute('data-initialized', 'true');
                }
            });
        });
    }
    
    initializeOnLoad() {
        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', () => this.initialize());
        } else {
            this.initialize();
        }
        
        // Re-initialize components after HTMX swaps
        document.addEventListener('htmx:afterSwap', (event) => {
            this.initialize(event.target);
        });
    }
}

// Global component manager
const components = new ComponentManager();

// Register components
components.register('[data-tooltip]', function(element) {
    // Tooltip initialization
    element.addEventListener('mouseenter', function() {
        showTooltip(this, this.dataset.tooltip);
    });
    
    element.addEventListener('mouseleave', function() {
        hideTooltip(this);
    });
});

components.register('[data-modal-trigger]', function(element) {
    element.addEventListener('click', function(e) {
        e.preventDefault();
        const modalId = this.dataset.modalTrigger;
        const modal = document.getElementById(modalId);
        if (modal) {
            openModal(modal);
        }
    });
});

components.register('.search-input', function(element) {
    const debouncedSearch = debounce(function(query) {
        htmx.trigger(element, 'search', { query });
    }, 300);
    
    element.addEventListener('input', function() {
        debouncedSearch(this.value);
    });
});
```

### Code Quality and Best Practices
- Use strict mode (`'use strict';`) in all JavaScript files
- Implement error boundaries for async operations
- Use meaningful variable and function names
- Comment complex business logic
- Avoid global variables except for truly global utilities

```javascript
'use strict';

// Good: Descriptive naming
const calculateChatResponseTime = (startTime, endTime) => {
    return endTime - startTime;
};

// Good: Error handling for async operations
async function fetchChatHistory(userId) {
    try {
        const response = await fetch(`/api/users/${userId}/chat-history`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return await response.json();
    } catch (error) {
        console.error('Failed to fetch chat history:', error);
        showNotification('Failed to load chat history', 'error');
        return [];
    }
}

// Good: Modular organization
const ChatUtils = {
    formatMessage(content, isFromUser) {
        return {
            content: content.trim(),
            timestamp: new Date().toISOString(),
            isFromUser: Boolean(isFromUser)
        };
    },
    
    validateMessage(message) {
        return message && 
               typeof message === 'string' && 
               message.trim().length > 0 && 
               message.length <= 1000;
    }
};
```
