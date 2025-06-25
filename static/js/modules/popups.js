import { debug } from '../utils/debug.js';

// Popup state
let isPopupOpen = false;
let isWorkPopupOpen = false;
let isMainPackageModalOpen = false;
let isTransitioning = false; // Prevent rapid state changes
let closeButtonHandlerInitialized = false; // Track if close handler is already added
let closeButtonHandler = null; // Store reference to handler for cleanup

// Package popup functionality
export function initializePackagePopups() {
  debug('Initializing package popups');
  
  // Remove any existing event listeners first
  document.removeEventListener('click', handlePackageClicks);
  
  // Add global click handler for package interactions
  document.addEventListener('click', handlePackageClicks);
  
  // Add direct handler for close buttons as fallback
  initializeCloseButtonHandlers();
  
  // Debug: Check if modal exists in DOM
  const modal = document.querySelector('[data-packages-main-modal]');
  const closeBtn = document.querySelector('.packages-main-modal-close');
  debug('Modal found:', !!modal, 'Close button found:', !!closeBtn);
  
  debug('Package popups initialized');
}

function initializeCloseButtonHandlers() {
  // Only add the handler once
  if (closeButtonHandlerInitialized) {
    debug('Close button handlers already initialized');
    return;
  }
  
  // Create the handler function
  closeButtonHandler = function(e) {
    // Check if the clicked element is the close button or its child (like the × symbol)
    if (e.target.classList.contains('packages-main-modal-close') || 
        e.target.parentElement?.classList.contains('packages-main-modal-close')) {
      debug('Main modal close button clicked - handler triggered');
      e.preventDefault();
      e.stopPropagation();
      e.stopImmediatePropagation(); // Stop all other handlers
      hideMainPackageModal();
      return;
    }
    
    if (e.target.classList.contains('package-popup-close') || 
        e.target.parentElement?.classList.contains('package-popup-close')) {
      debug('Package popup close button clicked');
      e.preventDefault();
      e.stopPropagation();
      e.stopImmediatePropagation(); // Stop all other handlers
      hideAllPackagePopups();
      return;
    }
  };
  
  // Use event delegation on document for dynamically added elements
  // Use capture phase with highest priority
  document.addEventListener('click', closeButtonHandler, true); // Use capture phase
  
  closeButtonHandlerInitialized = true;
  debug('Close button handlers initialized');
}


function showMainPackageModal() {
  // Prevent multiple modals from opening
  if (isMainPackageModalOpen || isTransitioning) {
    debug('Main package modal already open or transitioning');
    return;
  }
  
  const modal = document.querySelector('[data-packages-main-modal]');
  if (modal) {
    isTransitioning = true;
    
    // Hide any other popups first
    hideAllPackagePopups();
    hideAllExpertisePopups();
    
    modal.classList.add('active');
    isMainPackageModalOpen = true;
    
    // Prevent body scroll when modal is open
    document.body.style.overflow = 'hidden';
    
    debug('Main package modal shown');
    
    // Reset transition flag after animation
    setTimeout(() => {
      isTransitioning = false;
    }, 300);
  } else {
    debug('Main package modal element not found');
  }
}

function hideMainPackageModal() {
  debug('hideMainPackageModal called, isTransitioning:', isTransitioning);
  
  // Prevent rapid state changes
  if (isTransitioning) {
    debug('Transition in progress, skipping hide');
    return;
  }
  
  const modal = document.querySelector('[data-packages-main-modal]');
  if (modal) {
    isTransitioning = true;
    modal.classList.remove('active');
    isMainPackageModalOpen = false;
    
    // Also hide any individual package popups that might be open
    hideAllPackagePopups();
    
    // Restore body scroll
    document.body.style.overflow = '';
    
    debug('Main package modal hidden');
    
    // Reset transition flag after animation completes
    setTimeout(() => {
      isTransitioning = false;
    }, 300);
  } else {
    debug('Main package modal not found');
  }
}

function handlePackageClicks(e) {
  // Handle main modal trigger clicks first
  if (e.target.matches('[data-packages-modal]')) {
    e.preventDefault();
    e.stopPropagation();
    showMainPackageModal();
    return;
  }
  
  // Handle close button clicks - check both the button and its children
  if (e.target.classList.contains('packages-main-modal-close') || 
      e.target.closest('.packages-main-modal-close')) {
    debug('Close button clicked via handlePackageClicks');
    e.preventDefault();
    e.stopPropagation();
    hideMainPackageModal();
    return;
  }
  
  // Handle main modal background clicks to close
  if (e.target.classList.contains('packages-main-modal') && 
      !e.target.closest('.packages-main-modal-content')) {
    e.preventDefault();
    e.stopPropagation();
    hideMainPackageModal();
    return;
  }
  
  // Handle package popup close button clicks (check both target and closest)
  if (e.target.classList.contains('package-popup-close') || 
      e.target.closest('.package-popup-close')) {
    e.preventDefault();
    e.stopPropagation();
    hideAllPackagePopups();
    return;
  }
  
  // Handle expertise popup close button clicks
  if (e.target.classList.contains('expertise-popup-close')) {
    e.preventDefault();
    e.stopPropagation();
    hideAllExpertisePopups();
    return;
  }
  
  // Handle package popup background clicks to close
  if (e.target.classList.contains('package-popup')) {
    e.preventDefault();
    e.stopPropagation();
    hideAllPackagePopups();
    return;
  }
  
  // Handle expertise popup background clicks to close
  if (e.target.classList.contains('expertise-popup')) {
    e.preventDefault();
    e.stopPropagation();
    hideAllExpertisePopups();
    return;
  }
  
  // Handle package card clicks BEFORE checking modal content
  if (e.target.closest('.package-interactive')) {
    const packageCard = e.target.closest('.package-interactive');
    const packageId = packageCard.dataset.package;
    const popup = document.querySelector(`[data-package-popup="${packageId}"]`);
    
    if (popup) {
      e.preventDefault();
      e.stopPropagation();
      showPackagePopup(popup);
    }
    return;
  }
  
  // Don't handle clicks inside popup content
  if (e.target.closest('.package-popup-content') || 
      e.target.closest('.expertise-popup-content')) {
    return;
  }
  
  // Handle expertise card clicks
  if (e.target.closest('.expertise-interactive')) {
    const expertiseCard = e.target.closest('.expertise-interactive');
    const expertiseId = expertiseCard.dataset.expertise;
    const popup = document.querySelector(`[data-expertise-popup="${expertiseId}"]`);
    
    if (popup) {
      e.preventDefault();
      e.stopPropagation();
      showExpertisePopup(popup);
    }
  }
}

function showPackagePopup(popup) {
  // Prevent multiple popups from opening
  if (isPopupOpen) return;
  
  // Hide any other open popups first
  hideAllPackagePopups();
  
  // Show the selected popup with higher z-index to appear above main modal
  popup.classList.add('active');
  popup.style.zIndex = '1100'; // Higher than main modal's 1000
  isPopupOpen = true;
  
  debug('Package popup shown:', popup.dataset.packagePopup);
}

function hideAllPackagePopups() {
  const allPopups = document.querySelectorAll('.package-popup');
  allPopups.forEach(popup => {
    popup.classList.remove('active');
    popup.style.zIndex = ''; // Reset z-index
  });
  
  // Reset flag
  isPopupOpen = false;
  
  // Only restore body scroll if main modal is also closed
  if (!isMainPackageModalOpen) {
    document.body.style.overflow = '';
  }
}

function showExpertisePopup(popup) {
  // Prevent multiple popups from opening
  if (isPopupOpen) return;
  
  // Hide any other open popups first
  hideAllExpertisePopups();
  hideAllPackagePopups();
  
  // Show the selected popup
  popup.classList.add('active');
  isPopupOpen = true;
  
  // Prevent body scroll when popup is open
  document.body.style.overflow = 'hidden';
  
  debug('Expertise popup shown:', popup.dataset.expertisePopup);
}

function hideAllExpertisePopups() {
  const allPopups = document.querySelectorAll('.expertise-popup');
  allPopups.forEach(popup => {
    popup.classList.remove('active');
  });
  
  // Reset flag
  isPopupOpen = false;
  
  // Restore body scroll
  document.body.style.overflow = '';
}

// Work popup functionality
export function initializeWorkPopups() {
  // Remove any existing event listeners first
  document.removeEventListener('click', handleWorkClicks);
  
  // Add global click handler for work interactions
  document.addEventListener('click', handleWorkClicks);
}

function handleWorkClicks(e) {
  // Handle work popup close button clicks
  if (e.target.classList.contains('work-popup-close')) {
    e.preventDefault();
    e.stopPropagation();
    hideAllWorkPopups();
    return;
  }
  
  // Handle work popup background clicks to close
  if (e.target.classList.contains('work-popup')) {
    e.preventDefault();
    e.stopPropagation();
    hideAllWorkPopups();
    return;
  }
  
  // Don't handle clicks inside popup content
  if (e.target.closest('.work-popup-content')) {
    return;
  }
  
  // Handle work card clicks (old style)
  if (e.target.closest('.work-item.interactive')) {
    const workCard = e.target.closest('.work-item.interactive');
    const workId = workCard.dataset.workItem;
    const popup = document.querySelector(`[data-work-popup="${workId}"]`);
    
    if (popup) {
      e.preventDefault();
      e.stopPropagation();
      showWorkPopup(popup);
    }
  }
  
  // Handle work row clicks (new tabbed style)
  if (e.target.closest('.work-row.clickable')) {
    const workRow = e.target.closest('.work-row.clickable');
    const workId = workRow.dataset.workItem;
    const popup = document.querySelector(`[data-work-popup="${workId}"]`);
    
    if (popup) {
      e.preventDefault();
      e.stopPropagation();
      showWorkPopup(popup);
    }
  }
  
  // Handle tab switching (mobile)
  if (e.target.closest('.tab-btn')) {
    const tabBtn = e.target.closest('.tab-btn');
    const tabName = tabBtn.dataset.tab;
    
    // Remove active class from all tabs and contents
    document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
    
    // Add active class to clicked tab and corresponding content
    tabBtn.classList.add('active');
    const tabContent = document.querySelector(`[data-tab-content="${tabName}"]`);
    if (tabContent) {
      tabContent.classList.add('active');
    }
    
    debug('Tab switched to:', tabName);
  }
}

function showWorkPopup(popup) {
  // Prevent multiple popups from opening
  if (isWorkPopupOpen) return;
  
  // Hide any other open popups first
  hideAllWorkPopups();
  hideAllPackagePopups();
  hideAllExpertisePopups();
  
  // Show the selected popup
  popup.classList.add('active');
  isWorkPopupOpen = true;
  
  // Add class to container for browsers without :has() support
  const container = popup.closest('.work-popups');
  if (container) {
    container.classList.add('has-active-popup');
  }
  
  // Prevent body scroll when popup is open
  document.body.style.overflow = 'hidden';
  document.body.classList.add('work-modal-open');
  
  debug('Work popup shown:', popup.dataset.workPopup);
}

function hideAllWorkPopups() {
  const allPopups = document.querySelectorAll('.work-popup');
  allPopups.forEach(popup => {
    popup.classList.remove('active');
  });
  
  // Remove active class from containers
  const containers = document.querySelectorAll('.work-popups');
  containers.forEach(container => {
    container.classList.remove('has-active-popup');
  });
  
  // Reset flag
  isWorkPopupOpen = false;
  
  // Restore body scroll
  document.body.style.overflow = '';
  document.body.classList.remove('work-modal-open');
}

// Handle escape key to close popups
export function initializePopupKeyHandlers() {
  document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape') {
      // If individual popup is open, close it but keep main modal open
      if (isPopupOpen) {
        hideAllPackagePopups();
        hideAllExpertisePopups();
        hideAllWorkPopups();
      } else if (isMainPackageModalOpen) {
        // If only main modal is open, close it
        hideMainPackageModal();
      } else {
        // Close any other popups
        hideAllPackagePopups();
        hideAllExpertisePopups();
        hideAllWorkPopups();
      }
    }
  });
}

// Initialize all popup functionality
export function initializePopups() {
  debug('Initializing all popups');
  
  // Clean up existing handlers before reinitializing
  if (closeButtonHandler) {
    document.removeEventListener('click', closeButtonHandler, true);
    closeButtonHandler = null;
    closeButtonHandlerInitialized = false;
  }
  
  // Reset state when reinitializing (important for HTMX navigation)
  isPopupOpen = false;
  isWorkPopupOpen = false;
  isMainPackageModalOpen = false;
  isTransitioning = false;
  
  // Hide any open modals/popups
  const openModals = document.querySelectorAll('.packages-main-modal.active, .package-popup.active, .work-popup.active');
  openModals.forEach(modal => modal.classList.remove('active'));
  
  // Restore body scroll
  document.body.style.overflow = '';
  
  initializePackagePopups();
  initializeWorkPopups();
  initializePopupKeyHandlers();
  
  debug('All popups initialized');
}

// Export these for external use if needed
export { hideAllPackagePopups, hideAllExpertisePopups, hideAllWorkPopups, hideMainPackageModal };