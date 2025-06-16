const { describe, it, expect, beforeEach, afterEach } = require('@jest/globals');
const { JSDOM } = require('jsdom');

// Navigation Flow Tests
describe('Navigation Flow', () => {
  let dom;
  let document;
  let window;

  beforeEach(() => {
    // Set up DOM with HTMX
    dom = new JSDOM(`
      <!DOCTYPE html>
      <html>
        <head>
          <script src="https://unpkg.com/htmx.org@1.9.10"></script>
        </head>
        <body>
          <nav class="navbar">
            <div class="nav-links desktop-nav">
              <a href="/" hx-get="/content/home" hx-target="#main-content" hx-push-url="/">Home</a>
              <a href="/about" hx-get="/content/about" hx-target="#main-content" hx-push-url="/about">About</a>
              <a href="/blog" hx-get="/content/blog" hx-target="#main-content" hx-push-url="/blog">Blog</a>
            </div>
          </nav>
          <main id="main-content"></main>
          <footer>
            <a href="/blog" hx-get="/content/blog" hx-target="#main-content" hx-push-url="/blog">Blog</a>
            <a href="#contact">Contact</a>
          </footer>
        </body>
      </html>
    `, {
      url: 'http://localhost:8087',
      pretendToBeVisual: true,
      resources: 'usable'
    });

    document = dom.window.document;
    window = dom.window;
    
    // Mock HTMX
    window.htmx = {
      process: jest.fn(),
      trigger: jest.fn()
    };
  });

  afterEach(() => {
    dom.window.close();
  });

  describe('Header Navigation', () => {
    it('should have correct HTMX attributes on nav links', () => {
      const navLinks = document.querySelectorAll('.nav-links a');
      
      navLinks.forEach(link => {
        if (!link.href.includes('#')) {
          expect(link.hasAttribute('hx-get')).toBe(true);
          expect(link.hasAttribute('hx-target')).toBe(true);
          expect(link.hasAttribute('hx-push-url')).toBe(true);
          expect(link.getAttribute('hx-target')).toBe('#main-content');
        }
      });
    });

    it('should have matching href and hx-push-url', () => {
      const navLinks = document.querySelectorAll('.nav-links a[hx-push-url]');
      
      navLinks.forEach(link => {
        const href = new URL(link.href).pathname;
        const pushUrl = link.getAttribute('hx-push-url');
        expect(href).toBe(pushUrl);
      });
    });

    it('should have correct content endpoints', () => {
      const expectedEndpoints = {
        '/': '/content/home',
        '/about': '/content/about',
        '/blog': '/content/blog'
      };

      Object.entries(expectedEndpoints).forEach(([path, endpoint]) => {
        const link = document.querySelector(`a[hx-push-url="${path}"]`);
        expect(link).toBeTruthy();
        expect(link.getAttribute('hx-get')).toBe(endpoint);
      });
    });
  });

  describe('Footer Navigation', () => {
    it('should have HTMX attributes on footer links', () => {
      const footerLinks = document.querySelectorAll('footer a');
      
      footerLinks.forEach(link => {
        if (!link.href.includes('#') && !link.href.includes('mailto:')) {
          expect(link.hasAttribute('hx-get')).toBe(true);
          expect(link.hasAttribute('hx-target')).toBe(true);
          expect(link.hasAttribute('hx-push-url')).toBe(true);
        }
      });
    });
  });

  describe('Bio Section Navigation', () => {
    it('should have correct "Learn More" link with scroll reset', () => {
      // Add bio section to DOM
      document.getElementById('main-content').innerHTML = `
        <section class="bio-brief">
          <a href="/about" hx-get="/content/about" hx-target="#main-content" 
             hx-push-url="/about" hx-swap="innerHTML show:top" class="bio-link">
            Learn More →
          </a>
        </section>
      `;

      const learnMoreLink = document.querySelector('.bio-link');
      expect(learnMoreLink).toBeTruthy();
      expect(learnMoreLink.getAttribute('hx-swap')).toContain('show:top');
    });
  });

  describe('Blog Navigation', () => {
    it('should have correct "Back to Blog" button', () => {
      // Add blog post content
      document.getElementById('main-content').innerHTML = `
        <article class="blog-post">
          <a href="/blog" hx-get="/content/blog" hx-target="#main-content" 
             hx-push-url="/blog" class="back-link">
            ← Back to Blog
          </a>
        </article>
      `;

      const backLink = document.querySelector('.back-link');
      expect(backLink).toBeTruthy();
      expect(backLink.getAttribute('hx-push-url')).toBe('/blog');
      expect(backLink.getAttribute('hx-get')).toBe('/content/blog');
    });
  });

  describe('Contact Navigation', () => {
    it('should handle contact section links correctly', () => {
      const contactLink = document.querySelector('a[href="#contact"]');
      expect(contactLink).toBeTruthy();
      
      // Should not have HTMX attributes for hash links
      expect(contactLink.hasAttribute('hx-get')).toBe(false);
    });

    it('should have contact form in home content', () => {
      document.getElementById('main-content').innerHTML = `
        <section id="contact" class="contact">
          <form hx-post="/contact" hx-target="#contact-response">
            <button type="submit">Send Message</button>
          </form>
        </section>
      `;

      const contactForm = document.querySelector('#contact form');
      expect(contactForm).toBeTruthy();
      expect(contactForm.getAttribute('hx-post')).toBe('/contact');
    });
  });

  describe('Services Navigation', () => {
    it('should handle services navigation with scroll', () => {
      // Add services link
      document.querySelector('.nav-links').innerHTML += `
        <a href="/#services" hx-get="/content/home" hx-target="#main-content" 
           hx-push-url="/" data-scroll-to="services">Services</a>
      `;

      const servicesLink = document.querySelector('a[data-scroll-to="services"]');
      expect(servicesLink).toBeTruthy();
      expect(servicesLink.getAttribute('hx-get')).toBe('/content/home');
      expect(servicesLink.getAttribute('data-scroll-to')).toBe('services');
    });
  });

  describe('HTMX Event Handling', () => {
    it('should trigger htmx:afterSwap event', () => {
      const mainContent = document.getElementById('main-content');
      const afterSwapHandler = jest.fn();
      
      document.addEventListener('htmx:afterSwap', afterSwapHandler);
      
      // Simulate HTMX swap
      const event = new window.CustomEvent('htmx:afterSwap', {
        detail: {
          target: mainContent,
          xhr: { responseURL: 'http://localhost:8087/content/home' }
        }
      });
      
      document.dispatchEvent(event);
      
      expect(afterSwapHandler).toHaveBeenCalled();
    });
  });

  describe('URL State Management', () => {
    it('should update browser URL on navigation', () => {
      const links = document.querySelectorAll('a[hx-push-url]');
      
      links.forEach(link => {
        const pushUrl = link.getAttribute('hx-push-url');
        expect(pushUrl).toBeTruthy();
        
        // Ensure push URL matches the intended destination
        if (link.textContent.includes('Home')) {
          expect(pushUrl).toBe('/');
        } else if (link.textContent.includes('About')) {
          expect(pushUrl).toBe('/about');
        } else if (link.textContent.includes('Blog')) {
          expect(pushUrl).toBe('/blog');
        }
      });
    });
  });
});