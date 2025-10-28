import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { AnnounceMessage } from './announce';
import { Writer, Reader } from '../webtransport';
import { createIsolatedStreams } from './test-utils.test';

describe('AnnounceMessage', () => {
  it('should encode and decode', async () => {
    const suffix = 'test';
    const active = true;

    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      const message = new AnnounceMessage({ suffix, active });
      const encodeErr = await message.encode(writer);
      expect(encodeErr).toBeUndefined();

      // Close writer to signal end of stream
      await writer.close();

      const decodedMessage = new AnnounceMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      expect(decodeErr).toBeUndefined();
      expect(decodedMessage.suffix).toEqual(suffix);
      expect(decodedMessage.active).toEqual(active);
    } finally {
      await cleanup();
    }
  });
});
