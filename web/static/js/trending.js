// Trending topics functionality
class TrendingTopics {
    constructor(containerId) {
        this.container = document.getElementById(containerId);
        this.updateInterval = 5 * 60 * 1000; // 5 minutes
    }

    async fetchTrendingTopics() {
        try {
            const response = await fetch('/api/news/trending');
            if (!response.ok) {
                throw new Error('Failed to fetch trending topics');
            }
            return await response.json();
        } catch (error) {
            console.error('Error fetching trending topics:', error);
            return null;
        }
    }

    createTopicElement(topic) {
        const element = document.createElement('div');
        element.className = 'trending-topic';
        
        // Calculate size based on frequency (between 0.8em and 1.5em)
        const minSize = 0.8;
        const maxSize = 1.5;
        const size = minSize + (topic.frequency / 10) * (maxSize - minSize);
        
        element.style.fontSize = `${size}em`;
        element.innerHTML = `
            <span class="topic-text">${topic.topic}</span>
            <span class="topic-frequency">${topic.frequency}</span>
        `;
        return element;
    }

    async render() {
        const data = await this.fetchTrendingTopics();
        if (!data || !data.topics) {
            this.container.innerHTML = '<div class="error">No trending topics available</div>';
            return;
        }

        // Clear existing content
        this.container.innerHTML = '';

        // Create header
        const header = document.createElement('div');
        header.className = 'trending-header';
        header.innerHTML = `
            <h3>Trending Topics</h3>
            <span class="update-time">Updated: ${new Date(data.time).toLocaleTimeString()}</span>
        `;
        this.container.appendChild(header);

        // Create topics cloud
        const cloud = document.createElement('div');
        cloud.className = 'trending-cloud';
        
        data.topics.forEach(topic => {
            const topicElement = this.createTopicElement(topic);
            cloud.appendChild(topicElement);
            
            // Add click handler to filter news by topic
            topicElement.addEventListener('click', () => {
                // Dispatch custom event for news filtering
                const event = new CustomEvent('filter-by-topic', {
                    detail: { topic: topic.topic }
                });
                document.dispatchEvent(event);
            });
        });

        this.container.appendChild(cloud);
    }

    startAutoUpdate() {
        // Initial render
        this.render();

        // Set up auto-update
        setInterval(() => this.render(), this.updateInterval);
    }
}
