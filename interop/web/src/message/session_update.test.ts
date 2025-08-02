import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { SessionUpdateMessage } from './session_update';
import { Writer, Reader } from '../io';
import { createIsolatedStreams } from './test-utils.test';

describe('SessionUpdateMessage', () => {
  it('should encode and decode', async () => {
    const bitrate = 1000n;

    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      const [encodedMessage, encodeErr] = await SessionUpdateMessage.encode(writer, bitrate);
      expect(encodeErr).toBeUndefined();
      expect(encodedMessage).toBeDefined();
      expect(encodedMessage?.bitrate).toEqual(bitrate);

      // Close writer to signal end of stream
      await writer.close();

      const [decodedMessage, decodeErr] = await SessionUpdateMessage.decode(reader);
      expect(decodeErr).toBeUndefined();
      expect(decodedMessage).toBeDefined();
      expect(decodedMessage?.bitrate).toEqual(bitrate);
    } finally {
      await cleanup();
    }
  });
});
