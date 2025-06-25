import { debug } from '../utils/debug.js';

// Form Encryption Animation
let formEncryptionRunning = false;
let encryptionIntervals = [];

export function initializeFormEncryption() {
  // Find any contact forms (main page or modal)
  const contactForms = document.querySelectorAll('.contact-form, .modal-contact-form');
  if (contactForms.length === 0) {
    return;
  }
  
  debug('Initializing form encryption for', contactForms.length, 'forms');
  
  // Remove any existing listeners first
  document.removeEventListener('htmx:beforeRequest', handleFormBeforeRequest);
  document.removeEventListener('htmx:afterRequest', handleFormAfterRequest);
  
  // Add document-level HTMX event listeners (recommended approach)
  document.addEventListener('htmx:beforeRequest', handleFormBeforeRequest);
  document.addEventListener('htmx:afterRequest', handleFormAfterRequest);
  
  // Add a global function for manual testing
  window.testFormEncryption = startFormEncryptionAnimation;
}

function handleFormBeforeRequest(e) {
  // Only handle contact form submissions (both main and modal)
  if (!e.target.classList.contains('contact-form') && !e.target.classList.contains('modal-contact-form')) return;
  
  // Don't start animation if already running
  if (formEncryptionRunning) {
    e.preventDefault();
    return;
  }
  
  // Get form values for validation
  const form = e.target;
  const formData = new FormData(form);
  const values = {
    name: formData.get('name'),
    email: formData.get('email'), 
    message: formData.get('message')
  };
  
  // Client-side validation to prevent server errors
  const validationErrors = [];
  
  if (!values.name || values.name.trim().length < 2) {
    validationErrors.push('Name must be at least 2 characters long');
  }
  
  if (!values.email || !values.email.includes('@') || !values.email.includes('.')) {
    validationErrors.push('Please enter a valid email address');
  }
  
  if (!values.message || values.message.trim().length < 10) {
    validationErrors.push('Message must be at least 10 characters long');
  }
  
  if (validationErrors.length > 0) {
    e.preventDefault();
    
    // Show validation errors - check both possible response divs
    const responseDiv = document.getElementById('contact-response') || document.getElementById('modal-contact-response');
    if (responseDiv) {
      responseDiv.innerHTML = `<div class="alert error">${validationErrors.join('<br>')}</div>`;
    }
    return;
  }
}

function handleFormAfterRequest(e) {
  // Only handle contact form submissions (both main and modal)
  if (!e.target.classList.contains('contact-form') && !e.target.classList.contains('modal-contact-form')) return;
  
  // Check if the response contains a success message (alternative way to detect success)
  const responseText = e.detail.xhr.responseText;
  const hasSuccessMessage = responseText && responseText.includes('Message sent successfully');
  
  // Only run animation on successful submission (check both ways)
  if ((e.detail.successful || hasSuccessMessage) && !formEncryptionRunning) {
    // Clear any HTMX success message first
    const responseDiv = document.getElementById('contact-response') || document.getElementById('modal-contact-response');
    if (responseDiv) {
      responseDiv.innerHTML = '';
    }
    startFormEncryptionAnimation(e.target);
  }
}

function startFormEncryptionAnimation(targetForm) {
  if (formEncryptionRunning) return;
  
  debug('Starting form encryption animation');
  
  // Clear any existing intervals first
  encryptionIntervals.forEach(interval => {
    clearInterval(interval);
  });
  encryptionIntervals = [];
  
  // Use the provided form or find the first visible contact form
  const contactForm = targetForm || document.querySelector('.contact-form:not(.modal-contact-form)') || document.querySelector('.modal-contact-form');
  if (!contactForm) return;
  
  const nameField = contactForm.querySelector('input[name="name"]');
  const emailField = contactForm.querySelector('input[name="email"]');
  const messageField = contactForm.querySelector('textarea[name="message"]');
  const submitButton = contactForm.querySelector('button[type="submit"]');
  
  // Find the appropriate response div based on which form we're animating
  const responseDiv = contactForm.classList.contains('modal-contact-form') 
    ? document.getElementById('modal-contact-response')
    : document.getElementById('contact-response');
  
  if (!nameField || !emailField || !messageField) return;
  
  formEncryptionRunning = true;
  
  // Store original values
  const originalValues = {
    name: nameField.value,
    email: emailField.value,
    message: messageField.value
  };
  
  // Add encrypting class to form
  contactForm.classList.add('form-encrypting');
  
  // Hide any existing response
  if (responseDiv) {
    responseDiv.innerHTML = '';
  }
  
  // Start the encryption cycle
  let cycleCount = 0;
  const maxCycles = 3;
  
  function runEncryptionCycle() {
    cycleCount++;
    
    // Phase 1: Scramble text (500ms)
    scrambleFields([nameField, emailField, messageField], originalValues);
    
    setTimeout(() => {
      // Phase 2: Glitch effect (400ms)
      glitchFields([nameField, emailField, messageField]);
      
      setTimeout(() => {
        // Phase 3: Show "Encrypting..." text (700ms)
        showEncryptingText([nameField, emailField, messageField]);
        
        setTimeout(() => {
          if (cycleCount < maxCycles) {
            // Continue to next cycle
            runEncryptionCycle();
          } else {
            // Final cycle complete - clear fields and show success
            finishEncryption([nameField, emailField, messageField], responseDiv, contactForm);
          }
        }, 700);
        
      }, 400);
      
    }, 500);
  }
  
  // Start the first cycle
  runEncryptionCycle();
}

function scrambleFields(fields, originalValues) {
  fields.forEach(field => {
    field.classList.add('scrambling');
    field.classList.remove('glitching', 'encrypting-text');
    
    const originalValue = originalValues[field.name];
    let scrambledText = '';
    
    // Create scrambled version of the text
    for (let i = 0; i < originalValue.length; i++) {
      const char = originalValue[i];
      if (char === ' ') {
        scrambledText += ' ';
      } else {
        // Random character substitution
        const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()';
        scrambledText += chars[Math.floor(Math.random() * chars.length)];
      }
    }
    
    field.value = scrambledText;
  });
}

function glitchFields(fields) {
  fields.forEach(field => {
    field.classList.add('glitching');
    field.classList.remove('scrambling', 'encrypting-text');
    
    // Add more random corruption
    let glitchedText = field.value;
    const corruptionChars = '█▓▒░▄▀■□▪▫';
    
    // Replace some characters with corruption symbols
    glitchedText = glitchedText.split('').map(char => {
      if (Math.random() < 0.3) {
        return corruptionChars[Math.floor(Math.random() * corruptionChars.length)];
      }
      return char;
    }).join('');
    
    field.value = glitchedText;
  });
}

function showEncryptingText(fields) {
  fields.forEach(field => {
    field.classList.add('encrypting-text');
    field.classList.remove('scrambling', 'glitching');
    
    // Show encrypting status with animated dots
    const messages = ['Encrypting.', 'Encrypting..', 'Encrypting...'];
    let dotCount = 0;
    
    const interval = setInterval(() => {
      field.value = messages[dotCount % messages.length];
      dotCount++;
      
      if (dotCount >= messages.length * 2) {
        clearInterval(interval);
        // Remove from intervals array when done
        const index = encryptionIntervals.indexOf(interval);
        if (index > -1) {
          encryptionIntervals.splice(index, 1);
        }
      }
    }, 120);
    
    // Store interval so it can be cleared if needed
    encryptionIntervals.push(interval);
  });
}

function finishEncryption(fields, responseDiv, contactForm) {
  // Clear any running intervals first
  encryptionIntervals.forEach(interval => {
    clearInterval(interval);
  });
  encryptionIntervals = [];
  
  // Clear all fields completely
  fields.forEach(field => {
    field.value = '';
    field.classList.remove('scrambling', 'glitching', 'encrypting-text');
  });
  
  // Remove encrypting class from form (this should re-enable button via CSS)
  contactForm.classList.remove('form-encrypting');
  
  // Show success message with fade-out
  if (responseDiv) {
    responseDiv.innerHTML = `
      <div class="alert success" id="encryption-success-message">
        <strong>Encrypted message sent successfully!</strong><br>
        I'll get back to you within 24 hours.
      </div>
    `;
    
    // Auto-fade success message after 4 seconds
    setTimeout(() => {
      const successMessage = document.getElementById('encryption-success-message');
      if (successMessage) {
        successMessage.style.transition = 'opacity 1s ease-out';
        successMessage.style.opacity = '0';
        
        // Remove from DOM after fade completes
        setTimeout(() => {
          if (responseDiv && successMessage.parentNode === responseDiv) {
            responseDiv.innerHTML = '';
          }
        }, 1000);
      }
    }, 4000);
  }
  
  // Reset flag - form is ready for new submissions
  formEncryptionRunning = false;
}