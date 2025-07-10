// Jest setup file for Web Streams polyfill
require('web-streams-polyfill/dist/polyfill.js');

// Also ensure TextEncoder/TextDecoder are available
if (typeof TextEncoder === 'undefined') {
  global.TextEncoder = require('util').TextEncoder;
}
if (typeof TextDecoder === 'undefined') {
  global.TextDecoder = require('util').TextDecoder;
}
