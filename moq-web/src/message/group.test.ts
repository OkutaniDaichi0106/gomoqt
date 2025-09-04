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
      const msg = new GroupMessage({ subscribeId, sequence });
      const encodeErr = await msg.encode(writer);
      expect(encodeErr).toBeUndefined();

      // Close writer to signal end of stream
      await writer.close();

      const decodedMsg = new GroupMessage({});
      const decodeErr = await decodedMsg.decode(reader);
      expect(decodeErr).toBeUndefined();
      expect(decodedMsg.subscribeId).toEqual(subscribeId);
      expect(decodedMsg.sequence).toEqual(sequence);
    } finally {
      await cleanup();
    }
  });
});
