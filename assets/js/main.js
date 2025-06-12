// Main application JavaScript
document.addEventListener('DOMContentLoaded', function () {
    // Initialize HTMX
    console.log('HTMX initialized');

    // Add global HTMX event listeners
    document.body.addEventListener('htmx:beforeRequest', function (evt) {
        console.log('HTMX request starting:', evt.detail.elt);
        // Add loading states
        const indicator = evt.detail.elt.querySelector('.htmx-indicator');
        if (indicator) {
            indicator.style.opacity = '1';
        }
    });

    document.body.addEventListener('htmx:afterRequest', function (evt) {
        console.log('HTMX request completed:', evt.detail.elt);
        // Remove loading states
        const indicator = evt.detail.elt.querySelector('.htmx-indicator');
        if (indicator) {
            indicator.style.opacity = '0';
        }
    });

    document.body.addEventListener('htmx:responseError', function (evt) {
        console.error('HTMX error:', evt.detail);
        // Show error message
        showNotification('Request failed. Please try again.', 'error');
    });
});

// Notification system
function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = `fixed top-4 right-4 z-50 max-w-sm w-full ${getNotificationClasses(type)} p-4 rounded-lg shadow-lg transform transition-all duration-300 translate-x-full opacity-0`;
    notification.innerHTML = `
        <div class="flex items-center">
            <div class="flex-shrink-0">
                ${getNotificationIcon(type)}
            </div>
            <div class="ml-3">
                <p class="text-sm font-medium">${message}</p>
            </div>
            <div class="ml-auto pl-3">
                <button onclick="this.parentElement.parentElement.parentElement.remove()" class="inline-flex text-gray-400 hover:text-gray-600 focus:outline-none">
                    <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"></path>
                    </svg>
                </button>
            </div>
        </div>
    `;

    document.body.appendChild(notification);

    // Animate in
    setTimeout(() => {
        notification.classList.remove('translate-x-full', 'opacity-0');
    }, 100);

    // Auto remove after 5 seconds
    setTimeout(() => {
        notification.classList.add('translate-x-full', 'opacity-0');
        setTimeout(() => notification.remove(), 300);
    }, 5000);
}

function getNotificationClasses(type) {
    switch (type) {
        case 'success':
            return 'bg-green-50 border border-green-200 text-green-800';
        case 'error':
            return 'bg-red-50 border border-red-200 text-red-800';
        case 'warning':
            return 'bg-yellow-50 border border-yellow-200 text-yellow-800';
        default:
            return 'bg-blue-50 border border-blue-200 text-blue-800';
    }
}

function getNotificationIcon(type) {
    switch (type) {
        case 'success':
            return `<svg class="w-5 h-5 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"></path>
            </svg>`;
        case 'error':
            return `<svg class="w-5 h-5 text-red-400" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd"></path>
            </svg>`;
        case 'warning':
            return `<svg class="w-5 h-5 text-yellow-400" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"></path>
            </svg>`;
        default:
            return `<svg class="w-5 h-5 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"></path>
            </svg>`;
    }
}

// Chat widget functionality
class ChatWidget {
    constructor(containerId) {
        this.container = document.getElementById(containerId);
        this.isOpen = false;
        this.messages = [];
        this.init();
    }

    init() {
        if (!this.container) return;

        this.render();
        this.bindEvents();
    }

    render() {
        this.container.innerHTML = `
            <div class="chat-widget ${this.isOpen ? '' : 'hidden'}" id="chatWindow">
                <div class="chat-header">
                    <div class="flex items-center">
                        <div class="w-2 h-2 bg-green-400 rounded-full mr-2"></div>
                        <span class="font-medium">Support Chat</span>
                    </div>
                    <button onclick="chatWidget.toggle()" class="text-white hover:text-gray-200">
                        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                        </svg>
                    </button>
                </div>
                <div class="chat-messages" id="chatMessages">
                    <div class="message bot">
                        <div class="message-bubble">
                            👋 Hi! How can I help you today?
                        </div>
                    </div>
                </div>
                <div class="chat-input">
                    <form onsubmit="chatWidget.sendMessage(event)" class="flex space-x-2">
                        <input type="text" id="messageInput" placeholder="Type your message..." class="flex-1 input" required>
                        <button type="submit" class="btn-primary">Send</button>
                    </form>
                </div>
            </div>
            <button onclick="chatWidget.toggle()" class="fixed bottom-4 right-4 w-14 h-14 bg-primary-600 text-white rounded-full shadow-lg hover:bg-primary-700 transition-colors duration-200 flex items-center justify-center ${this.isOpen ? 'hidden' : ''}" id="chatToggle">
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-3.582 8-8 8a9.863 9.863 0 01-4.906-1.294A6 6 0 006 18L3 21l3-3h.01M21 12c0-4.418-3.582-8-8-8s-8 3.582-8 8c0 2.421 1.076 4.595 2.777 6.095 0 0 .426.383.777.905"></path>
                </svg>
            </button>
        `;
    }

    toggle() {
        this.isOpen = !this.isOpen;
        const chatWindow = document.getElementById('chatWindow');
        const chatToggle = document.getElementById('chatToggle');

        if (this.isOpen) {
            chatWindow.classList.remove('hidden');
            chatToggle.classList.add('hidden');
        } else {
            chatWindow.classList.add('hidden');
            chatToggle.classList.remove('hidden');
        }
    }

    sendMessage(event) {
        event.preventDefault();
        const input = document.getElementById('messageInput');
        const message = input.value.trim();

        if (!message) return;

        this.addMessage(message, 'user');
        input.value = '';

        // Simulate bot response
        setTimeout(() => {
            this.addBotResponse(message);
        }, 1000);

        // Track analytics
        this.trackMessage(message);
    }

    addMessage(text, sender) {
        const messagesContainer = document.getElementById('chatMessages');
        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${sender}`;
        messageDiv.innerHTML = `<div class="message-bubble">${this.escapeHtml(text)}</div>`;

        messagesContainer.appendChild(messageDiv);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;

        this.messages.push({ text, sender, timestamp: new Date() });
    }

    addBotResponse(userMessage) {
        const responses = [
            "Thanks for your message! I'm here to help.",
            "That's a great question. Let me connect you with someone who can assist.",
            "I understand your concern. How can I help you resolve this?",
            "Thank you for reaching out. What specific information do you need?",
            "I'm happy to help! Can you provide more details about your request?"
        ];

        const response = responses[ Math.floor(Math.random() * responses.length) ];
        this.addMessage(response, 'bot');
    }

    trackMessage(message) {
        // Send analytics to backend
        fetch('/api/widget/usage', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                action: 'message_sent',
                message: message,
                timestamp: new Date().toISOString()
            })
        }).catch(err => console.error('Failed to track message:', err));
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Initialize chat widget
let chatWidget;
document.addEventListener('DOMContentLoaded', function () {
    chatWidget = new ChatWidget('chatWidget');
});

// Real-time updates for dashboard
function initDashboardUpdates() {
    // Update metrics every 30 seconds
    setInterval(() => {
        htmx.trigger('#metrics-container', 'refresh');
    }, 30000);

    // Update charts every 60 seconds
    setInterval(() => {
        htmx.trigger('#charts-container', 'refresh');
    }, 60000);
}

// Form helpers
function resetForm(formId) {
    const form = document.getElementById(formId);
    if (form) {
        form.reset();
        // Clear any validation states
        form.querySelectorAll('.is-invalid').forEach(el => {
            el.classList.remove('is-invalid');
        });
    }
}

// Copy to clipboard helper
function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => {
        showNotification('Copied to clipboard!', 'success');
    }).catch(() => {
        showNotification('Failed to copy to clipboard', 'error');
    });
}

// Export functions to global scope
window.showNotification = showNotification;
window.copyToClipboard = copyToClipboard;
window.resetForm = resetForm;
window.initDashboardUpdates = initDashboardUpdates;
