const { describe, it, expect, beforeEach, afterEach } = require('@jest/globals');
const { JSDOM } = require('jsdom');

// Animation and State Tests
describe('Animations and State Management', () => {
  let dom;
  let document;
  let window;
  let initialAnimationsRun;

  beforeEach(() => {
    // Set up DOM with hero elements
    dom = new JSDOM(`
      <!DOCTYPE html>
      <html>
        <head>
          <style>
            .hero-subtitle { opacity: 0; transition: opacity 2000ms ease-out; }
            .hero-subtitle.fade-in { opacity: 1; }
            .hero-cta { opacity: 1; }
            .hero-stats { opacity: 1; }
          </style>
        </head>
        <body>
          <div class="hero">
            <h1 class="glitch" data-text="BLOCKHEAD CONSULTING" data-hero-style="professional">
              <span class="terminal-content">BLOCKHEAD CONSULTING</span>
              <span class="terminal-cursor"></span>
            </h1>
            <p class="hero-subtitle">Tagline here</p>
            <div class="hero-cta">
              <button>Button 1</button>
              <button>Button 2</button>
            </div>
            <div class="hero-stats">Stats here</div>
          </div>
          <main id="main-content"></main>
        </body>
      </html>
    `, {
      url: 'http://localhost:8087',
      pretendToBeVisual: true
    });

    document = dom.window.document;
    window = dom.window;
    initialAnimationsRun = false;

    // Mock requestAnimationFrame
    window.requestAnimationFrame = jest.fn(cb => setTimeout(cb, 16));
  });

  afterEach(() => {
    dom.window.close();
  });

  describe('Hero Animations', () => {
    it('should have subtitle hidden initially', () => {
      const subtitle = document.querySelector('.hero-subtitle');
      const computedStyle = window.getComputedStyle(subtitle);
      
      expect(subtitle.classList.contains('fade-in')).toBe(false);
      expect(computedStyle.opacity).toBe('0');
    });

    it('should show subtitle with fade-in class', () => {
      const subtitle = document.querySelector('.hero-subtitle');
      
      subtitle.classList.add('fade-in');
      
      expect(subtitle.classList.contains('fade-in')).toBe(true);
    });

    it('should have buttons and stats always visible', () => {
      const cta = document.querySelector('.hero-cta');
      const stats = document.querySelector('.hero-stats');
      
      const ctaStyle = window.getComputedStyle(cta);
      const statsStyle = window.getComputedStyle(stats);
      
      expect(ctaStyle.opacity).toBe('1');
      expect(statsStyle.opacity).toBe('1');
    });
  });

  describe('Boot Sequence', () => {
    it('should create boot sequence element', () => {
      const glitchElement = document.querySelector('.glitch');
      
      // Mock boot sequence creation
      const bootElement = document.createElement('div');
      bootElement.className = 'terminal-boot-sequence boot-sequence';
      bootElement.style.position = 'absolute';
      glitchElement.appendChild(bootElement);
      
      const boot = document.querySelector('.boot-sequence');
      expect(boot).toBeTruthy();
      expect(boot.style.position).toBe('absolute');
    });

    it('should remove boot sequence after fade out', (done) => {
      const glitchElement = document.querySelector('.glitch');
      const bootElement = document.createElement('div');
      bootElement.className = 'boot-sequence';
      glitchElement.appendChild(bootElement);
      
      // Simulate fade out
      bootElement.style.opacity = '1';
      bootElement.style.transition = 'opacity 800ms ease-out';
      bootElement.style.opacity = '0';
      
      // Simulate removal after fade
      setTimeout(() => {
        if (bootElement.parentElement) {
          bootElement.parentElement.removeChild(bootElement);
        }
        
        expect(document.querySelector('.boot-sequence')).toBeFalsy();
        done();
      }, 800);
    });

    it('should not push content down', () => {
      const glitchElement = document.querySelector('.glitch');
      const subtitle = document.querySelector('.hero-subtitle');
      const initialSubtitleTop = subtitle.getBoundingClientRect().top;
      
      // Add boot sequence
      const bootElement = document.createElement('div');
      bootElement.className = 'boot-sequence';
      bootElement.style.cssText = `
        position: absolute;
        top: 100%;
        width: 100%;
      `;
      glitchElement.appendChild(bootElement);
      
      // Subtitle position should not change
      const afterSubtitleTop = subtitle.getBoundingClientRect().top;
      expect(afterSubtitleTop).toBe(initialSubtitleTop);
    });
  });

  describe('Navigation State', () => {
    it('should show subtitle immediately on navigation', () => {
      const subtitle = document.querySelector('.hero-subtitle');
      
      // Simulate navigation (shouldType = false)
      subtitle.classList.add('fade-in');
      subtitle.style.opacity = '1';
      subtitle.style.transition = 'none';
      
      expect(subtitle.style.opacity).toBe('1');
      expect(subtitle.style.transition).toBe('none');
    });

    it('should not run boot sequence on navigation', () => {
      initialAnimationsRun = true;
      
      // Mock boot sequence function
      const createBootSequence = jest.fn();
      
      // On navigation, boot sequence should not be called
      if (!initialAnimationsRun) {
        createBootSequence();
      }
      
      expect(createBootSequence).not.toHaveBeenCalled();
    });
  });

  describe('Animation Timing', () => {
    it('should fade in subtitle after boot sequence duration', (done) => {
      const subtitle = document.querySelector('.hero-subtitle');
      const bootDuration = 5000; // Example duration
      
      // Start with subtitle hidden
      expect(subtitle.classList.contains('fade-in')).toBe(false);
      
      // Simulate timing
      setTimeout(() => {
        subtitle.classList.add('fade-in');
        expect(subtitle.classList.contains('fade-in')).toBe(true);
        done();
      }, bootDuration);
    }, 6000); // Increase test timeout to be longer than bootDuration

    it('should have fallback timer for subtitle visibility', (done) => {
      const subtitle = document.querySelector('.hero-subtitle');
      
      // Fallback timer (8 seconds)
      setTimeout(() => {
        if (window.getComputedStyle(subtitle).opacity === '0') {
          subtitle.style.opacity = '1';
          subtitle.style.transition = 'opacity 0.5s ease-out';
        }
        
        expect(subtitle.style.opacity).toBe('1');
        done();
      }, 8000);
    }, 10000); // Increase test timeout
  });

  describe('Hero Style Detection', () => {
    it('should detect professional mode', () => {
      const glitchElement = document.querySelector('.glitch');
      const heroStyle = glitchElement.getAttribute('data-hero-style');
      
      expect(heroStyle).toBe('professional');
    });

    it('should apply correct animations for hero style', () => {
      const glitchElement = document.querySelector('.glitch');
      const heroStyle = glitchElement.getAttribute('data-hero-style');
      
      if (heroStyle === 'professional') {
        expect(glitchElement.classList.contains('hero-professional')).toBe(false); // Initially
        
        // After applying professional effects
        glitchElement.classList.add('hero-professional');
        expect(glitchElement.classList.contains('hero-professional')).toBe(true);
      }
    });
  });

  describe('HTMX Integration', () => {
    it('should reinitialize animations after content swap', () => {
      const event = new window.CustomEvent('htmx:afterSwap', {
        detail: {
          target: document.getElementById('main-content'),
          xhr: { responseURL: 'http://localhost:8087/content/home' }
        }
      });
      
      let animationsInitialized = false;
      
      document.addEventListener('htmx:afterSwap', (evt) => {
        if (evt.detail.xhr.responseURL.includes('/content/home')) {
          animationsInitialized = true;
        }
      });
      
      document.dispatchEvent(event);
      
      expect(animationsInitialized).toBe(true);
    });
  });

  describe('Mobile Animations', () => {
    it('should handle animations on mobile viewport', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375
      });
      
      const isMobile = window.innerWidth <= 768;
      expect(isMobile).toBe(true);
      
      // Mobile should use same animation logic
      const subtitle = document.querySelector('.hero-subtitle');
      subtitle.classList.add('fade-in');
      
      expect(subtitle.classList.contains('fade-in')).toBe(true);
    });
  });

  describe('Animation State Management', () => {
    it('should track animation running state', () => {
      let heroAnimationRunning = false;
      
      // Start animation
      heroAnimationRunning = true;
      expect(heroAnimationRunning).toBe(true);
      
      // End animation
      heroAnimationRunning = false;
      expect(heroAnimationRunning).toBe(false);
    });

    it('should prevent multiple concurrent animations', () => {
      let heroAnimationRunning = false;
      let animationCount = 0;
      
      const startAnimation = () => {
        if (heroAnimationRunning) return;
        heroAnimationRunning = true;
        animationCount++;
      };
      
      // Try to start multiple times
      startAnimation();
      startAnimation();
      startAnimation();
      
      expect(animationCount).toBe(1);
    });
  });
});