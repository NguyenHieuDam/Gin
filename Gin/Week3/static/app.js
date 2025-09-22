// Chat App JavaScript
class ChatApp {
    constructor() {
        this.ws = null;
        this.currentUser = null;
        this.currentRoom = 'general';
        this.isConnected = false;
        this.typingTimeout = null;
        this.typingUsers = new Set();
        
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.checkAuth();
    }

    setupEventListeners() {
        // Login form
        document.getElementById('loginForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleLogin();
        });

        // Register form
        document.getElementById('registerForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleRegister();
        });

        // Message input
        const messageInput = document.getElementById('messageInput');
        messageInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });

        messageInput.addEventListener('input', () => {
            this.handleTyping();
        });

        // Send button
        document.getElementById('sendButton').addEventListener('click', () => {
            this.sendMessage();
        });

        // Room switching
        document.querySelectorAll('.room-item').forEach(item => {
            item.addEventListener('click', () => {
                const room = item.dataset.room;
                this.switchRoom(room);
            });
        });
    }

    async checkAuth() {
        const token = localStorage.getItem('token');
        const user = localStorage.getItem('user');
        
        if (token && user) {
            this.currentUser = JSON.parse(user);
            this.showChatInterface();
            this.connectWebSocket();
        } else {
            this.showLoginModal();
        }
    }

    async handleLogin() {
        const formData = new FormData(document.getElementById('loginForm'));
        const data = {
            username: formData.get('username'),
            password: formData.get('password')
        };

        try {
            const response = await fetch('/api/v1/auth/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(data)
            });

            const result = await response.json();

            if (response.ok) {
                this.currentUser = result.user;
                localStorage.setItem('token', result.token);
                localStorage.setItem('user', JSON.stringify(result.user));
                
                this.showNotification('Đăng nhập thành công!', 'success');
                this.showChatInterface();
                this.connectWebSocket();
            } else {
                this.showNotification(result.error || 'Đăng nhập thất bại', 'error');
            }
        } catch (error) {
            this.showNotification('Lỗi kết nối', 'error');
            console.error('Login error:', error);
        }
    }

    async handleRegister() {
        const formData = new FormData(document.getElementById('registerForm'));
        const data = {
            username: formData.get('username'),
            email: formData.get('email'),
            password: formData.get('password')
        };

        try {
            const response = await fetch('/api/v1/auth/register', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(data)
            });

            const result = await response.json();

            if (response.ok) {
                this.showNotification('Đăng ký thành công! Vui lòng đăng nhập.', 'success');
                this.showLoginModal();
                // Clear register form
                document.getElementById('registerForm').reset();
            } else {
                this.showNotification(result.error || 'Đăng ký thất bại', 'error');
            }
        } catch (error) {
            this.showNotification('Lỗi kết nối', 'error');
            console.error('Register error:', error);
        }
    }

    connectWebSocket() {
        if (this.ws) {
            this.ws.close();
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/api/v1/ws/?user_id=${this.currentUser.id}&username=${encodeURIComponent(this.currentUser.username)}&room_id=${this.currentRoom}`;
        
        this.showLoading(true);
        
        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            this.isConnected = true;
            this.showLoading(false);
            this.showNotification('Đã kết nối', 'success');
            this.loadRecentMessages();
        };

        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.handleWebSocketMessage(message);
        };

        this.ws.onclose = () => {
            this.isConnected = false;
            this.showLoading(false);
            this.showNotification('Mất kết nối', 'warning');
            
            // Try to reconnect after 3 seconds
            setTimeout(() => {
                if (!this.isConnected) {
                    this.connectWebSocket();
                }
            }, 3000);
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.showNotification('Lỗi kết nối WebSocket', 'error');
        };
    }

    handleWebSocketMessage(message) {
        switch (message.type) {
            case 'message':
                this.addMessage(message.data);
                break;
            case 'user_joined':
                this.handleUserJoined(message.user);
                break;
            case 'user_left':
                this.handleUserLeft(message.user);
                break;
            case 'typing':
                this.handleTypingIndicator(message.user);
                break;
            case 'pong':
                // Heartbeat response
                break;
            default:
                console.log('Unknown message type:', message.type);
        }
    }

    addMessage(message) {
        const messagesContainer = document.getElementById('messages');
        const messageElement = this.createMessageElement(message);
        messagesContainer.appendChild(messageElement);
        
        // Auto scroll to bottom
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }

    createMessageElement(message) {
        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${message.user_id === this.currentUser.id ? 'own' : ''}`;
        
        const time = new Date(message.created_at).toLocaleTimeString('vi-VN', {
            hour: '2-digit',
            minute: '2-digit'
        });

        messageDiv.innerHTML = `
            <div class="message-avatar">
                ${message.username.charAt(0).toUpperCase()}
            </div>
            <div class="message-content">
                <div class="message-header">
                    <span class="message-username">${message.username}</span>
                    <span class="message-time">${time}</span>
                </div>
                <div class="message-text">${this.escapeHtml(message.content)}</div>
            </div>
        `;

        return messageDiv;
    }

    sendMessage() {
        const messageInput = document.getElementById('messageInput');
        const content = messageInput.value.trim();
        
        if (!content || !this.isConnected) {
            return;
        }

        const message = {
            type: 'message',
            data: {
                content: content,
                room_id: this.currentRoom
            }
        };

        this.ws.send(JSON.stringify(message));
        messageInput.value = '';
        
        // Clear typing indicator
        this.clearTypingIndicator();
    }

    handleTyping() {
        if (!this.isConnected) return;

        const message = {
            type: 'typing'
        };

        this.ws.send(JSON.stringify(message));

        // Clear previous timeout
        if (this.typingTimeout) {
            clearTimeout(this.typingTimeout);
        }

        // Set new timeout to stop typing indicator
        this.typingTimeout = setTimeout(() => {
            this.clearTypingIndicator();
        }, 2000);
    }

    handleTypingIndicator(user) {
        if (user.id === this.currentUser.id) return;

        this.typingUsers.add(user.username);
        this.updateTypingIndicator();
    }

    clearTypingIndicator() {
        this.typingUsers.clear();
        this.updateTypingIndicator();
    }

    updateTypingIndicator() {
        const indicator = document.getElementById('typingIndicator');
        const usersSpan = document.getElementById('typingUsers');
        
        if (this.typingUsers.size === 0) {
            indicator.style.display = 'none';
        } else {
            const users = Array.from(this.typingUsers);
            if (users.length === 1) {
                usersSpan.textContent = `${users[0]} đang gõ...`;
            } else if (users.length === 2) {
                usersSpan.textContent = `${users[0]} và ${users[1]} đang gõ...`;
            } else {
                usersSpan.textContent = `${users[0]} và ${users.length - 1} người khác đang gõ...`;
            }
            indicator.style.display = 'block';
        }
    }

    handleUserJoined(user) {
        this.showNotification(`${user.username} đã tham gia`, 'success');
        this.updateOnlineUsers();
    }

    handleUserLeft(user) {
        this.showNotification(`${user.username} đã rời khỏi`, 'warning');
        this.updateOnlineUsers();
    }

    async updateOnlineUsers() {
        try {
            const response = await fetch(`/api/v1/ws/${this.currentRoom}/users`);
            const result = await response.json();
            
            if (response.ok) {
                this.displayOnlineUsers(result.online_users);
                document.getElementById('onlineCount').textContent = `${result.count} người trực tuyến`;
            }
        } catch (error) {
            console.error('Error fetching online users:', error);
        }
    }

    displayOnlineUsers(users) {
        const container = document.getElementById('onlineUsers');
        container.innerHTML = '';
        
        users.forEach(user => {
            const userElement = document.createElement('div');
            userElement.className = 'online-user';
            userElement.innerHTML = `
                <div class="user-avatar">${user.username.charAt(0).toUpperCase()}</div>
                <span>${user.username}</span>
                <div class="user-status"></div>
            `;
            container.appendChild(userElement);
        });
    }

    switchRoom(roomId) {
        if (roomId === this.currentRoom) return;

        this.currentRoom = roomId;
        document.getElementById('currentRoom').textContent = roomId;
        
        // Update active room in sidebar
        document.querySelectorAll('.room-item').forEach(item => {
            item.classList.remove('active');
        });
        document.querySelector(`[data-room="${roomId}"]`).classList.add('active');
        
        // Clear messages
        document.getElementById('messages').innerHTML = '';
        
        // Reconnect WebSocket with new room
        if (this.isConnected) {
            this.connectWebSocket();
        }
    }

    async loadRecentMessages() {
        try {
            const response = await fetch(`/api/v1/messages/${this.currentRoom}/recent?limit=20`);
            const result = await response.json();
            
            if (response.ok) {
                result.messages.reverse().forEach(message => {
                    this.addMessage(message);
                });
            }
        } catch (error) {
            console.error('Error loading messages:', error);
        }
    }

    clearMessages() {
        if (confirm('Bạn có chắc muốn xóa tất cả tin nhắn?')) {
            document.getElementById('messages').innerHTML = '';
            this.showNotification('Đã xóa tin nhắn', 'success');
        }
    }

    searchMessages() {
        const query = prompt('Nhập từ khóa tìm kiếm:');
        if (query) {
            this.performSearch(query);
        }
    }

    async performSearch(query) {
        try {
            const response = await fetch(`/api/v1/messages/${this.currentRoom}/search?q=${encodeURIComponent(query)}`);
            const result = await response.json();
            
            if (response.ok) {
                this.showNotification(`Tìm thấy ${result.messages.length} tin nhắn`, 'success');
                // In a real app, you might want to display search results in a separate modal
                console.log('Search results:', result.messages);
            }
        } catch (error) {
            console.error('Search error:', error);
        }
    }

    logout() {
        if (confirm('Bạn có chắc muốn đăng xuất?')) {
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            
            if (this.ws) {
                this.ws.close();
            }
            
            this.currentUser = null;
            this.showLoginModal();
            this.showNotification('Đã đăng xuất', 'success');
        }
    }

    showLoginModal() {
        document.getElementById('loginModal').style.display = 'flex';
        document.getElementById('registerModal').style.display = 'none';
        document.getElementById('chatApp').style.display = 'none';
    }

    showRegisterModal() {
        document.getElementById('loginModal').style.display = 'none';
        document.getElementById('registerModal').style.display = 'flex';
    }

    showChatInterface() {
        document.getElementById('loginModal').style.display = 'none';
        document.getElementById('registerModal').style.display = 'none';
        document.getElementById('chatApp').style.display = 'flex';
        
        document.getElementById('currentUser').textContent = this.currentUser.username;
    }

    showLoading(show) {
        document.getElementById('loadingOverlay').style.display = show ? 'flex' : 'none';
    }

    showNotification(message, type = 'success') {
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.textContent = message;
        
        document.getElementById('notifications').appendChild(notification);
        
        // Auto remove after 3 seconds
        setTimeout(() => {
            notification.remove();
        }, 3000);
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.chatApp = new ChatApp();
});

// Global functions for HTML onclick handlers
function showLoginModal() {
    window.chatApp.showLoginModal();
}

function showRegisterModal() {
    window.chatApp.showRegisterModal();
}

function logout() {
    window.chatApp.logout();
}

function clearMessages() {
    window.chatApp.clearMessages();
}

function searchMessages() {
    window.chatApp.searchMessages();
}
