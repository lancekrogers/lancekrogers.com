// Debug utilities with conditional logging based on configuration
let debugLogging = false; // Will be set from server config

export function setDebugLogging(enabled) {
  debugLogging = enabled;
}

export function debug(...args) {
  // Check both module variable and global window variable
  if (debugLogging || window.debugLogging) {
    console.log(...args);
  }
}

export function debugError(...args) {
  // Check both module variable and global window variable
  if (debugLogging || window.debugLogging) {
    console.error(...args);
  }
}