// Vitest setup file for Web Streams polyfill
import { TextEncoder as NodeTextEncoder, TextDecoder as NodeTextDecoder } from 'util';

// Conditionally load web-streams-polyfill
if (typeof globalThis.ReadableStream === 'undefined') {
  try {
    await import('web-streams-polyfill/dist/polyfill.js');
  } catch (e: any) {
    // Fallback if polyfill fails to load
    console.warn('Web Streams polyfill failed to load:', e.message);
  }
}

// Ensure TextEncoder/TextDecoder are available in Node.js environment
if (typeof TextEncoder === 'undefined') {
  (global as any).TextEncoder = NodeTextEncoder;
}
if (typeof TextDecoder === 'undefined') {
  (global as any).TextDecoder = NodeTextDecoder;
}

// Suppress console output to speed up test execution and reduce noise
const originalConsoleLog = console.log;
const originalConsoleDebug = console.debug;
const originalConsoleWarn = console.warn;
const originalConsoleInfo = console.info;
const originalConsoleError = console.error;

// Override console.log for cleaner test output
console.log = (...args: any[]) => {
  // Only output logs in verbose test mode
  if (process.env.NODE_ENV === 'test' && !process.env.VERBOSE_TESTS) {
    return;
  }
  originalConsoleLog.apply(console, args);
};

// Override console.debug to suppress debug messages during tests
console.debug = (...args: any[]) => {
  // Suppress debug logs in test environment unless verbose mode is enabled
  if (process.env.NODE_ENV === 'test' && !process.env.VERBOSE_TESTS) {
    return;
  }
  originalConsoleDebug.apply(console, args);
};

// Override console.info for cleaner output
console.info = (...args: any[]) => {
  // Suppress info logs in test environment unless verbose mode is enabled
  if (process.env.NODE_ENV === 'test' && !process.env.VERBOSE_TESTS) {
    return;
  }
  originalConsoleInfo.apply(console, args);
};

// Override console.warn to suppress warnings during tests (optional)
console.warn = (...args: any[]) => {
  // Optionally suppress warnings, or log them in verbose mode
  if (process.env.NODE_ENV === 'test' && !process.env.VERBOSE_TESTS) {
    return;
  }
  originalConsoleWarn.apply(console, args);
};

// Override console.error to suppress errors during tests (optional)
console.error = (...args: any[]) => {
  // Optionally suppress errors, or log them in verbose mode
  if (process.env.NODE_ENV === 'test' && !process.env.VERBOSE_TESTS) {
    return;
  }
  originalConsoleError.apply(console, args);
};
