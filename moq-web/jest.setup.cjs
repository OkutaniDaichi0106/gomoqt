// Jest setup file for Web Streams polyfill
// Conditionally load web-streams-polyfill
if (typeof globalThis.ReadableStream === 'undefined') {
  try {
    require('web-streams-polyfill/dist/polyfill.js');
  } catch (e) {
    // Fallback if polyfill fails to load
    console.warn('Web Streams polyfill failed to load:', e.message);
  }
}

// Ensure TextEncoder/TextDecoder are available in Node.js environment
if (typeof TextEncoder === 'undefined') {
  global.TextEncoder = require('util').TextEncoder;
}
if (typeof TextDecoder === 'undefined') {
  global.TextDecoder = require('util').TextDecoder;
}

// Suppress console output to speed up test execution and reduce noise
const originalConsoleLog = console.log;
const originalConsoleDebug = console.debug;
const originalConsoleWarn = console.warn;
const originalConsoleInfo = console.info;
const originalConsoleError = console.error;

// Override console.log for cleaner test output
console.log = (...args) => {
  // Only output logs in verbose test mode
  if (process.env.NODE_ENV === 'test' && !process.env.VERBOSE_TESTS) {
    return;
  }
  originalConsoleLog.apply(console, args);
};

// Override console.debug to suppress debug messages during tests
console.debug = (...args) => {
  // Suppress debug logs in test environment unless verbose mode is enabled
  if (process.env.NODE_ENV === 'test' && !process.env.VERBOSE_TESTS) {
    return;
  }
  originalConsoleDebug.apply(console, args);
};

// Override console.info for cleaner output
console.info = (...args) => {
  // Suppress info logs in test environment unless verbose mode is enabled
  if (process.env.NODE_ENV === 'test' && !process.env.VERBOSE_TESTS) {
    return;
  }
  originalConsoleInfo.apply(console, args);
};

// Override console.warn with selective filtering
console.warn = (...args) => {
  // Allow warn logs but filter out debug-level warnings in test mode
  if (process.env.NODE_ENV === 'test' && !process.env.VERBOSE_TESTS) {
    const message = args[0];
    if (typeof message === 'string') {
      // Filter out specific noisy warnings
      if (message.includes('[TrackMux]') || 
          message.includes('Web Streams polyfill')) {
        return;
      }
    }
  }
  originalConsoleWarn.apply(console, args);
};

// Override console.error to suppress error logs in test environment
console.error = (...args) => {
  // Suppress error logs in test environment unless verbose mode is enabled
  if (process.env.NODE_ENV === 'test' && !process.env.VERBOSE_TESTS) {
    return;
  }
  originalConsoleError.apply(console, args);
};
