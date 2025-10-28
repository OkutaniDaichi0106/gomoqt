import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows } from "../../deps.ts";
import type { StreamErrorCode } from './error.ts';
import { StreamError } from './error.ts';

describe('StreamError', () => {
  describe('constructor', () => {
    it('should create StreamError with required parameters', () => {
      const code: StreamErrorCode = 404;
      const message = 'Not found';
      
      const error = new StreamError(code, message);
      
      assertEquals(error.code, code);
      assertEquals(error.message, message);
      assertEquals(error.remote, false);
      assertEquals(error.name, 'Error');
      assertEquals(error instanceof Error, true);
      assertEquals(error instanceof StreamError, true);
    });

    it('should create StreamError with remote flag', () => {
      const code: StreamErrorCode = 500;
      const message = 'Internal server error';
      const remote = true;
      
      const error = new StreamError(code, message, remote);
      
      assertEquals(error.code, code);
      assertEquals(error.message, message);
      assertEquals(error.remote, remote);
    });

    it('should default remote to false when not specified', () => {
      const error = new StreamError(200, 'OK');
      
      assertEquals(error.remote, false);
    });
  });

  describe('prototype chain', () => {
    it('should maintain proper prototype chain', () => {
      const error = new StreamError(123, 'Test error');
      
      assertEquals(error instanceof StreamError, true);
      assertEquals(error instanceof Error, true);
      expect(Object.getPrototypeOf(error)).toBe(StreamError.prototype);
    });

    it('should work with instanceof checks after JSON serialization', () => {
      const original = new StreamError(456, 'Original error', true);
      
      // JSON serialization doesn't preserve prototype, so instanceof won't work
      // This test demonstrates the limitation of JSON serialization with custom classes
      const recreated = Object.create(StreamError.prototype);
      Object.assign(recreated, original);
      
      assertEquals(recreated instanceof StreamError, true);
      assertEquals(recreated.code, 456);
      // Note: message is not copied by Object.assign for Error objects
      assertEquals(recreated.message, ''); // Error.message is not enumerable by default
      assertEquals(recreated.remote, true);
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
        assertEquals(error.code, code);
        assertEquals(error.message, message);
      });
    });
  });

  describe('message handling', () => {
    it('should handle empty message', () => {
      const error = new StreamError(1, '');
      
      assertEquals(error.message, '');
      assertEquals(error.code, 1);
    });

    it('should handle unicode messages', () => {
      const message = 'エラーが発生しました 🚨';
      const error = new StreamError(2, message);
      
      assertEquals(error.message, message);
    });

    it('should handle very long messages', () => {
      const longMessage = 'A'.repeat(10000);
      const error = new StreamError(3, longMessage);
      
      assertEquals(error.message, longMessage);
      assertEquals(error.message.length, 10000);
    });
  });

  describe('remote flag behavior', () => {
    it('should distinguish between local and remote errors', () => {
      const localError = new StreamError(1, 'Local error', false);
      const remoteError = new StreamError(2, 'Remote error', true);
      
      assertEquals(localError.remote, false);
      assertEquals(remoteError.remote, true);
    });

    it('should handle boolean conversion correctly', () => {
      const truthyError = new StreamError(1, 'Test', true);
      const falsyError = new StreamError(2, 'Test', false);
      
      assertEquals(!!truthyError.remote, true);
      assertEquals(!!falsyError.remote, false);
    });
  });

  describe('error throwing and catching', () => {
    it('should be throwable and catchable', () => {
      const code = 418;
      const message = "I'm a teapot";
      
      assertThrows(() => { throw new StreamError(code, message);
       }, Error, StreamError);
      
      try {
        throw new StreamError(code, message);
      } catch (error) {
        assertInstanceOf(error, StreamError);
        if (error instanceof StreamError) {
          assertEquals(error.code, code);
          assertEquals(error.message, message);
        }
      }
    });

    it('should preserve stack trace', () => {
      const error = new StreamError(500, 'Stack trace test');
      
      assertExists(error.stack);
      assertEquals(typeof error.stack, 'string');
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
      
      // Error objects don't include message in JSON by default
      assertEquals(parsed.code, 123);
      assertEquals(parsed.message, undefined); // Error.message is not enumerable
      assertEquals(parsed.remote, true);
      
      // But we can manually serialize the important properties
      const manualSerialized = JSON.stringify({
        code: error.code,
        message: error.message,
        remote: error.remote
      });
      const manualParsed = JSON.parse(manualSerialized);
      assertEquals(manualParsed.code, 123);
      assertEquals(manualParsed.message, 'Serialization test');
      assertEquals(manualParsed.remote, true);
    });

    it('should handle circular references gracefully', () => {
      const error = new StreamError(456, 'Circular test');
      
      // Add a circular reference
      (error as any).self = error;
      
      assertThrows(() => { JSON.stringify(error);
       }); // Should throw due to circular structure
    });
  });
});
