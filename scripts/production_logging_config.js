// Production logging configuration
// This file should be included to override console methods in production

(function() {
  'use strict';
  
  // Check if we're in production mode
  const isProduction = window.location.hostname !== 'localhost' && 
                       window.location.hostname !== '127.0.0.1' &&
                       !window.location.hostname.includes('.local');
  
  if (isProduction || window.debugLogging === false) {
    // Store original console methods for critical errors
    const originalError = console.error;
    const originalWarn = console.warn;
    
    // Override console methods
    console.log = function() {}; // Silence all logs
    console.debug = function() {}; // Silence debug
    console.info = function() {}; // Silence info
    
    // Keep error and warn but filter them
    console.error = function(...args) {
      // Only log critical errors, not development errors
      const message = args.join(' ');
      if (!message.includes('Mermaid') && 
          !message.includes('Failed to fetch') &&
          !message.includes('navigation')) {
        originalError.apply(console, args);
      }
    };
    
    console.warn = function(...args) {
      // Only log important warnings
      const message = args.join(' ');
      if (!message.includes('Cleaning up stuck request') &&
          !message.includes('debounced')) {
        originalWarn.apply(console, args);
      }
    };
  }
})();