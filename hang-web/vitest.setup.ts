// Vitest setup file
import 'web-streams-polyfill/polyfill';
import { vi } from 'vitest';

// Ensure TextEncoder/TextDecoder are available
if (typeof TextEncoder === 'undefined') {
  const { TextEncoder, TextDecoder } = await import('util');
  global.TextEncoder = TextEncoder;
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

// Make vi globally available (similar to jest)
global.vi = vi;
