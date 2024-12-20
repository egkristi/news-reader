// Function to fetch and display news sources
async function loadNewsSources() {
    try {
        const response = await fetch('/api/preferences');
        const preferences = await response.json();
        displayNewsSources(preferences.sources);
    } catch (error) {
        console.error('Error loading news sources:', error);
    }
}

// Function to display news sources with checkboxes
function displayNewsSources(sources) {
    const sourcesContainer = document.getElementById('sources-container');
    if (!sourcesContainer) return;

    // Group sources by category
    const sourcesByCategory = sources.reduce((acc, source) => {
        if (!acc[source.category]) {
            acc[source.category] = [];
        }
        acc[source.category].push(source);
        return acc;
    }, {});

    // Clear existing content
    sourcesContainer.innerHTML = '';

    // Create category sections
    Object.entries(sourcesByCategory).forEach(([category, categorySources]) => {
        const categorySection = document.createElement('div');
        categorySection.className = 'source-category';
        
        const categoryHeader = document.createElement('h3');
        categoryHeader.textContent = category;
        categorySection.appendChild(categoryHeader);

        // Add sources for this category
        categorySources.forEach(source => {
            const sourceDiv = document.createElement('div');
            sourceDiv.className = 'source-item';

            const checkbox = document.createElement('input');
            checkbox.type = 'checkbox';
            checkbox.id = `source-${source.name.replace(/\s+/g, '-')}`;
            checkbox.checked = source.enabled;
            checkbox.addEventListener('change', () => updateSourceStatus(source.name, checkbox.checked));

            const label = document.createElement('label');
            label.htmlFor = checkbox.id;
            label.textContent = source.name;

            sourceDiv.appendChild(checkbox);
            sourceDiv.appendChild(label);
            categorySection.appendChild(sourceDiv);
        });

        sourcesContainer.appendChild(categorySection);
    });
}

// Function to update source enabled status
async function updateSourceStatus(sourceName, enabled) {
    try {
        const response = await fetch('/api/preferences');
        const preferences = await response.json();
        
        // Update the enabled status of the specified source
        preferences.sources = preferences.sources.map(source => {
            if (source.name === sourceName) {
                return { ...source, enabled };
            }
            return source;
        });

        // Save updated preferences
        await fetch('/api/preferences', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(preferences),
        });

        // Refresh news after updating sources
        loadNews();
    } catch (error) {
        console.error('Error updating source status:', error);
    }
}

// Load sources when the page loads
document.addEventListener('DOMContentLoaded', loadNewsSources);
