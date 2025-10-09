// Vitest setup file
import 'web-streams-polyfill/polyfill';
import { vi } from 'vitest';
import { setupGlobalMocks } from './src/test-utils.test';

// Set up navigator for browser detection
Object.defineProperty(navigator, 'userAgent', {
  value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36',
  writable: true,
});

// Ensure TextEncoder/TextDecoder are available
if (typeof TextEncoder === 'undefined') {
  const { TextEncoder, TextDecoder } = await import('util');
  global.TextEncoder = TextEncoder as any;
  global.TextDecoder = TextDecoder as any;
}

// Polyfill for HTMLElement in Node environment (if needed)
if (typeof HTMLElement === 'undefined') {
  global.HTMLElement = class HTMLElement {
    innerHTML = '';
    textContent = '';
    className = '';
    setAttribute(name: string, value: any) {}
    getAttribute(name: string) { return null; }
    appendChild(child: any) {}
    removeChild(child: any) {}
    querySelector(selector: string) { return null; }
    addEventListener(event: string, handler: any) {}
    dispatchEvent(event: any) { return true; }
  } as any;
}

// Polyfill for EncodedAudioChunk
if (typeof EncodedAudioChunk === 'undefined') {
  global.EncodedAudioChunk = class EncodedAudioChunk {
    type: string;
    timestamp: number;
    data: Uint8Array;
    constructor(options: { type: string; timestamp: number; data: Uint8Array }) {
      this.type = options.type;
      this.timestamp = options.timestamp;
      this.data = options.data;
    }
  } as any;
}

// Make vi globally available (similar to jest)
global.vi = vi;

// Set up global mocks for Web Audio API and other browser APIs
setupGlobalMocks();
