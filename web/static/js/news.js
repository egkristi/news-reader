// News display component
class NewsDisplay {
    constructor(containerId) {
        this.container = document.getElementById(containerId);
        this.allNews = [];

        // Wire up filter controls
        const searchInput = document.getElementById('searchInput');
        const categoryFilter = document.getElementById('categoryFilter');
        const sourceFilter = document.getElementById('sourceFilter');

        if (searchInput) searchInput.addEventListener('input', () => this.filterNews());
        if (categoryFilter) categoryFilter.addEventListener('change', () => this.filterNews());
        if (sourceFilter) sourceFilter.addEventListener('change', () => this.filterNews());
    }

    async loadNews() {
        this.container.innerHTML = '<div class="loading">Loading news...</div>';
        try {
            this.allNews = await API.get('/api/news');
            this.populateFilters();
            this.filterNews();
        } catch (err) {
            this.container.innerHTML = '<div class="error">Failed to load news. Please try again.</div>';
            console.error('Error loading news:', err);
        }
    }

    populateFilters() {
        const categories = [...new Set(this.allNews.map(n => n.category).filter(Boolean))].sort();
        const sources = [...new Set(this.allNews.map(n => n.source).filter(Boolean))].sort();

        const categoryFilter = document.getElementById('categoryFilter');
        const sourceFilter = document.getElementById('sourceFilter');

        if (categoryFilter) {
            categoryFilter.innerHTML = '<option value="">All Categories</option>' +
                categories.map(c => `<option value="${c}">${c}</option>`).join('');
        }
        if (sourceFilter) {
            sourceFilter.innerHTML = '<option value="">All Sources</option>' +
                sources.map(s => `<option value="${s}">${s}</option>`).join('');
        }
    }

    filterNews() {
        const search = (document.getElementById('searchInput')?.value || '').toLowerCase();
        const category = document.getElementById('categoryFilter')?.value || '';
        const source = document.getElementById('sourceFilter')?.value || '';

        const filtered = this.allNews.filter(item => {
            if (category && item.category !== category) return false;
            if (source && item.source !== source) return false;
            if (search) {
                const text = (item.title + ' ' + item.description + ' ' + item.source).toLowerCase();
                if (!text.includes(search)) return false;
            }
            return true;
        });

        this.renderNews(filtered);
    }

    escapeHtml(text) {
        if (!text) return '';
        const el = document.createElement('div');
        el.textContent = text;
        return el.innerHTML;
    }

    stripHtml(html) {
        if (!html) return '';
        const tmp = document.createElement('div');
        tmp.innerHTML = html;
        return tmp.textContent || tmp.innerText || '';
    }

    truncate(text, maxLen) {
        if (!text) return '';
        const clean = this.stripHtml(text);
        return clean.length > maxLen ? clean.substring(0, maxLen) + '...' : clean;
    }

    formatDate(dateStr) {
        const date = new Date(dateStr);
        if (isNaN(date)) return '';
        const now = new Date();
        const diffMs = now - date;
        const diffHrs = Math.floor(diffMs / 3600000);
        if (diffHrs < 1) return 'Just now';
        if (diffHrs < 24) return diffHrs + 'h ago';
        const diffDays = Math.floor(diffHrs / 24);
        if (diffDays < 7) return diffDays + 'd ago';
        return date.toLocaleDateString();
    }

    renderTags(tags) {
        if (!tags || tags.length === 0) return '';
        return '<div class="news-tags">' +
            tags.map(tag =>
                `<span class="news-tag" style="background:${this.escapeHtml(tag.color)}">${this.escapeHtml(tag.name)}</span>`
            ).join('') +
            '</div>';
    }

    contentIcon(type) {
        switch (type) {
            case 'video': return '🎬';
            case 'podcast': return '🎧';
            default: return '📰';
        }
    }

    renderNews(items) {
        if (!items || items.length === 0) {
            this.container.innerHTML = '<div class="empty-state">No news articles found.</div>';
            return;
        }

        this.container.innerHTML = items.map(item => `
            <article class="news-card">
                ${item.thumbnail ? `<img class="news-thumb" src="${this.escapeHtml(item.thumbnail)}" alt="" loading="lazy" onerror="this.style.display='none'">` : ''}
                <div class="news-body">
                    <div class="news-meta">
                        <span class="news-type">${this.contentIcon(item.contentType)}</span>
                        <span class="news-source">${this.escapeHtml(item.source)}</span>
                        <span class="news-category">${this.escapeHtml(item.category)}</span>
                        <span class="news-date">${this.formatDate(item.published)}</span>
                    </div>
                    <h3 class="news-title">
                        <a href="${this.escapeHtml(item.link)}" target="_blank" rel="noopener noreferrer">${this.escapeHtml(item.title)}</a>
                    </h3>
                    <p class="news-desc">${this.escapeHtml(this.truncate(item.description, 200))}</p>
                    ${this.renderTags(item.tags)}
                    ${item.audioUrl ? `<audio controls preload="none" class="news-audio"><source src="${this.escapeHtml(item.audioUrl)}"></audio>` : ''}
                </div>
            </article>
        `).join('');
    }
}
