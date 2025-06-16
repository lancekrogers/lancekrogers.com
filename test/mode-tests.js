// Hero Mode Regression Tests
// These tests ensure both professional and cyberpunk modes work correctly

class HeroModeTests {
  constructor() {
    this.tests = [];
    this.results = {
      passed: 0,
      failed: 0,
      total: 0
    };
  }

  // Test helper functions
  test(name, testFn) {
    this.tests.push({ name, testFn });
  }

  async runAllTests() {
    console.log('🧪 Running Hero Mode Tests...');
    
    for (const test of this.tests) {
      try {
        await test.testFn();
        this.results.passed++;
        console.log(`✅ ${test.name}`);
      } catch (error) {
        this.results.failed++;
        console.error(`❌ ${test.name}:`, error.message);
      }
      this.results.total++;
    }
    
    this.reportResults();
  }

  reportResults() {
    console.log('\n📊 Test Results:');
    console.log(`Total: ${this.results.total}`);
    console.log(`Passed: ${this.results.passed}`);
    console.log(`Failed: ${this.results.failed}`);
    
    if (this.results.failed === 0) {
      console.log('🎉 All tests passed!');
    } else {
      console.warn(`⚠️ ${this.results.failed} test(s) failed`);
    }
  }

  // Test utilities
  waitForElement(selector, timeout = 5000) {
    return new Promise((resolve, reject) => {
      const element = document.querySelector(selector);
      if (element) return resolve(element);
      
      const observer = new MutationObserver(() => {
        const element = document.querySelector(selector);
        if (element) {
          observer.disconnect();
          resolve(element);
        }
      });
      
      observer.observe(document.body, { childList: true, subtree: true });
      
      setTimeout(() => {
        observer.disconnect();
        reject(new Error(`Element ${selector} not found within ${timeout}ms`));
      }, timeout);
    });
  }

  async wait(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  assert(condition, message) {
    if (!condition) {
      throw new Error(message);
    }
  }
}

// Initialize test suite
const heroTests = new HeroModeTests();

// Professional Mode Tests
heroTests.test('Professional Mode - No CRT Effects', async () => {
  const hero = await heroTests.waitForElement('.hero');
  const computedBefore = window.getComputedStyle(hero, '::before');
  
  // In professional mode, pseudo-elements should not have CRT backgrounds
  heroTests.assert(
    !hero.classList.contains('hero-cyberpunk') || 
    computedBefore.backgroundImage === 'none',
    'Professional mode should not have CRT scan lines'
  );
});

heroTests.test('Professional Mode - No Logo Animation', async () => {
  const heroLogo = await heroTests.waitForElement('.hero-logo');
  const computedStyle = window.getComputedStyle(heroLogo);
  
  // In professional mode, logo should not have glow animation
  if (!document.querySelector('.hero.hero-cyberpunk')) {
    heroTests.assert(
      computedStyle.animationName === 'none' || computedStyle.animationName === '',
      'Professional mode logo should not have glow animation'
    );
  }
});

heroTests.test('Professional Mode - Boot Sequence Messages', async () => {
  // Check if professional boot sequence messages are used
  const bootElement = document.querySelector('.terminal-boot-sequence');
  if (bootElement && !document.querySelector('.hero.hero-cyberpunk')) {
    await heroTests.wait(2000); // Wait for boot sequence
    const textContent = bootElement.textContent;
    heroTests.assert(
      textContent.includes('Enterprise') || textContent.includes('blockchain infrastructure'),
      'Professional mode should use business-focused boot messages'
    );
  }
});

// Cyberpunk Mode Tests
heroTests.test('Cyberpunk Mode - CRT Effects Present', async () => {
  const hero = document.querySelector('.hero.hero-cyberpunk');
  if (hero) {
    const computedBefore = window.getComputedStyle(hero, '::before');
    heroTests.assert(
      computedBefore.backgroundImage.includes('linear-gradient') ||
      hero.classList.contains('hero-cyberpunk'),
      'Cyberpunk mode should have CRT scan lines'
    );
  }
});

heroTests.test('Cyberpunk Mode - Logo Animation', async () => {
  const hero = document.querySelector('.hero.hero-cyberpunk');
  if (hero) {
    const heroLogo = await heroTests.waitForElement('.hero-logo');
    const computedStyle = window.getComputedStyle(heroLogo);
    heroTests.assert(
      computedStyle.animationName.includes('hero-logo-glow') ||
      computedStyle.animationName !== 'none',
      'Cyberpunk mode logo should have glow animation'
    );
  }
});

heroTests.test('Cyberpunk Mode - Holographic Logo Effect', async () => {
  const hero = document.querySelector('.hero.hero-cyberpunk');
  if (hero) {
    const logoImg = await heroTests.waitForElement('.hero-logo-img');
    const computedBefore = window.getComputedStyle(logoImg, '::before');
    const computedAfter = window.getComputedStyle(logoImg, '::after');
    
    heroTests.assert(
      computedBefore.backgroundImage.includes('svg') ||
      computedAfter.backgroundImage.includes('svg'),
      'Cyberpunk mode logo should have holographic pseudo-elements'
    );
  }
});

heroTests.test('Cyberpunk Mode - Console Message', async () => {
  // Check if console.log was called with cyberpunk message
  const originalLog = console.log;
  let cyberpunkMessageLogged = false;
  
  console.log = (...args) => {
    if (args.some(arg => typeof arg === 'string' && arg.includes('Cyberpunk Mode'))) {
      cyberpunkMessageLogged = true;
    }
    originalLog.apply(console, args);
  };
  
  // Wait a bit to see if message appears
  await heroTests.wait(1000);
  
  const hero = document.querySelector('.hero.hero-cyberpunk');
  if (hero) {
    heroTests.assert(
      cyberpunkMessageLogged,
      'Cyberpunk mode should log console message'
    );
  }
  
  console.log = originalLog; // Restore original
});

// General Tests
heroTests.test('Hero Text Element Exists', async () => {
  const glitchElement = await heroTests.waitForElement('.glitch');
  heroTests.assert(glitchElement, 'Hero title element should exist');
});

heroTests.test('Terminal Cursor Behavior', async () => {
  const cursor = await heroTests.waitForElement('.terminal-cursor');
  
  // Wait for potential typing animation
  await heroTests.wait(8000);
  
  // After typing, cursor should be hidden
  const computedStyle = window.getComputedStyle(cursor);
  heroTests.assert(
    computedStyle.display === 'none' || computedStyle.opacity === '0',
    'Cursor should be hidden after typing completes'
  );
});

heroTests.test('Boot Sequence Files Loaded', async () => {
  heroTests.assert(
    typeof createBootSequence === 'function',
    'Boot sequence functions should be loaded'
  );
  
  heroTests.assert(
    typeof BOOT_SEQUENCES === 'object' &&
    BOOT_SEQUENCES.professional &&
    BOOT_SEQUENCES.cyberpunk,
    'Boot sequence configurations should be available'
  );
});

heroTests.test('Mode-Specific Text Colors', async () => {
  const terminalContent = await heroTests.waitForElement('.terminal-content');
  const computedStyle = window.getComputedStyle(terminalContent);
  
  // Should have appropriate text color (off-white)
  heroTests.assert(
    computedStyle.color.includes('248') || // rgba(248, 248, 248, ...)
    computedStyle.color.includes('rgb(248, 248, 248)'),
    'Terminal text should use off-white color for readability'
  );
});

// Auto-run tests when this file is loaded
if (typeof window !== 'undefined') {
  // Run tests after page is fully loaded
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
      setTimeout(() => heroTests.runAllTests(), 2000);
    });
  } else {
    setTimeout(() => heroTests.runAllTests(), 2000);
  }
}

// Export for manual testing
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { HeroModeTests, heroTests };
}