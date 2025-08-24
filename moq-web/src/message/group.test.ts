import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { GroupMessage } from './group';
import { Writer, Reader } from '../io';
import { createIsolatedStreams } from './test-utils.test';

describe('GroupMessage', () => {
  it('should encode and decode', async () => {
    const subscribeId = 123n;
    const sequence = 456n;

    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      const [encodedMessage, encodeErr] = await GroupMessage.encode(writer, subscribeId, sequence);
      expect(encodeErr).toBeUndefined();
      expect(encodedMessage).toBeDefined();
      expect(encodedMessage?.subscribeId).toEqual(subscribeId);
      expect(encodedMessage?.sequence).toEqual(sequence);

      // Close writer to signal end of stream
      await writer.close();

      const [decodedMessage, decodeErr] = await GroupMessage.decode(reader);
      expect(decodeErr).toBeUndefined();
      expect(decodedMessage).toBeDefined();
      expect(decodedMessage?.subscribeId).toEqual(subscribeId);
      expect(decodedMessage?.sequence).toEqual(sequence);
    } finally {
      await cleanup();
    }
  });
});
