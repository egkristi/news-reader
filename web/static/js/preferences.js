// Preferences modal component
class Preferences {
    constructor(modalId, formId) {
        this.modal = document.getElementById(modalId);
        this.form = document.getElementById(formId);
        this.prefs = null;

        document.getElementById('savePreferences')?.addEventListener('click', () => this.save());
        document.getElementById('cancelPreferences')?.addEventListener('click', () => this.hide());

        // Close on backdrop click
        this.modal?.addEventListener('click', (e) => {
            if (e.target === this.modal) this.hide();
        });
    }

    async show() {
        try {
            this.prefs = await API.get('/api/preferences');
            this.render();
            this.modal.classList.add('active');
        } catch (err) {
            console.error('Error loading preferences:', err);
        }
    }

    hide() {
        this.modal.classList.remove('active');
    }

    render() {
        if (!this.prefs) return;

        const interests = (this.prefs.interests || []).join(', ');
        const categories = this.prefs.categories || [];

        this.form.innerHTML = `
            <div class="pref-group">
                <label for="pref-interests">Interests (comma-separated)</label>
                <input type="text" id="pref-interests" value="${interests}" placeholder="technology, science, sports...">
            </div>
            <div class="pref-group">
                <label>Content Types</label>
                <div class="pref-checkboxes">
                    ${['rss', 'video', 'podcast'].map(ct => `
                        <label class="pref-checkbox">
                            <input type="checkbox" name="contentType" value="${ct}"
                                ${(this.prefs.contentTypes || []).includes(ct) ? 'checked' : ''}>
                            ${ct.charAt(0).toUpperCase() + ct.slice(1)}
                        </label>
                    `).join('')}
                </div>
            </div>
        `;
    }

    async save() {
        const interests = document.getElementById('pref-interests')?.value
            .split(',')
            .map(s => s.trim())
            .filter(Boolean);

        const contentTypes = [...document.querySelectorAll('input[name="contentType"]:checked')]
            .map(cb => cb.value);

        const updated = {
            ...this.prefs,
            interests: interests || [],
            contentTypes: contentTypes,
        };

        try {
            await API.put('/api/preferences', updated);
            this.hide();
            // Reload news with new preferences
            location.reload();
        } catch (err) {
            console.error('Error saving preferences:', err);
        }
    }
}
