/**
 * Enhanced Blog JavaScript with Search, Pagination, and HTMX Integration
 * Implements the features described in WEEK2_INFRASTRUCTURE.md
 */

// Blog Search Class with Debouncing and FTS5 Integration
// Prevent redeclaration if script is loaded multiple times
if (typeof window.BlogSearch === 'undefined') {
window.BlogSearch = class BlogSearch {
    constructor(options = {}) {
        this.options = Object.assign({
            searchInputId: 'blog-search',
            resultsContainerId: 'blog-posts',
            loadingClass: 'search-loading',
            noResultsClass: 'no-results',
            debounceMs: 300,
            minQueryLength: 2,
            apiEndpoint: '/api/blog/search'
        }, options);
        
        this.searchInput = document.getElementById(this.options.searchInputId);
        this.resultsContainer = document.getElementById(this.options.resultsContainerId);
        this.currentQuery = '';
        this.searchTimeout = null;
        this.isSearching = false;
        
        this.init();
    }
    
    init() {
        if (!this.searchInput || !this.resultsContainer) {
            console.warn('Blog search elements not found');
            return;
        }
        
        // Add event listeners
        this.searchInput.addEventListener('input', (e) => this.handleSearchInput(e));
        this.searchInput.addEventListener('keydown', (e) => this.handleKeydown(e));
        
        // Handle initial search query from URL
        const urlParams = new URLSearchParams(window.location.search);
        const initialQuery = urlParams.get('search');
        if (initialQuery) {
            this.searchInput.value = initialQuery;
            this.performSearch(initialQuery, false);
        }
        
        console.log('Blog search initialized');
    }
    
    handleSearchInput(e) {
        const query = e.target.value.trim();
        
        // Clear search if query too short
        if (query.length < this.options.minQueryLength) {
            this.clearSearch();
            return;
        }
        
        // Debounce search requests
        clearTimeout(this.searchTimeout);
        this.searchTimeout = setTimeout(() => {
            this.performSearch(query);
        }, this.options.debounceMs);
    }
    
    handleKeydown(e) {
        if (e.key === 'Escape') {
            this.clearSearch();
        }
    }
    
    async performSearch(query, updateUrl = true) {
        if (query === this.currentQuery || this.isSearching) return;
        
        this.currentQuery = query;
        this.isSearching = true;
        this.showLoading();
        
        try {
            const response = await fetch(
                `${this.options.apiEndpoint}?q=${encodeURIComponent(query)}&page=1`
            );
            
            if (!response.ok) {
                throw new Error(`Search request failed: ${response.status}`);
            }
            
            const data = await response.json();
            this.displayResults(data, query);
            
            if (updateUrl) {
                this.updateURL(query);
            }
            
        } catch (error) {
            console.error('Search failed:', error);
            this.showError('Search failed. Please try again.');
        } finally {
            this.isSearching = false;
        }
    }
    
    displayResults(data, query) {
        if (!data.posts || data.posts.length === 0) {
            this.showNoResults(query);
            return;
        }
        
        // Use HTMX to update results with search context
        const searchUrl = `/content/blog?search=${encodeURIComponent(query)}`;
        
        // Check if we're on the blog page with the proper container
        const blogContainer = document.getElementById('blog-posts-container');
        const mainContent = document.getElementById('main-content');
        
        // Update via HTMX
        if (blogContainer) {
            htmx.ajax('GET', searchUrl, {
                target: '#blog-posts-container',
                swap: 'innerHTML'
            });
        } else if (mainContent) {
            htmx.ajax('GET', searchUrl, {
                target: '#main-content',
                swap: 'innerHTML'
            });
        }
        
        // Update result count indicator
        this.updateResultCount(data.total, query);
    }
    
    showLoading() {
        this.resultsContainer.innerHTML = `
            <div class="${this.options.loadingClass}">
                <div class="loading-spinner"></div>
                <p>Searching...</p>
            </div>
        `;
    }
    
    showNoResults(query) {
        this.resultsContainer.innerHTML = `
            <div class="${this.options.noResultsClass}">
                <h3>No posts found</h3>
                <p>No posts match "${this.escapeHtml(query)}". Try different keywords.</p>
                <button class="btn btn-secondary" onclick="blogFeatures.search.clearSearch()">
                    View All Posts
                </button>
            </div>
        `;
    }
    
    showError(message) {
        this.resultsContainer.innerHTML = `
            <div class="alert alert-error">
                <h3>Search Error</h3>
                <p>${this.escapeHtml(message)}</p>
                <button class="btn btn-secondary" onclick="blogFeatures.search.clearSearch()">
                    View All Posts
                </button>
            </div>
        `;
    }
    
    clearSearch() {
        this.searchInput.value = '';
        this.currentQuery = '';
        this.isSearching = false;
        
        // Check if we're on the blog page with the proper container
        const blogContainer = document.getElementById('blog-posts-container');
        const mainContent = document.getElementById('main-content');
        
        // Load first page of all posts
        if (blogContainer) {
            htmx.ajax('GET', '/content/blog', {
                target: '#blog-posts-container',
                swap: 'innerHTML'
            });
        } else if (mainContent) {
            htmx.ajax('GET', '/content/blog', {
                target: '#main-content',
                swap: 'innerHTML'
            });
        }
        
        // Clear URL search parameter
        const url = new URL(window.location);
        url.searchParams.delete('search');
        window.history.pushState({}, '', url);
        
        this.clearResultCount();
    }
    
    updateURL(query) {
        const url = new URL(window.location);
        url.searchParams.set('search', query);
        url.searchParams.delete('page'); // Reset to first page
        window.history.pushState({}, '', url);
    }
    
    updateResultCount(count, query) {
        let indicator = document.getElementById('search-result-count');
        if (!indicator) {
            indicator = document.createElement('div');
            indicator.id = 'search-result-count';
            indicator.className = 'search-result-indicator';
            this.searchInput.parentNode.appendChild(indicator);
        }
        
        indicator.innerHTML = `
            Found ${count} post${count !== 1 ? 's' : ''} for "${this.escapeHtml(query)}"
            <button onclick="blogFeatures.search.clearSearch()" aria-label="Clear search">×</button>
        `;
    }
    
    clearResultCount() {
        const indicator = document.getElementById('search-result-count');
        if (indicator) {
            indicator.remove();
        }
    }
    
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
};
}

// Blog Pagination Class with HTMX Integration
if (typeof window.BlogPagination === 'undefined') {
window.BlogPagination = class BlogPagination {
    constructor(options = {}) {
        this.options = Object.assign({
            apiEndpoint: '/api/blog/posts',
            containerClass: 'pagination-container',
            currentPageClass: 'current-page',
            disabledClass: 'disabled'
        }, options);
        
        this.currentPage = 1;
        this.totalPages = 1;
        this.currentFilters = {};
        
        this.init();
    }
    
    init() {
        this.extractCurrentState();
        this.bindEvents();
        console.log('Blog pagination initialized');
    }
    
    extractCurrentState() {
        const urlParams = new URLSearchParams(window.location.search);
        this.currentPage = parseInt(urlParams.get('page')) || 1;
        this.currentFilters = {
            category: urlParams.get('category') || '',
            tag: urlParams.get('tag') || '',
            search: urlParams.get('search') || ''
        };
    }
    
    bindEvents() {
        // Delegate pagination clicks
        document.addEventListener('click', (e) => {
            if (e.target.matches('.pagination-link[data-page]')) {
                e.preventDefault();
                const page = parseInt(e.target.getAttribute('data-page'));
                this.navigateToPage(page);
            }
        });
        
        // Handle browser back/forward
        window.addEventListener('popstate', () => {
            this.extractCurrentState();
            this.loadPage(this.currentPage, false);
        });
    }
    
    navigateToPage(page) {
        if (page === this.currentPage || page < 1) return;
        
        this.currentPage = page;
        this.loadPage(page, true);
    }
    
    async loadPage(page, updateUrl = true) {
        const params = new URLSearchParams({
            page: page.toString(),
            per_page: '12',
            ...this.currentFilters
        });
        
        const url = `/content/blog?${params.toString()}`;
        
        try {
            // Check if we're on the blog page with the proper container
            const blogContainer = document.getElementById('blog-posts-container');
            const mainContent = document.getElementById('main-content');
            
            // Use HTMX for seamless updates
            if (blogContainer) {
                // We're on the blog page, update just the posts container
                htmx.ajax('GET', url, {
                    target: '#blog-posts-container',
                    swap: 'innerHTML',
                    headers: {
                        'HX-Request': 'true'
                    }
                });
            } else if (mainContent) {
                // We're elsewhere (like a blog post), update the main content
                htmx.ajax('GET', url, {
                    target: '#main-content',
                    swap: 'innerHTML',
                    headers: {
                        'HX-Request': 'true'
                    }
                });
            }
            
            if (updateUrl) {
                this.updateURL(page);
            }
            
        } catch (error) {
            console.error('Pagination failed:', error);
        }
    }
    
    updateURL(page) {
        const url = new URL(window.location);
        if (page > 1) {
            url.searchParams.set('page', page.toString());
        } else {
            url.searchParams.delete('page');
        }
        window.history.pushState({}, '', url);
    }
};
}

// Social Sharing Class
if (typeof window.SocialSharing === 'undefined') {
window.SocialSharing = class SocialSharing {
    constructor(options = {}) {
        this.options = Object.assign({
            services: ['twitter', 'linkedin', 'facebook', 'copy'],
            copySuccessMessage: 'Link copied to clipboard!',
            copyErrorMessage: 'Failed to copy link'
        }, options);
        
        this.init();
    }
    
    init() {
        this.bindEvents();
        console.log('Social sharing initialized');
    }
    
    bindEvents() {
        document.addEventListener('click', (e) => {
            if (e.target.matches('[data-share]')) {
                e.preventDefault();
                const service = e.target.getAttribute('data-share');
                const url = e.target.getAttribute('data-url') || window.location.href;
                const title = e.target.getAttribute('data-title') || document.title;
                
                this.share(service, url, title);
            }
        });
    }
    
    share(service, url, title) {
        const encodedUrl = encodeURIComponent(url);
        const encodedTitle = encodeURIComponent(title);
        
        const services = {
            twitter: `https://twitter.com/intent/tweet?url=${encodedUrl}&text=${encodedTitle}`,
            linkedin: `https://www.linkedin.com/sharing/share-offsite/?url=${encodedUrl}`,
            facebook: `https://www.facebook.com/sharer/sharer.php?u=${encodedUrl}`,
            copy: null // Special handling below
        };
        
        if (service === 'copy') {
            this.copyToClipboard(url);
        } else if (services[service]) {
            window.open(services[service], '_blank', 'width=600,height=400');
        }
    }
    
    async copyToClipboard(text) {
        try {
            await navigator.clipboard.writeText(text);
            this.showCopySuccess();
        } catch (err) {
            console.error('Copy failed:', err);
            this.showCopyError();
        }
    }
    
    showCopySuccess() {
        this.showToast(this.options.copySuccessMessage, 'success');
    }
    
    showCopyError() {
        this.showToast(this.options.copyErrorMessage, 'error');
    }
    
    showToast(message, type = 'info') {
        // Create toast element
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        toast.textContent = message;
        
        // Style the toast
        Object.assign(toast.style, {
            position: 'fixed',
            top: '20px',
            right: '20px',
            padding: '12px 20px',
            borderRadius: '4px',
            color: 'white',
            backgroundColor: type === 'success' ? '#00ff88' : '#ff6b6b',
            zIndex: '9999',
            opacity: '0',
            transform: 'translateY(-20px)',
            transition: 'all 0.3s ease'
        });
        
        // Add to DOM
        document.body.appendChild(toast);
        
        // Animate in
        requestAnimationFrame(() => {
            toast.style.opacity = '1';
            toast.style.transform = 'translateY(0)';
        });
        
        // Remove after delay
        setTimeout(() => {
            toast.style.opacity = '0';
            toast.style.transform = 'translateY(-20px)';
            setTimeout(() => {
                if (toast.parentNode) {
                    document.body.removeChild(toast);
                }
            }, 300);
        }, 3000);
    }
};
}

// Keyboard Navigation Class
if (typeof window.KeyboardNavigation === 'undefined') {
window.KeyboardNavigation = class KeyboardNavigation {
    constructor() {
        this.init();
    }
    
    init() {
        document.addEventListener('keydown', (e) => this.handleKeyPress(e));
        console.log('Keyboard navigation initialized');
    }
    
    handleKeyPress(e) {
        // Only handle keys when not in form inputs
        if (e.target.matches('input, textarea, select')) return;
        
        switch (e.key) {
            case 'j':
            case 'ArrowDown':
                e.preventDefault();
                this.navigateNext();
                break;
            case 'k':
            case 'ArrowUp':
                e.preventDefault();
                this.navigatePrevious();
                break;
            case '/':
                e.preventDefault();
                this.focusSearch();
                break;
            case 'Escape':
                this.clearFocus();
                break;
        }
    }
    
    navigateNext() {
        const nextLink = document.querySelector('.post-nav-next');
        if (nextLink) {
            nextLink.click();
        }
    }
    
    navigatePrevious() {
        const prevLink = document.querySelector('.post-nav-prev');
        if (prevLink) {
            prevLink.click();
        }
    }
    
    focusSearch() {
        const searchInput = document.getElementById('blog-search');
        if (searchInput) {
            searchInput.focus();
        }
    }
    
    clearFocus() {
        document.activeElement.blur();
    }
};
}

// Reading Time Calculator
if (typeof window.ReadingTimeCalculator === 'undefined') {
window.ReadingTimeCalculator = class ReadingTimeCalculator {
    constructor(options = {}) {
        this.options = Object.assign({
            wordsPerMinute: 225,
            selector: '[data-reading-time]'
        }, options);
        
        this.init();
    }
    
    init() {
        this.updateReadingTimes();
        console.log('Reading time calculator initialized');
    }
    
    updateReadingTimes() {
        const elements = document.querySelectorAll(this.options.selector);
        
        elements.forEach(element => {
            const content = element.getAttribute('data-content') || element.textContent;
            const readingTime = this.calculateReadingTime(content);
            
            const display = element.querySelector('.reading-time-display') || element;
            display.textContent = `${readingTime} min read`;
        });
    }
    
    calculateReadingTime(text) {
        // Strip HTML tags
        const plainText = text.replace(/<[^>]*>/g, '');
        const wordCount = plainText.trim().split(/\s+/).length;
        const minutes = Math.ceil(wordCount / this.options.wordsPerMinute);
        
        return Math.max(1, minutes);
    }
};
}

// Initialize all blog features when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    // Check if we're on a blog page
    const isBlogPage = document.body.classList.contains('page-blog') || 
                      document.querySelector('[data-page="blog"]') ||
                      window.location.pathname.startsWith('/blog');
    
    if (!isBlogPage) {
        console.log('Not on blog page, skipping blog features initialization');
        return;
    }
    
    console.log('Initializing blog features...');
    
    // Initialize features
    window.blogFeatures = {
        search: new window.BlogSearch(),
        pagination: new window.BlogPagination(),
        sharing: new window.SocialSharing(),
        keyboard: new window.KeyboardNavigation(),
        readingTime: new window.ReadingTimeCalculator()
    };
    
    console.log('All blog features initialized');
    
    // Emit custom event for other scripts
    document.dispatchEvent(new CustomEvent('blogFeaturesReady', {
        detail: window.blogFeatures
    }));
});

// HTMX integration improvements
document.addEventListener('htmx:afterSwap', (e) => {
    // Re-initialize reading time calculation after HTMX swaps
    if (window.blogFeatures && window.blogFeatures.readingTime) {
        window.blogFeatures.readingTime.updateReadingTimes();
    }
    
    // Update pagination state
    if (window.blogFeatures && window.blogFeatures.pagination) {
        window.blogFeatures.pagination.extractCurrentState();
    }
});

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        BlogSearch,
        BlogPagination,
        SocialSharing,
        KeyboardNavigation,
        ReadingTimeCalculator
    };
}