import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { StreamError, StreamErrorCode } from './error';

describe('StreamError', () => {
  describe('constructor', () => {
    it('should create StreamError with required parameters', () => {
      const code: StreamErrorCode = 404;
      const message = 'Not found';
      
      const error = new StreamError(code, message);
      
      expect(error.code).toBe(code);
      expect(error.message).toBe(message);
      expect(error.remote).toBe(false);
      expect(error.name).toBe('Error');
      expect(error instanceof Error).toBe(true);
      expect(error instanceof StreamError).toBe(true);
    });

    it('should create StreamError with remote flag', () => {
      const code: StreamErrorCode = 500;
      const message = 'Internal server error';
      const remote = true;
      
      const error = new StreamError(code, message, remote);
      
      expect(error.code).toBe(code);
      expect(error.message).toBe(message);
      expect(error.remote).toBe(remote);
    });

    it('should default remote to false when not specified', () => {
      const error = new StreamError(200, 'OK');
      
      expect(error.remote).toBe(false);
    });
  });

  describe('prototype chain', () => {
    it('should maintain proper prototype chain', () => {
      const error = new StreamError(123, 'Test error');
      
      expect(error instanceof StreamError).toBe(true);
      expect(error instanceof Error).toBe(true);
      expect(Object.getPrototypeOf(error)).toBe(StreamError.prototype);
    });

    it('should work with instanceof checks after JSON serialization', () => {
      const original = new StreamError(456, 'Original error', true);
      
      // Simulate what might happen during serialization/deserialization
      const recreated = Object.create(StreamError.prototype);
      Object.assign(recreated, original);
      
      expect(recreated instanceof StreamError).toBe(true);
      expect(recreated.code).toBe(456);
      expect(recreated.message).toBe('Original error');
      expect(recreated.remote).toBe(true);
    });
  });

  describe('error codes', () => {
    it('should handle various error codes', () => {
      const testCases: Array<[StreamErrorCode, string]> = [
        [0, 'Success'],
        [400, 'Bad Request'],
        [401, 'Unauthorized'],
        [403, 'Forbidden'],
        [404, 'Not Found'],
        [500, 'Internal Server Error'],
        [503, 'Service Unavailable'],
        [-1, 'Custom negative code'],
        [999999, 'Large error code']
      ];

      testCases.forEach(([code, message]) => {
        const error = new StreamError(code, message);
        expect(error.code).toBe(code);
        expect(error.message).toBe(message);
      });
    });
  });

  describe('message handling', () => {
    it('should handle empty message', () => {
      const error = new StreamError(1, '');
      
      expect(error.message).toBe('');
      expect(error.code).toBe(1);
    });

    it('should handle unicode messages', () => {
      const message = 'エラーが発生しました 🚨';
      const error = new StreamError(2, message);
      
      expect(error.message).toBe(message);
    });

    it('should handle very long messages', () => {
      const longMessage = 'A'.repeat(10000);
      const error = new StreamError(3, longMessage);
      
      expect(error.message).toBe(longMessage);
      expect(error.message.length).toBe(10000);
    });
  });

  describe('remote flag behavior', () => {
    it('should distinguish between local and remote errors', () => {
      const localError = new StreamError(1, 'Local error', false);
      const remoteError = new StreamError(2, 'Remote error', true);
      
      expect(localError.remote).toBe(false);
      expect(remoteError.remote).toBe(true);
    });

    it('should handle boolean conversion correctly', () => {
      const truthyError = new StreamError(1, 'Test', true);
      const falsyError = new StreamError(2, 'Test', false);
      
      expect(!!truthyError.remote).toBe(true);
      expect(!!falsyError.remote).toBe(false);
    });
  });

  describe('error throwing and catching', () => {
    it('should be throwable and catchable', () => {
      const code = 418;
      const message = "I'm a teapot";
      
      expect(() => {
        throw new StreamError(code, message);
      }).toThrow(StreamError);
      
      try {
        throw new StreamError(code, message);
      } catch (error) {
        expect(error).toBeInstanceOf(StreamError);
        if (error instanceof StreamError) {
          expect(error.code).toBe(code);
          expect(error.message).toBe(message);
        }
      }
    });

    it('should preserve stack trace', () => {
      const error = new StreamError(500, 'Stack trace test');
      
      expect(error.stack).toBeDefined();
      expect(typeof error.stack).toBe('string');
      if (error.stack) {
        // Check for the test file name or function in stack trace instead
        expect(error.stack.includes('error.test.ts') || error.stack.includes('Object.<anonymous>')).toBe(true);
      }
    });
  });

  describe('serialization', () => {
    it('should be JSON serializable', () => {
      const error = new StreamError(123, 'Serialization test', true);
      
      const serialized = JSON.stringify(error);
      const parsed = JSON.parse(serialized);
      
      expect(parsed.code).toBe(123);
      expect(parsed.message).toBe('Serialization test');
      expect(parsed.remote).toBe(true);
    });

    it('should handle circular references gracefully', () => {
      const error = new StreamError(456, 'Circular test');
      
      // Add a circular reference
      (error as any).self = error;
      
      expect(() => {
        JSON.stringify(error);
      }).toThrow(); // Should throw due to circular structure
    });
  });
});
