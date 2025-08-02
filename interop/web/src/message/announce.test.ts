import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { AnnounceMessage } from './announce';
import { Writer, Reader } from '../io';
import { createIsolatedStreams } from './test-utils.test';

describe('AnnounceMessage', () => {
  it('should encode and decode', async () => {
    const suffix = 'test';
    const active = true;

    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      const [encodedMessage, encodeErr] = await AnnounceMessage.encode(writer, suffix, active);
      expect(encodeErr).toBeUndefined();
      expect(encodedMessage).toBeDefined();
      expect(encodedMessage?.suffix).toEqual(suffix);
      expect(encodedMessage?.active).toEqual(active);

      // Close writer to signal end of stream
      await writer.close();

      const [decodedMessage, decodeErr] = await AnnounceMessage.decode(reader);
      expect(decodeErr).toBeUndefined();
      expect(decodedMessage).toBeDefined();
      expect(decodedMessage?.suffix).toEqual(suffix);
      expect(decodedMessage?.active).toEqual(active);
    } finally {
      await cleanup();
    }
  });
});
