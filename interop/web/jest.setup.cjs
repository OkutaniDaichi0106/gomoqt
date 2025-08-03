// Jest setup file for Web Streams polyfill
require('web-streams-polyfill/dist/polyfill.js');

// Also ensure TextEncoder/TextDecoder are available
if (typeof TextEncoder === 'undefined') {
  global.TextEncoder = require('util').TextEncoder;
}
if (typeof TextDecoder === 'undefined') {
  global.TextDecoder = require('util').TextDecoder;
}

// Suppress console.log to speed up test execution
const originalConsoleLog = console.log;
console.log = (...args) => {
  // Only output important error logs in test environment
  if (process.env.NODE_ENV === 'test' && !process.env.VERBOSE_TESTS) {
    return;
  }
  originalConsoleLog.apply(console, args);
};
