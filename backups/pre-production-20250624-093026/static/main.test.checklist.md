// Basic JavaScript tests for SPA navigation functionality
// Note: These are manual tests since we don't have a JS testing framework set up yet

/**
 * Manual test checklist for SPA navigation:
 * 
 * 1. Home page direct load:
 *    - [ ] Full glitch animation plays
 *    - [ ] Home button is hidden in navigation
 *    - [ ] Hero text animation plays after glitch
 * 
 * 2. Navigation from home page:
 *    - [ ] Blog link loads blog content via HTMX (no refresh)
 *    - [ ] Calendar link loads calendar content via HTMX (no refresh)
 *    - [ ] Services link scrolls to services section
 * 
 * 3. Navigation from blog page:
 *    - [ ] Home link loads home content via HTMX (hero text only, no full glitch)
 *    - [ ] Calendar link loads calendar content via HTMX (no refresh)
 *    - [ ] Services link loads home page and scrolls to services section
 * 
 * 4. Navigation from calendar page:
 *    - [ ] Home link loads home content via HTMX (hero text only, no full glitch)
 *    - [ ] Blog link loads blog content via HTMX (no refresh)
 *    - [ ] Services link loads home page and scrolls to services section
 * 
 * 5. Calendar functionality:
 *    - [ ] Time slots load when calendar page is accessed
 *    - [ ] Week navigation works (prev/next buttons)
 *    - [ ] Time slot selection works
 *    - [ ] Booking form appears when slot selected
 *    - [ ] Form submission works
 * 
 * 6. Navigation state management:
 *    - [ ] Active nav item updates correctly
 *    - [ ] Home button visibility toggles correctly
 *    - [ ] URL updates properly with HTMX navigation
 */

// Test helper functions that could be used with a proper testing framework

function testHomeButtonVisibility() {
  const homeLink = document.querySelector(".nav-links a[href='/']");
  const isHomePage = document.querySelector('.hero') !== null || 
                     document.querySelector('.glitch') !== null ||
                     window.location.pathname === '/';
  
  if (isHomePage && homeLink && homeLink.style.display !== 'none') {
    console.error('Test failed: Home button should be hidden on home page');
    return false;
  }
  
  if (!isHomePage && homeLink && homeLink.style.display === 'none') {
    console.error('Test failed: Home button should be visible on non-home pages');
    return false;
  }
  
  console.log('Test passed: Home button visibility is correct');
  return true;
}

function testCalendarInitialization() {
  const calendarContainer = document.getElementById('calendar-container');
  const timeSlots = document.getElementById('time-slots');
  
  if (!calendarContainer || !timeSlots) {
    console.log('Test skipped: Not on calendar page');
    return true;
  }
  
  // Check if calendar JavaScript has loaded
  if (typeof loadSlots !== 'function') {
    console.error('Test failed: Calendar functions not available');
    return false;
  }
  
  console.log('Test passed: Calendar initialization successful');
  return true;
}

function testAnimationInitialization() {
  const glitchElement = document.querySelector('.glitch');
  
  if (!glitchElement) {
    console.log('Test skipped: Not on home page');
    return true;
  }
  
  // Check if animation functions are available
  if (typeof initializeAnimations !== 'function' || 
      typeof initializeHeroTextAnimation !== 'function') {
    console.error('Test failed: Animation functions not available');
    return false;
  }
  
  console.log('Test passed: Animation functions available');
  return true;
}

// Auto-run basic tests when this file is loaded
if (typeof window !== 'undefined' && window.document) {
  document.addEventListener('DOMContentLoaded', function() {
    console.log('Running basic SPA navigation tests...');
    
    setTimeout(function() {
      testHomeButtonVisibility();
      testCalendarInitialization();
      testAnimationInitialization();
    }, 100);
  });
  
  // Test after HTMX content swaps
  document.addEventListener('htmx:afterSwap', function() {
    setTimeout(function() {
      testHomeButtonVisibility();
      testCalendarInitialization();
    }, 50);
  });
}