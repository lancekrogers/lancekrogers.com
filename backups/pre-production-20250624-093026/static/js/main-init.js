// Main initialization module that coordinates all features
import { setDebugLogging, debug } from './utils/debug.js';

// Add debug logging for about button clicks
document.addEventListener('click', function(e) {
  const link = e.target.closest('a');
  if (link) {
    const href = link.getAttribute('href');
    const hxGet = link.getAttribute('hx-get');
    
    // Log about button clicks specifically
    if (href === '/about' || hxGet === '/content/about') {
      console.log('🔍 About button clicked:', {
        href: href,
        hxGet: hxGet,
        target: e.target,
        link: link,
        defaultPrevented: e.defaultPrevented,
        propagationStopped: e.cancelBubble,
        timestamp: new Date().toISOString()
      });
    }
  }
}, true); // Use capture phase

import { initializeNavigation, resetMobileMenuState, updateNavigation, cancelPendingScrolls } from './modules/navigation.js';
import { initializeAnimations, hasInitialAnimationsRun } from './modules/animations.js';
import { initializePopups } from './modules/popups.js';
import { initializeCalendar } from './modules/calendar.js';
import { initializeFormEncryption } from './modules/form-encryption.js';
import { initializeContactModal, setupContactButtons } from './modules/contact-modal.js';

// Remove all scroll tracking - let browser handle it natively

// Track active HTMX requests to prevent duplicates
const activeRequests = new Map();

// Configure HTMX to prevent scroll interference and disable boost completely
htmx.config.scrollIntoViewOnBoost = false;
htmx.config.defaultFocusScroll = false;
htmx.config.includeIndicatorStyles = false;
htmx.config.globalViewTransitions = false;
htmx.config.scrollBehavior = 'auto';
// Disable boost completely - we use explicit hx-get attributes instead
htmx.config.boost = false;

// Now that Blog link has proper HTMX attributes, anchor links should work naturally
// No special handling needed since HTMX won't interfere with anchor scrolling

// Simple click management
let lastClickTime = 0;
document.addEventListener('click', function(e) {
  const now = Date.now();
  const timeSinceLastClick = now - lastClickTime;
  
  // Only apply debouncing to navigation links with HTMX (not anchor links)
  const htmxNavLink = e.target.closest('a[hx-get]');
  if (htmxNavLink && htmxNavLink.closest('.nav-links, .mobile-menu')) {
    if (timeSinceLastClick < 300) { // Prevent clicks within 300ms
      console.log('🛑 Click too fast, ignoring');
      e.preventDefault();
      e.stopImmediatePropagation();
      return false;
    }
    lastClickTime = now;
  }
  
}); // Remove capture phase to not interfere with normal links

// Debounce function to prevent rapid-fire requests
function debounce(func, wait) {
  let timeout;
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout);
      func(...args);
    };
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
}

// Initialize project cards functionality (handled by popups module)
function initializeProjectCards() {
  // Project cards functionality is now handled by the popups module
  debug('Project cards functionality handled by popups module');
}

// Initialize all features on page load
function initializeAll() {
  debug('Initializing all features');
  
  // Set debug logging based on environment (you can configure this)
  setDebugLogging(true); // Set to true for development
  console.log('Debug logging enabled:', true);
  
  // Initialize navigation features
  initializeNavigation();
  
  // Initialize hero animations
  initializeAnimations();
  
  // Initialize popups
  initializePopups();
  
  // Initialize calendar if on calendar page
  if (document.getElementById('time-slots') || document.querySelector('.calendar-page')) {
    initializeCalendar();
  }
  
  // Initialize form encryption
  initializeFormEncryption();
  
  // Initialize project cards
  initializeProjectCards();
  
  // Initialize contact modal
  initializeContactModal();
  setupContactButtons();
}

// Remove all scroll tracking - native browser handles it

// Run when HTMX loads new content
document.addEventListener('htmx:afterSwap', function(e) {
  debug('HTMX content swapped:', e.detail.target);
  
  // Reset body overflow in case a popup was left open
  document.body.style.overflow = '';
  
  // Scroll to top for non-home pages (but not for home page to preserve anchor scrolling)
  const isHomePage = document.querySelector('.hero') !== null;
  if (!isHomePage) {
    window.scrollTo({ top: 0, behavior: 'auto' });
    debug('Scrolled to top for non-home page');
  }
  
  // Always update navigation state
  updateNavigation();
  
  // Reset and reinitialize mobile menu
  resetMobileMenuState();
  initializeNavigation();
  
  // Reinitialize animations for home page content
  if (e.detail.target.id === 'main-content' || e.detail.target.classList.contains('main-content')) {
    // Check if we loaded home page content
    if (document.querySelector('.hero')) {
      debug('Home page content loaded, initializing animations');
      initializeAnimations();
    }
    
    // Initialize popups if there are any modal triggers or interactive elements
    if (document.querySelector('.packages-modal-trigger') ||
        document.querySelector('[data-packages-modal]') ||
        document.querySelector('.package-interactive') || 
        document.querySelector('.expertise-interactive') || 
        document.querySelector('.work-item.interactive') ||
        document.querySelector('[data-packages-main-modal]')) {
      debug('Reinitializing popups after HTMX content swap');
      initializePopups();
    }
    
    // Initialize calendar if loaded
    if (document.getElementById('time-slots')) {
      initializeCalendar();
    }
    
    // Initialize form encryption if contact form loaded
    if (document.querySelector('.contact-form')) {
      initializeFormEncryption();
    }
    
    // Always reinitialize contact modal and buttons after content swap
    // The modal is in the base template, not the swapped content
    initializeContactModal();
    setupContactButtons();
    
    // Initialize project cards if work page loaded
    if (document.querySelector('.work-page') || document.querySelector('.work-section')) {
      initializeProjectCards();
    }
  }
  
  // Don't interfere with scrolling - let browser handle it naturally
  
  // Fire custom event for other scripts
  document.dispatchEvent(new CustomEvent('contentSwapped', { 
    detail: { target: e.detail.target } 
  }));
});

// Handle browser back/forward navigation (but not hash changes)
window.addEventListener('popstate', function(e) {
  // Ignore hash-only changes (anchor links)
  if (e.state || window.location.pathname !== lastPathname) {
    debug('Browser navigation detected');
    updateNavigation();
    lastPathname = window.location.pathname;
  }
});

let lastPathname = window.location.pathname;

// Initialize everything when DOM is ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', initializeAll);
} else {
  initializeAll();
}

// Add HTMX request tracking with timing
const htmxRequests = new Map();

document.addEventListener('htmx:beforeRequest', function(e) {
  const path = e.detail.path || e.detail.requestConfig?.path;
  const target = e.detail.target;
  
  
  // Only prevent duplicates for navigation requests
  const isNavigation = e.detail.elt?.closest('.nav-links, .mobile-menu');
  if (isNavigation && activeRequests.has(path)) {
    const existingRequest = activeRequests.get(path);
    const elapsed = Date.now() - existingRequest.startTime;
    console.log('🛑 Cancelling duplicate navigation to:', path, `(existing request running for ${elapsed}ms)`);
    e.preventDefault();
    return;
  }
  
  const requestId = Date.now();
  const requestInfo = {
    target: target,
    verb: e.detail.verb || e.detail.requestConfig?.verb,
    path: path,
    startTime: requestId,
    timestamp: new Date().toISOString()
  };
  
  // Store request info
  activeRequests.set(path, requestInfo);
  htmxRequests.set(target, requestInfo);
  
  console.log('HTMX request started:', requestInfo);
});

document.addEventListener('htmx:afterRequest', function(e) {
  const requestInfo = htmxRequests.get(e.detail.target);
  if (requestInfo) {
    const elapsed = Date.now() - requestInfo.startTime;
    htmxRequests.delete(e.detail.target);
    
    // Remove from active requests
    activeRequests.delete(requestInfo.path);
    
    
    console.log('HTMX request completed:', {
      target: e.detail.target,
      successful: e.detail.successful,
      elapsed: elapsed + 'ms',
      timestamp: new Date().toISOString()
    });
    
    if (elapsed > 1000) {
      console.warn('⚠️ Slow HTMX request detected:', elapsed + 'ms');
    }
  }
});

document.addEventListener('htmx:responseError', function(e) {
  const requestInfo = htmxRequests.get(e.detail.target);
  htmxRequests.delete(e.detail.target);
  
  // Remove from active requests
  if (requestInfo) {
    activeRequests.delete(requestInfo.path);
  }
  
  console.error('HTMX request error:', {
    target: e.detail.target,
    error: e.detail.error,
    xhr: e.detail.xhr,
    requestInfo: requestInfo,
    timestamp: new Date().toISOString()
  });
});

// Also track timeout errors
document.addEventListener('htmx:timeout', function(e) {
  const requestInfo = htmxRequests.get(e.detail.target);
  htmxRequests.delete(e.detail.target);
  
  // Clear from active requests
  if (requestInfo) {
    activeRequests.delete(requestInfo.path);
  }
  
  console.error('HTMX request timeout:', {
    target: e.detail.target,
    requestInfo: requestInfo,
    timestamp: new Date().toISOString()
  });
});

// Clean up stuck requests periodically
setInterval(() => {
  const now = Date.now();
  const stuckRequests = [];
  
  activeRequests.forEach((request, key) => {
    const elapsed = now - request.startTime;
    if (elapsed > 30000) { // 30 seconds is definitely too long
      stuckRequests.push(key);
    }
  });
  
  if (stuckRequests.length > 0) {
    console.warn('⚠️ Cleaning up stuck requests:', stuckRequests);
    stuckRequests.forEach(key => activeRequests.delete(key));
  }
}, 10000); // Check every 10 seconds

// Guard against duplicate initialization
if (!window.blockheadInit) {
  // Export for global access if needed
  window.blockheadInit = {
    initializeAll,
    initializeNavigation,
    initializeAnimations,
    initializePopups,
    initializeCalendar,
    initializeFormEncryption,
    initializeProjectCards,
    initializeContactModal
  };
}