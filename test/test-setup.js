// Jest setup file for DOM testing
// This file runs before each test file

// Polyfill TextEncoder/TextDecoder for JSDOM
const { TextEncoder, TextDecoder } = require('util');
global.TextEncoder = TextEncoder;
global.TextDecoder = TextDecoder;

// Mock HTMX global object
global.htmx = {
  process: jest.fn(),
  on: jest.fn(),
  trigger: jest.fn()
};

// Mock window.performance for animation timing
global.performance = {
  now: jest.fn(() => Date.now())
};

// Mock requestAnimationFrame
global.requestAnimationFrame = jest.fn(cb => setTimeout(cb, 16));

// Mock fetch for API calls
global.fetch = jest.fn(() =>
  Promise.resolve({
    ok: true,
    json: () => Promise.resolve([]),
    headers: new Map()
  })
);

// Reset all mocks before each test
beforeEach(() => {
  jest.clearAllMocks();
});