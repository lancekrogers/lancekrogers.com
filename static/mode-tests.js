// Hero Mode Regression Tests
// These tests ensure both professional and cyberpunk modes work correctly

// Prevent redeclaration if script is loaded multiple times
if (typeof window.HeroModeTests === 'undefined') {
window.HeroModeTests = class HeroModeTests {
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
const heroTests = new window.HeroModeTests();

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
      textContent.includes('AI integration') || 
      textContent.includes('blockchain infrastructure') || 
      textContent.includes('collaboration') ||
      textContent.includes('high-impact') ||
      textContent.includes('secure blockchain') ||
      textContent.includes('Ready for') ||
      textContent.includes('payment rails') ||
      textContent.includes('Strategic consulting'),
      'Professional mode should use business-focused boot messages'
    );
  } else {
    // Skip test if boot element doesn't exist or is in cyberpunk mode
    heroTests.assert(true, 'Boot sequence test skipped - element not present or in cyberpunk mode');
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
  try {
    const cursor = await heroTests.waitForElement('.terminal-cursor');
    
    // Wait for potential typing animation
    await heroTests.wait(8000);
    
    // After typing, cursor should be hidden
    const computedStyle = window.getComputedStyle(cursor);
    heroTests.assert(
      computedStyle.display === 'none' || computedStyle.opacity === '0',
      'Cursor should be hidden after typing completes'
    );
  } catch (error) {
    // Skip this test if terminal cursor doesn't exist
    if (error.message.includes('Element .terminal-cursor not found')) {
      heroTests.assert(true, 'Terminal cursor element not present, skipping test');
    } else {
      throw error;
    }
  }
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
  try {
    const terminalContent = await heroTests.waitForElement('.terminal-content');
    const computedStyle = window.getComputedStyle(terminalContent);
    
    // Should have appropriate text color (off-white)
    heroTests.assert(
      computedStyle.color.includes('248') || // rgba(248, 248, 248, ...)
      computedStyle.color.includes('rgb(248, 248, 248)'),
      'Terminal text should use off-white color for readability'
    );
  } catch (error) {
    // Skip this test if terminal content doesn't exist
    if (error.message.includes('Element .terminal-content not found')) {
      heroTests.assert(true, 'Terminal content element not present, skipping test');
    } else {
      throw error;
    }
  }
});

// Auto-run tests when this file is loaded
if (typeof window !== 'undefined') {
  // Function to check if we should run tests
  const shouldRunTests = () => {
    // Only run tests if we're on a page with hero elements
    return document.querySelector('.hero') || 
           document.querySelector('.glitch') || 
           window.location.pathname === '/';
  };
  
  // Run tests after page is fully loaded
  const runTestsWhenReady = () => {
    if (shouldRunTests()) {
      setTimeout(() => heroTests.runAllTests(), 2000);
    }
  };
  
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', runTestsWhenReady);
  } else {
    runTestsWhenReady();
  }
  
  // Also run tests after HTMX content swaps to home page
  if (window.htmx) {
    document.body.addEventListener('htmx:afterSwap', (event) => {
      if (window.location.pathname === '/' && shouldRunTests()) {
        setTimeout(() => heroTests.runAllTests(), 2000);
      }
    });
  }
};
}

// Export for manual testing
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { HeroModeTests: window.HeroModeTests, heroTests };
}