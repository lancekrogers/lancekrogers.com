// Contact Modal Module
import { debug } from '../utils/debug.js';
import { initializeFormEncryption } from './form-encryption.js';

let modal = null;
let isOpen = false;
let initialized = false;

// Initialize modal functionality
export function initializeContactModal() {
  debug('Initializing contact modal', 'already initialized:', initialized);
  
  // Debug DOM state
  debug('Document ready state:', document.readyState);
  debug('Body exists:', document.body !== null);
  debug('All modal overlays:', document.querySelectorAll('.modal-overlay').length);
  
  // Get or create modal
  modal = document.getElementById('contact-modal');
  if (!modal) {
    debug('Contact modal not found in DOM');
    // Extra debugging when modal not found
    const allIds = Array.from(document.querySelectorAll('[id]')).map(el => el.id);
    debug('All element IDs in DOM:', allIds.filter(id => id.includes('modal')));
    return;
  }
  
  // Reset initialization if modal is different
  const modalId = modal.getAttribute('data-modal-id') || 'default';
  if (initialized && modal.getAttribute('data-modal-initialized') === modalId) {
    debug('Contact modal already initialized for this instance');
    return;
  }
  
  // Mark this specific modal as initialized
  modal.setAttribute('data-modal-initialized', modalId);
  
  // Set up close button
  const closeButton = modal.querySelector('.modal-close');
  if (closeButton) {
    closeButton.addEventListener('click', closeModal);
  }
  
  // Close on backdrop click
  modal.addEventListener('click', (e) => {
    if (e.target === modal) {
      closeModal();
    }
  });
  
  // Close on escape key
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && isOpen) {
      closeModal();
    }
  });
  
  // Handle form submission response
  document.addEventListener('htmx:afterSwap', (e) => {
    if (e.detail.target.id === 'modal-contact-response') {
      const response = e.detail.target;
      if (response.textContent.includes('Message sent successfully')) {
        // Don't close immediately - let the encryption animation play
        // The form-encryption module will handle clearing the form
        // We'll close the modal after the encryption animation completes
        setTimeout(() => {
          closeModal();
          response.innerHTML = '';
        }, 8000); // Give enough time for animation + reading the success message
      }
    }
  });
  
  // Mark as initialized
  initialized = true;
  debug('Contact modal initialization complete');
}

// Open modal
export function openModal() {
  if (!modal) {
    debug('Cannot open modal - not initialized');
    return;
  }
  
  debug('Opening contact modal');
  
  // Trigger animation
  requestAnimationFrame(() => {
    modal.classList.add('active');
    isOpen = true;
    
    // Initialize form encryption for the modal form
    initializeFormEncryption();
    
    // Focus first input
    const firstInput = modal.querySelector('input[type="text"]');
    if (firstInput) {
      setTimeout(() => firstInput.focus(), 100);
    }
  });
  
  // Prevent body scroll
  document.body.style.overflow = 'hidden';
}

// Close modal
export function closeModal() {
  if (!modal || !isOpen) return;
  
  debug('Closing contact modal');
  modal.classList.remove('active');
  isOpen = false;
  
  // Wait for animation to complete
  setTimeout(() => {
    // Restore body scroll
    document.body.style.overflow = '';
  }, 300);
}

// Handle contact button clicks
export function handleContactButtonClick(e) {
  const button = e.target.closest('[data-contact-modal]');
  if (button) {
    e.preventDefault();
    openModal();
  }
}

// Track if global listener is set up
let contactButtonListenerSetup = false;

// Set up global listener for contact buttons
export function setupContactButtons() {
  if (!contactButtonListenerSetup) {
    debug('Setting up contact button listeners');
    document.addEventListener('click', handleContactButtonClick);
    contactButtonListenerSetup = true;
  } else {
    debug('Contact button listeners already set up');
  }
}