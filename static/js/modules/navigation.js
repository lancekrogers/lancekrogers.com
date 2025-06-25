import { debug } from '../utils/debug.js';

// Mobile menu state
let mobileMenuInitialized = false;
let globalMobileHandlerAttached = false;
let currentHamburgerHandler = null;


// Manage smooth scrolling based on current page
export function initializeSmoothScroll() {
  updateSmoothScrollForCurrentPage();
  
  // Disable smooth scrolling during HTMX navigation to prevent unwanted animations
  document.addEventListener('htmx:beforeRequest', function() {
    document.documentElement.style.scrollBehavior = 'auto';
  });
  
  document.addEventListener('htmx:afterSwap', function() {
    // Update smooth scrolling based on the new page content
    setTimeout(() => {
      updateSmoothScrollForCurrentPage();
    }, 50);
  });
}

function updateSmoothScrollForCurrentPage() {
  // Check if we're on the home page
  const isHomePage = isCurrentPageHome();
  
  if (isHomePage) {
    // Enable smooth scrolling only on home page for anchor links
    document.documentElement.style.scrollBehavior = 'smooth';
    document.body.classList.add('home-page');
    debug('Smooth scrolling enabled for home page');
  } else {
    // Disable smooth scrolling on other pages
    document.documentElement.style.scrollBehavior = 'auto';
    document.body.classList.remove('home-page');
    debug('Smooth scrolling disabled for non-home page');
  }
}

function isCurrentPageHome() {
  // Multiple ways to detect home page
  return location.pathname === '/' || 
         document.querySelector('.hero') !== null ||
         document.querySelector('.glitch') !== null ||
         document.querySelector('[data-text="BLOCKHEAD CONSULTING"]') !== null;
}


// Mobile hamburger menu functionality
export function initializeMobileMenu() {
  const hamburgerToggle = document.getElementById('hamburger-toggle');
  const mobileMenu = document.getElementById('mobile-menu');
  
  if (hamburgerToggle && mobileMenu) {
    debug('Initializing mobile menu');
    
    // Reset menu state first
    hamburgerToggle.classList.remove('active');
    mobileMenu.classList.remove('active');
    
    // Remove existing event listener if it exists
    if (currentHamburgerHandler) {
      hamburgerToggle.removeEventListener('click', currentHamburgerHandler);
    }
    
    // Create new event handler
    currentHamburgerHandler = function(e) {
      e.preventDefault();
      e.stopPropagation();
      debug('Hamburger clicked, toggling menu');
      
      const isActive = hamburgerToggle.classList.contains('active');
      if (isActive) {
        hamburgerToggle.classList.remove('active');
        mobileMenu.classList.remove('active');
      } else {
        hamburgerToggle.classList.add('active');
        mobileMenu.classList.add('active');
      }
    };
    
    // Add the new event listener
    hamburgerToggle.addEventListener('click', currentHamburgerHandler);

    // Close menu when clicking on any mobile menu link (only add once)
    if (!mobileMenuInitialized) {
      mobileMenu.addEventListener('click', function(e) {
        if (e.target.tagName === 'A') {
          debug('Mobile menu link clicked, closing menu');
          hamburgerToggle.classList.remove('active');
          mobileMenu.classList.remove('active');
        }
      });
    }
    
    mobileMenuInitialized = true;
  } else if (!hamburgerToggle) {
    debug('Hamburger toggle not found');
  } else if (!mobileMenu) {
    debug('Mobile menu not found');
  }
}

// Global click handler for closing mobile menu (only attach once)
export function attachGlobalMobileMenuHandler() {
  if (!globalMobileHandlerAttached) {
    document.addEventListener('click', function(event) {
      const hamburgerToggle = document.getElementById('hamburger-toggle');
      const mobileMenu = document.getElementById('mobile-menu');
      
      if (hamburgerToggle && mobileMenu && mobileMenu.classList.contains('active')) {
        // Close menu if clicked outside of both hamburger and menu
        if (!hamburgerToggle.contains(event.target) && !mobileMenu.contains(event.target)) {
          debug('Clicked outside mobile menu, closing');
          hamburgerToggle.classList.remove('active');
          mobileMenu.classList.remove('active');
        }
      }
    });
    
    globalMobileHandlerAttached = true;
    debug('Global mobile menu handler attached');
  }
}

// Add active class to current nav item and hide home link on home page
export function updateNavigation() {
  const currentLocation = location.pathname;
  const desktopMenuItems = document.querySelectorAll(".desktop-nav a");
  const mobileMenuItems = document.querySelectorAll(".mobile-menu a");
  const desktopHomeLink = document.querySelector(".desktop-nav a[href='/']");
  const mobileHomeLink = document.querySelector(".mobile-menu a[href='/']");
  
  debug('Updating navigation for path:', currentLocation);
  
  // Update desktop navigation
  desktopMenuItems.forEach((item) => {
    item.classList.remove("active");
    if (item.getAttribute("href") === currentLocation) {
      item.classList.add("active");
    }
  });
  
  // Update mobile navigation
  mobileMenuItems.forEach((item) => {
    item.classList.remove("active");
    if (item.getAttribute("href") === currentLocation) {
      item.classList.add("active");
    }
  });
  
  // Hide home link when on home page
  const isHomePage = currentLocation === '/' || 
                     document.querySelector('.hero') !== null ||
                     document.querySelector('.glitch') !== null;
  
  if (desktopHomeLink) {
    if (isHomePage) {
      desktopHomeLink.style.display = 'none';
    } else {
      desktopHomeLink.style.display = '';
    }
  }
  
  if (mobileHomeLink) {
    if (isHomePage) {
      mobileHomeLink.style.display = 'none';
    } else {
      mobileHomeLink.style.display = '';
    }
  }
}

// Reset mobile menu state (useful after HTMX navigation)
export function resetMobileMenuState() {
  // Close the menu if it's open, but don't reset initialization flag
  // since we handle event listener cleanup in initializeMobileMenu
  const hamburgerToggle = document.getElementById('hamburger-toggle');
  const mobileMenu = document.getElementById('mobile-menu');
  
  if (hamburgerToggle && mobileMenu) {
    hamburgerToggle.classList.remove('active');
    mobileMenu.classList.remove('active');
  }
}

// Export empty functions for compatibility
export function isScrollInProgress() {
  return false;
}

export function cancelPendingScrolls() {
  // No-op
}

// Initialize all navigation features
export function initializeNavigation() {
  initializeSmoothScroll();
  initializeMobileMenu();
  attachGlobalMobileMenuHandler();
  updateNavigation();
}