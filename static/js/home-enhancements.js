/**
 * Clean Homepage Effects
 * Minimal, professional interactions
 */

class HomeEnhancements {
    constructor() {
        this.init();
    }

    init() {
        // Only keep smooth scrolling functionality
        this.initSmoothScrolling();
    }

    /**
     * Enhanced smooth scrolling for anchor links
     */
    initSmoothScrolling() {
        document.querySelectorAll('a[href^="#"]').forEach(anchor => {
            anchor.addEventListener('click', function (e) {
                e.preventDefault();
                const target = document.querySelector(this.getAttribute('href'));
                if (target) {
                    target.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });
                }
            });
        });
    }
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    new HomeEnhancements();
});

// Also initialize if already loaded
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => new HomeEnhancements());
} else {
    new HomeEnhancements();
}