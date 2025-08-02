import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { SubscribeOkMessage } from './subscribe_ok';
import { Writer, Reader } from '../io';
import { createIsolatedStreams } from './test-utils.test';

describe('SubscribeOkMessage', () => {
  it('should encode and decode', async () => {
    const groupOrder = 1n;

    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      const [encodedMessage, encodeErr] = await SubscribeOkMessage.encode(writer, groupOrder);
      expect(encodeErr).toBeUndefined();
      expect(encodedMessage).toBeDefined();
      expect(encodedMessage?.groupOrder).toEqual(groupOrder);

      // Close writer to signal end of stream
      await writer.close();

      const [decodedMessage, decodeErr] = await SubscribeOkMessage.decode(reader);
      expect(decodeErr).toBeUndefined();
      expect(decodedMessage).toBeDefined();
      expect(decodedMessage?.groupOrder).toEqual(groupOrder);
    } finally {
      await cleanup();
    }
  });
});
