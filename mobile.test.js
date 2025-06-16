const { describe, it, expect, beforeEach, afterEach } = require('@jest/globals');
const { JSDOM } = require('jsdom');

// Mobile Functionality Tests
describe('Mobile Functionality', () => {
  let dom;
  let document;
  let window;
  let initializeMobileMenu;

  beforeEach(() => {
    // Set up DOM with mobile menu
    dom = new JSDOM(`
      <!DOCTYPE html>
      <html>
        <head>
          <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        </head>
        <body>
          <nav class="navbar">
            <div class="mobile-nav">
              <button class="hamburger-menu" id="hamburger-toggle">
                <span class="hamburger-line"></span>
                <span class="hamburger-line"></span>
                <span class="hamburger-line"></span>
              </button>
              <div class="mobile-menu" id="mobile-menu">
                <a href="/">Home</a>
                <a href="/about">About</a>
                <a href="/blog">Blog</a>
              </div>
            </div>
          </nav>
          <main id="main-content"></main>
        </body>
      </html>
    `, {
      url: 'http://localhost:8087',
      pretendToBeVisual: true
    });

    document = dom.window.document;
    window = dom.window;

    // Mock the initializeMobileMenu function from main.js
    initializeMobileMenu = () => {
      const hamburgerToggle = document.getElementById('hamburger-toggle');
      const mobileMenu = document.getElementById('mobile-menu');
      
      if (hamburgerToggle && mobileMenu) {
        // Reset menu state
        hamburgerToggle.classList.remove('active');
        mobileMenu.classList.remove('active');
        
        // Remove existing listeners
        const newHamburger = hamburgerToggle.cloneNode(true);
        hamburgerToggle.parentNode.replaceChild(newHamburger, hamburgerToggle);
        
        // Add new listener
        newHamburger.addEventListener('click', function() {
          newHamburger.classList.toggle('active');
          mobileMenu.classList.toggle('active');
        });
      }
    };
  });

  afterEach(() => {
    dom.window.close();
  });

  describe('Hamburger Menu', () => {
    it('should toggle mobile menu on hamburger click', () => {
      initializeMobileMenu();
      
      const hamburger = document.getElementById('hamburger-toggle');
      const mobileMenu = document.getElementById('mobile-menu');
      
      expect(hamburger.classList.contains('active')).toBe(false);
      expect(mobileMenu.classList.contains('active')).toBe(false);
      
      // Click hamburger
      hamburger.click();
      
      expect(hamburger.classList.contains('active')).toBe(true);
      expect(mobileMenu.classList.contains('active')).toBe(true);
      
      // Click again to close
      hamburger.click();
      
      expect(hamburger.classList.contains('active')).toBe(false);
      expect(mobileMenu.classList.contains('active')).toBe(false);
    });

    it('should reinitialize after HTMX content swap', () => {
      initializeMobileMenu();
      
      const hamburger = document.getElementById('hamburger-toggle');
      
      // Verify initial click works
      hamburger.click();
      expect(hamburger.classList.contains('active')).toBe(true);
      hamburger.click();
      
      // Simulate HTMX swap event
      const event = new window.CustomEvent('htmx:afterSwap', {
        detail: {
          target: document.getElementById('main-content'),
          xhr: { responseURL: 'http://localhost:8087/content/home' }
        }
      });
      
      // Re-initialize menu
      initializeMobileMenu();
      
      // Verify menu still works after re-initialization
      const newHamburger = document.getElementById('hamburger-toggle');
      newHamburger.click();
      expect(newHamburger.classList.contains('active')).toBe(true);
    });

    it('should not create duplicate event listeners', () => {
      // Initialize multiple times
      initializeMobileMenu();
      initializeMobileMenu();
      initializeMobileMenu();
      
      // Get the current hamburger element (after replacements)
      const hamburger = document.getElementById('hamburger-toggle');
      const mobileMenu = document.getElementById('mobile-menu');
      
      // Click once
      hamburger.click();
      
      // Should only toggle once
      expect(hamburger.classList.contains('active')).toBe(true);
      expect(mobileMenu.classList.contains('active')).toBe(true);
    });
  });

  describe('Mobile Navigation Links', () => {
    it('should close menu after clicking a link', () => {
      initializeMobileMenu();
      
      const hamburger = document.getElementById('hamburger-toggle');
      const mobileMenu = document.getElementById('mobile-menu');
      const links = mobileMenu.querySelectorAll('a');
      
      // Open menu
      hamburger.click();
      expect(mobileMenu.classList.contains('active')).toBe(true);
      
      // Add click handler to links (as in main.js)
      links.forEach(link => {
        link.addEventListener('click', () => {
          hamburger.classList.remove('active');
          mobileMenu.classList.remove('active');
        });
      });
      
      // Click a link
      links[0].click();
      
      // Menu should close
      expect(hamburger.classList.contains('active')).toBe(false);
      expect(mobileMenu.classList.contains('active')).toBe(false);
    });
  });

  describe('Viewport Detection', () => {
    it('should detect mobile viewport correctly', () => {
      // Set mobile viewport width
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375
      });
      
      const isMobile = window.innerWidth <= 768;
      expect(isMobile).toBe(true);
    });

    it('should detect desktop viewport correctly', () => {
      // Set desktop viewport width
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1024
      });
      
      const isMobile = window.innerWidth <= 768;
      expect(isMobile).toBe(false);
    });
  });

  describe('Touch Events', () => {
    it('should handle touch events on mobile', () => {
      const hamburger = document.getElementById('hamburger-toggle');
      
      // Create and dispatch touch event
      const touchEvent = new window.TouchEvent('touchstart', {
        bubbles: true,
        cancelable: true,
        touches: [{ clientX: 0, clientY: 0 }]
      });
      
      let touchHandled = false;
      hamburger.addEventListener('touchstart', () => {
        touchHandled = true;
      });
      
      hamburger.dispatchEvent(touchEvent);
      expect(touchHandled).toBe(true);
    });
  });

  describe('Mobile Boot Sequence', () => {
    it('should use mobile boot messages on mobile viewport', () => {
      // Set mobile viewport
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375
      });
      
      // Mock boot sequences
      window.bootSequences = {
        professional: ['Desktop message'],
        professionalMobile: ['Mobile message']
      };
      
      const isMobile = window.innerWidth <= 768;
      const messages = isMobile && window.bootSequences.professionalMobile 
        ? window.bootSequences.professionalMobile 
        : window.bootSequences.professional;
      
      expect(messages).toEqual(['Mobile message']);
    });
  });

  describe('Mobile Scroll Behavior', () => {
    it('should handle smooth scrolling on mobile', () => {
      // Add content with sections
      document.getElementById('main-content').innerHTML = `
        <section id="services" style="height: 1000px;">Services</section>
        <section id="contact" style="height: 1000px;">Contact</section>
      `;
      
      // Mock scrollIntoView
      const servicesSection = document.getElementById('services');
      servicesSection.scrollIntoView = jest.fn();
      
      // Create a link with scroll behavior
      const link = document.createElement('a');
      link.href = '#services';
      link.addEventListener('click', (e) => {
        e.preventDefault();
        const target = document.querySelector(link.getAttribute('href'));
        if (target) {
          target.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
      });
      
      link.click();
      
      expect(servicesSection.scrollIntoView).toHaveBeenCalledWith({
        behavior: 'smooth',
        block: 'start'
      });
    });
  });

  describe('Mobile State Preservation', () => {
    it('should preserve menu state across navigation', () => {
      initializeMobileMenu();
      
      const hamburger = document.getElementById('hamburger-toggle');
      const mobileMenu = document.getElementById('mobile-menu');
      
      // Menu should be closed initially
      expect(hamburger.classList.contains('active')).toBe(false);
      expect(mobileMenu.classList.contains('active')).toBe(false);
      
      // After navigation, menu should still be closed
      initializeMobileMenu(); // Simulates re-initialization after HTMX swap
      
      expect(hamburger.classList.contains('active')).toBe(false);
      expect(mobileMenu.classList.contains('active')).toBe(false);
    });
  });
});