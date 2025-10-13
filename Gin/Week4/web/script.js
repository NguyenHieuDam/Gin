// News Aggregator Frontend JavaScript

class NewsApp {
    constructor() {
        this.apiBaseUrl = 'http://localhost:8080/api/v1';
        this.collectorUrl = 'http://localhost:8081';
        this.currentPage = 1;
        this.currentCategory = 'all';
        this.articlesPerPage = 6;
        this.allArticles = [];
        this.filteredArticles = [];
        
        this.init();
    }

    init() {
        this.bindEvents();
        this.checkApiStatus();
        this.loadNews();
        this.startAutoRefresh();
    }

    bindEvents() {
        // Navigation buttons
        document.querySelectorAll('.nav-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                this.setActiveCategory(e.target.dataset.category);
            });
        });

        // Search functionality
        document.getElementById('searchInput').addEventListener('input', (e) => {
            this.filterArticles(e.target.value);
        });

        // Refresh button
        document.getElementById('refreshBtn').addEventListener('click', () => {
            this.loadNews();
        });

        // Collect news button
        document.getElementById('collectBtn').addEventListener('click', () => {
            this.collectNews();
        });

        // Pagination
        document.getElementById('prevBtn').addEventListener('click', () => {
            this.previousPage();
        });

        document.getElementById('nextBtn').addEventListener('click', () => {
            this.nextPage();
        });

        // Modal
        document.getElementById('closeModal').addEventListener('click', () => {
            this.closeModal();
        });

        document.getElementById('articleModal').addEventListener('click', (e) => {
            if (e.target.id === 'articleModal') {
                this.closeModal();
            }
        });

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.closeModal();
            }
        });
    }

    async checkApiStatus() {
        try {
            const response = await fetch(`${this.apiBaseUrl.replace('/api/v1', '')}/health`);
            const data = await response.json();
            
            const statusElement = document.getElementById('apiStatus');
            if (response.ok) {
                statusElement.textContent = 'Hoạt động';
                statusElement.style.color = '#4CAF50';
            } else {
                statusElement.textContent = 'Lỗi';
                statusElement.style.color = '#f44336';
            }
        } catch (error) {
            document.getElementById('apiStatus').textContent = 'Không kết nối được';
            document.getElementById('apiStatus').style.color = '#f44336';
        }
    }

    async loadNews() {
        this.showLoading(true);
        
        try {
            const response = await fetch(`${this.apiBaseUrl}/news?limit=100&offset=0`);
            const data = await response.json();
            
            if (response.ok) {
                this.allArticles = data.articles || [];
                this.filteredArticles = [...this.allArticles];
                this.updateStats();
                this.renderNews();
            } else {
                this.showError('Không thể tải tin tức');
            }
        } catch (error) {
            console.error('Error loading news:', error);
            this.showError('Lỗi kết nối đến server');
        } finally {
            this.showLoading(false);
        }
    }

    async collectNews() {
        const collectBtn = document.getElementById('collectBtn');
        const originalText = collectBtn.innerHTML;
        
        collectBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Đang thu thập...';
        collectBtn.disabled = true;

        try {
            const response = await fetch(`${this.collectorUrl}/collect`, {
                method: 'POST'
            });
            
            if (response.ok) {
                this.showNotification('Thu thập tin tức thành công!', 'success');
                // Reload news after collection
                setTimeout(() => {
                    this.loadNews();
                }, 1000);
            } else {
                this.showNotification('Lỗi khi thu thập tin tức', 'error');
            }
        } catch (error) {
            console.error('Error collecting news:', error);
            this.showNotification('Lỗi kết nối đến collector', 'error');
        } finally {
            collectBtn.innerHTML = originalText;
            collectBtn.disabled = false;
        }
    }

    setActiveCategory(category) {
        this.currentCategory = category;
        this.currentPage = 1;
        
        // Update active button
        document.querySelectorAll('.nav-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        document.querySelector(`[data-category="${category}"]`).classList.add('active');
        
        // Filter articles
        this.filterArticles(document.getElementById('searchInput').value);
    }

    filterArticles(searchTerm) {
        let filtered = [...this.allArticles];
        
        // Filter by category
        if (this.currentCategory !== 'all') {
            filtered = filtered.filter(article => 
                article.category === this.currentCategory
            );
        }
        
        // Filter by search term
        if (searchTerm.trim()) {
            const term = searchTerm.toLowerCase();
            filtered = filtered.filter(article => 
                article.title.toLowerCase().includes(term) ||
                article.content.toLowerCase().includes(term) ||
                article.source.toLowerCase().includes(term)
            );
        }
        
        this.filteredArticles = filtered;
        this.currentPage = 1;
        this.renderNews();
    }

    renderNews() {
        const newsGrid = document.getElementById('newsGrid');
        const noNews = document.getElementById('noNews');
        
        if (this.filteredArticles.length === 0) {
            newsGrid.style.display = 'none';
            noNews.style.display = 'block';
            this.updatePagination();
            return;
        }
        
        newsGrid.style.display = 'grid';
        noNews.style.display = 'none';
        
        // Calculate pagination
        const startIndex = (this.currentPage - 1) * this.articlesPerPage;
        const endIndex = startIndex + this.articlesPerPage;
        const articlesToShow = this.filteredArticles.slice(startIndex, endIndex);
        
        newsGrid.innerHTML = articlesToShow.map(article => this.createNewsCard(article)).join('');
        
        this.updatePagination();
    }

    createNewsCard(article) {
        const publishedDate = new Date(article.published_at).toLocaleDateString('vi-VN');
        const categoryClass = article.category || 'default';
        
        return `
            <div class="news-card" onclick="app.openArticle('${article.id}')">
                <div class="news-card-header">
                    <div class="news-card-meta">
                        <span class="category-badge ${categoryClass}">${this.getCategoryName(article.category)}</span>
                        <span class="source">${article.source}</span>
                    </div>
                    <h3>${article.title}</h3>
                </div>
                <div class="news-card-content">
                    <p>${this.truncateText(article.content, 150)}</p>
                </div>
                <div class="news-card-footer">
                    <span class="date">${publishedDate}</span>
                    <span class="read-more">Đọc thêm →</span>
                </div>
            </div>
        `;
    }

    async openArticle(articleId) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/news/${articleId}`);
            const article = await response.json();
            
            if (response.ok) {
                this.showArticleModal(article);
            } else {
                this.showNotification('Không tìm thấy bài viết', 'error');
            }
        } catch (error) {
            console.error('Error loading article:', error);
            this.showNotification('Lỗi khi tải bài viết', 'error');
        }
    }

    showArticleModal(article) {
        const modal = document.getElementById('articleModal');
        const publishedDate = new Date(article.published_at).toLocaleString('vi-VN');
        
        document.getElementById('modalTitle').textContent = article.title;
        document.getElementById('modalSource').textContent = article.source;
        document.getElementById('modalCategory').textContent = this.getCategoryName(article.category);
        document.getElementById('modalDate').textContent = publishedDate;
        document.getElementById('modalContent').textContent = article.content;
        document.getElementById('modalUrl').href = article.url;
        
        modal.style.display = 'block';
        document.body.style.overflow = 'hidden';
    }

    closeModal() {
        const modal = document.getElementById('articleModal');
        modal.style.display = 'none';
        document.body.style.overflow = 'auto';
    }

    updatePagination() {
        const totalPages = Math.ceil(this.filteredArticles.length / this.articlesPerPage);
        const prevBtn = document.getElementById('prevBtn');
        const nextBtn = document.getElementById('nextBtn');
        const pageInfo = document.getElementById('pageInfo');
        
        prevBtn.disabled = this.currentPage === 1;
        nextBtn.disabled = this.currentPage === totalPages || totalPages === 0;
        
        pageInfo.textContent = `Trang ${this.currentPage} / ${totalPages || 1}`;
    }

    previousPage() {
        if (this.currentPage > 1) {
            this.currentPage--;
            this.renderNews();
        }
    }

    nextPage() {
        const totalPages = Math.ceil(this.filteredArticles.length / this.articlesPerPage);
        if (this.currentPage < totalPages) {
            this.currentPage++;
            this.renderNews();
        }
    }

    updateStats() {
        document.getElementById('totalArticles').textContent = this.allArticles.length;
        document.getElementById('lastUpdate').textContent = new Date().toLocaleTimeString('vi-VN');
    }

    showLoading(show) {
        const loading = document.getElementById('loading');
        const newsGrid = document.getElementById('newsGrid');
        
        if (show) {
            loading.style.display = 'flex';
            newsGrid.style.display = 'none';
        } else {
            loading.style.display = 'none';
            newsGrid.style.display = 'grid';
        }
    }

    showError(message) {
        const newsGrid = document.getElementById('newsGrid');
        newsGrid.innerHTML = `
            <div style="grid-column: 1 / -1; text-align: center; padding: 2rem; color: white;">
                <i class="fas fa-exclamation-triangle" style="font-size: 3rem; margin-bottom: 1rem;"></i>
                <h3>${message}</h3>
                <p>Vui lòng thử lại sau</p>
            </div>
        `;
    }

    showNotification(message, type = 'info') {
        // Create notification element
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.innerHTML = `
            <i class="fas fa-${type === 'success' ? 'check-circle' : type === 'error' ? 'exclamation-circle' : 'info-circle'}"></i>
            <span>${message}</span>
        `;
        
        // Add styles
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: ${type === 'success' ? '#4CAF50' : type === 'error' ? '#f44336' : '#2196F3'};
            color: white;
            padding: 15px 20px;
            border-radius: 10px;
            box-shadow: 0 5px 15px rgba(0,0,0,0.2);
            z-index: 1001;
            display: flex;
            align-items: center;
            gap: 10px;
            animation: slideIn 0.3s ease;
        `;
        
        document.body.appendChild(notification);
        
        // Remove after 3 seconds
        setTimeout(() => {
            notification.style.animation = 'slideOut 0.3s ease';
            setTimeout(() => {
                document.body.removeChild(notification);
            }, 300);
        }, 3000);
    }

    startAutoRefresh() {
        // Auto refresh every 5 minutes
        setInterval(() => {
            this.loadNews();
        }, 5 * 60 * 1000);
    }

    getCategoryName(category) {
        const categoryNames = {
            'technology': 'Công nghệ',
            'business': 'Kinh doanh',
            'sports': 'Thể thao',
            'default': 'Khác'
        };
        return categoryNames[category] || 'Khác';
    }

    truncateText(text, maxLength) {
        if (text.length <= maxLength) return text;
        return text.substring(0, maxLength) + '...';
    }
}

// Add CSS animations
const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from {
            transform: translateX(100%);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }
    
    @keyframes slideOut {
        from {
            transform: translateX(0);
            opacity: 1;
        }
        to {
            transform: translateX(100%);
            opacity: 0;
        }
    }
`;
document.head.appendChild(style);

// Initialize app when DOM is loaded
let app;
document.addEventListener('DOMContentLoaded', () => {
    app = new NewsApp();
});
