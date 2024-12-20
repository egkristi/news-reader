class VersionInfo {
    constructor(containerId) {
        this.container = document.getElementById(containerId);
        this.fetchAndDisplay();
    }

    async fetchVersion() {
        try {
            const response = await fetch('/api/version');
            if (!response.ok) {
                throw new Error('Failed to fetch version info');
            }
            return await response.json();
        } catch (error) {
            console.error('Error fetching version:', error);
            return null;
        }
    }

    formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleString();
    }

    async fetchAndDisplay() {
        const data = await this.fetchVersion();
        if (!data) {
            this.container.innerHTML = '<span class="error">Version info unavailable</span>';
            return;
        }

        this.container.innerHTML = `
            <span>v${data.version}</span>
            <span class="git-commit">${data.gitCommit}</span>
            <span>Built: ${this.formatDate(data.buildTime)}</span>
        `;
    }
}
