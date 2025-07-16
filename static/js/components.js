// PancakesGit UI Components

// Modal Component
class Modal {
    constructor(element) {
        this.element = element;
        this.isOpen = false;
        this.init();
    }
    
    init() {
        // Close button
        const closeBtn = this.element.querySelector('.modal-close');
        if (closeBtn) {
            closeBtn.addEventListener('click', () => this.close());
        }
        
        // Backdrop click
        this.element.addEventListener('click', (e) => {
            if (e.target === this.element) {
                this.close();
            }
        });
        
        // Escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.isOpen) {
                this.close();
            }
        });
    }
    
    open() {
        this.element.classList.add('show');
        document.body.classList.add('modal-open');
        this.isOpen = true;
        
        // Focus first focusable element
        const focusable = this.element.querySelector('input, button, textarea, select, a[href]');
        if (focusable) {
            focusable.focus();
        }
    }
    
    close() {
        this.element.classList.remove('show');
        document.body.classList.remove('modal-open');
        this.isOpen = false;
    }
}

// Tooltip Component
class Tooltip {
    constructor(element) {
        this.element = element;
        this.tooltip = null;
        this.init();
    }
    
    init() {
        this.element.addEventListener('mouseenter', () => this.show());
        this.element.addEventListener('mouseleave', () => this.hide());
        this.element.addEventListener('focus', () => this.show());
        this.element.addEventListener('blur', () => this.hide());
    }
    
    show() {
        const text = this.element.getAttribute('data-tooltip');
        if (!text) return;
        
        this.tooltip = document.createElement('div');
        this.tooltip.className = 'tooltip';
        this.tooltip.textContent = text;
        document.body.appendChild(this.tooltip);
        
        this.position();
    }
    
    hide() {
        if (this.tooltip) {
            this.tooltip.remove();
            this.tooltip = null;
        }
    }
    
    position() {
        if (!this.tooltip) return;
        
        const rect = this.element.getBoundingClientRect();
        const tooltipRect = this.tooltip.getBoundingClientRect();
        
        let top = rect.top - tooltipRect.height - 8;
        let left = rect.left + (rect.width - tooltipRect.width) / 2;
        
        // Adjust if tooltip goes off screen
        if (top < 0) {
            top = rect.bottom + 8;
            this.tooltip.classList.add('tooltip-bottom');
        }
        
        if (left < 0) {
            left = 8;
        } else if (left + tooltipRect.width > window.innerWidth) {
            left = window.innerWidth - tooltipRect.width - 8;
        }
        
        this.tooltip.style.top = `${top}px`;
        this.tooltip.style.left = `${left}px`;
    }
}

// Tab Component
class Tabs {
    constructor(element) {
        this.element = element;
        this.tabs = this.element.querySelectorAll('.tab-btn');
        this.panels = this.element.querySelectorAll('.tab-panel');
        this.init();
    }
    
    init() {
        this.tabs.forEach((tab, index) => {
            tab.addEventListener('click', () => this.switchTab(index));
            tab.addEventListener('keydown', (e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    this.switchTab(index);
                }
            });
        });
    }
    
    switchTab(index) {
        // Remove active class from all tabs and panels
        this.tabs.forEach(tab => tab.classList.remove('active'));
        this.panels.forEach(panel => panel.classList.remove('active'));
        
        // Add active class to selected tab and panel
        this.tabs[index].classList.add('active');
        if (this.panels[index]) {
            this.panels[index].classList.add('active');
        }
        
        // Emit custom event
        this.element.dispatchEvent(new CustomEvent('tabchange', {
            detail: { index, tab: this.tabs[index] }
        }));
    }
}

// File Upload Component
class FileUpload {
    constructor(element) {
        this.element = element;
        this.input = this.element.querySelector('input[type="file"]');
        this.dropZone = this.element.querySelector('.drop-zone');
        this.preview = this.element.querySelector('.file-preview');
        this.init();
    }
    
    init() {
        if (!this.input || !this.dropZone) return;
        
        // Drag and drop
        this.dropZone.addEventListener('dragover', (e) => {
            e.preventDefault();
            this.dropZone.classList.add('drag-over');
        });
        
        this.dropZone.addEventListener('dragleave', () => {
            this.dropZone.classList.remove('drag-over');
        });
        
        this.dropZone.addEventListener('drop', (e) => {
            e.preventDefault();
            this.dropZone.classList.remove('drag-over');
            this.handleFiles(e.dataTransfer.files);
        });
        
        // File input change
        this.input.addEventListener('change', (e) => {
            this.handleFiles(e.target.files);
        });
        
        // Click to select
        this.dropZone.addEventListener('click', () => {
            this.input.click();
        });
    }
    
    handleFiles(files) {
        if (files.length === 0) return;
        
        const file = files[0];
        this.showPreview(file);
        
        // Emit custom event
        this.element.dispatchEvent(new CustomEvent('fileselected', {
            detail: { file }
        }));
    }
    
    showPreview(file) {
        if (!this.preview) return;
        
        this.preview.innerHTML = `
            <div class="file-info">
                <i data-feather="file"></i>
                <span class="file-name">${file.name}</span>
                <span class="file-size">${this.formatFileSize(file.size)}</span>
            </div>
        `;
        
        // Re-initialize feather icons
        if (window.feather) {
            feather.replace();
        }
    }
    
    formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
}

// Progress Bar Component
class ProgressBar {
    constructor(element) {
        this.element = element;
        this.bar = this.element.querySelector('.progress-bar');
        this.text = this.element.querySelector('.progress-text');
        this.value = 0;
    }
    
    setValue(value, text = '') {
        this.value = Math.max(0, Math.min(100, value));
        
        if (this.bar) {
            this.bar.style.width = `${this.value}%`;
            this.bar.setAttribute('aria-valuenow', this.value);
        }
        
        if (this.text) {
            this.text.textContent = text || `${this.value}%`;
        }
    }
    
    setIndeterminate(indeterminate = true) {
        if (indeterminate) {
            this.element.classList.add('progress-indeterminate');
        } else {
            this.element.classList.remove('progress-indeterminate');
        }
    }
}

// Code Editor Component (simple)
class CodeEditor {
    constructor(element) {
        this.element = element;
        this.textarea = this.element.querySelector('textarea');
        this.init();
    }
    
    init() {
        if (!this.textarea) return;
        
        // Tab key support
        this.textarea.addEventListener('keydown', (e) => {
            if (e.key === 'Tab') {
                e.preventDefault();
                const start = this.textarea.selectionStart;
                const end = this.textarea.selectionEnd;
                const value = this.textarea.value;
                
                this.textarea.value = value.substring(0, start) + '  ' + value.substring(end);
                this.textarea.selectionStart = this.textarea.selectionEnd = start + 2;
            }
        });
        
        // Line numbers
        this.addLineNumbers();
        this.textarea.addEventListener('input', () => this.updateLineNumbers());
        this.textarea.addEventListener('scroll', () => this.syncScroll());
    }
    
    addLineNumbers() {
        const lineNumbers = document.createElement('div');
        lineNumbers.className = 'line-numbers';
        this.element.insertBefore(lineNumbers, this.textarea);
        this.lineNumbers = lineNumbers;
        this.updateLineNumbers();
    }
    
    updateLineNumbers() {
        if (!this.lineNumbers) return;
        
        const lines = this.textarea.value.split('\n').length;
        const numbers = Array.from({ length: lines }, (_, i) => i + 1).join('\n');
        this.lineNumbers.textContent = numbers;
    }
    
    syncScroll() {
        if (this.lineNumbers) {
            this.lineNumbers.scrollTop = this.textarea.scrollTop;
        }
    }
}

// Initialize all components
function initializeComponents() {
    // Modals
    document.querySelectorAll('.modal').forEach(el => new Modal(el));
    
    // Tooltips
    document.querySelectorAll('[data-tooltip]').forEach(el => new Tooltip(el));
    
    // Tabs
    document.querySelectorAll('.tabs').forEach(el => new Tabs(el));
    
    // File uploads
    document.querySelectorAll('.file-upload').forEach(el => new FileUpload(el));
    
    // Code editors
    document.querySelectorAll('.code-editor').forEach(el => new CodeEditor(el));
    
    // Progress bars
    document.querySelectorAll('.progress').forEach(el => {
        el.progressBar = new ProgressBar(el);
    });
}

// Auto-initialize on DOM load
document.addEventListener('DOMContentLoaded', initializeComponents);

// Export components for manual initialization
window.PancakesGitComponents = {
    Modal,
    Tooltip,
    Tabs,
    FileUpload,
    ProgressBar,
    CodeEditor,
    initializeComponents
};
