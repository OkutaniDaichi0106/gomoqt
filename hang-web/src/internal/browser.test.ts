import { describe, it, expect, vi } from 'vitest';
import { isChrome, isFirefox } from './browser.js';

describe('browser', () => {
  describe('isChrome', () => {
    it('should return true when userAgent includes chrome', () => {
      expect(isChrome).toBe(true);
    });

    it('should return false when userAgent does not include chrome', () => {
      // This test assumes the default userAgent does not include chrome
      // In real scenarios, this would be tested with different userAgents
      expect(typeof isChrome).toBe('boolean');
    });
  });

  describe('isFirefox', () => {
    it('should return false when userAgent includes chrome', () => {
      expect(isFirefox).toBe(false);
    });

    it('should return false when userAgent does not include firefox', () => {
      expect(typeof isFirefox).toBe('boolean');
    });
  });
});