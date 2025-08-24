import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { AnnouncePleaseMessage } from './announce_please';
import { Writer, Reader } from '../io';
import { createIsolatedStreams } from './test-utils.test';

describe('AnnouncePleaseMessage', () => {
  it('should encode and decode', async () => {
    const prefix = 'test';

    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      const [encodedMessage, encodeErr] = await AnnouncePleaseMessage.encode(writer, prefix);
      expect(encodeErr).toBeUndefined();
      expect(encodedMessage).toBeDefined();
      expect(encodedMessage?.prefix).toEqual(prefix);

      // Close writer to signal end of stream
      await writer.close();

      const [decodedMessage, decodeErr] = await AnnouncePleaseMessage.decode(reader);
      expect(decodeErr).toBeUndefined();
      expect(decodedMessage).toBeDefined();
      expect(decodedMessage?.prefix).toEqual(prefix);
    } finally {
      await cleanup();
    }
  });
});
