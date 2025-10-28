import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { AnnouncePleaseMessage } from './announce_please';
import { Writer, Reader } from '../io';
import { createIsolatedStreams } from './test-utils.test';

describe('AnnouncePleaseMessage', () => {
  it('should encode and decode', async () => {
    const prefix = 'test';

    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      const message = new AnnouncePleaseMessage({ prefix });
      const encodeErr = await message.encode(writer);
      expect(encodeErr).toBeUndefined();

      // Close writer to signal end of stream
      await writer.close();

      const decodedMessage = new AnnouncePleaseMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      expect(decodeErr).toBeUndefined();
      expect(decodedMessage.prefix).toEqual(prefix);
    } finally {
      await cleanup();
    }
  });
});
