import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { SessionUpdateMessage } from './session_update';
import { Writer, Reader } from '../io';
import { createIsolatedStreams } from './test-utils.test';

describe('SessionUpdateMessage', () => {
  it('should encode and decode', async () => {
    const bitrate = 1000n;

    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      const message = new SessionUpdateMessage({ bitrate });
      const encodeErr = await message.encode(writer);
      expect(encodeErr).toBeUndefined();

      // Close writer to signal end of stream
      await writer.close();

      const decodedMessage = new SessionUpdateMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      expect(decodeErr).toBeUndefined();
      expect(decodedMessage.bitrate).toEqual(bitrate);
    } finally {
      await cleanup();
    }
  });
});
