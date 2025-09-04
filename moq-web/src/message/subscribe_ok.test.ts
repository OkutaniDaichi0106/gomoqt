import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { SubscribeOkMessage } from './subscribe_ok';
import { Writer, Reader } from '../io';
import { createIsolatedStreams } from './test-utils.test';

describe('SubscribeOkMessage', () => {
  it('should encode and decode', async () => {
    const groupPeriod = 1;

    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      const message = new SubscribeOkMessage({ groupPeriod });
      const encodeErr = await message.encode(writer);
      expect(encodeErr).toBeUndefined();

      // Close writer to signal end of stream
      await writer.close();

      const decodedMessage = new SubscribeOkMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      expect(decodeErr).toBeUndefined();
      expect(decodedMessage.groupPeriod).toEqual(groupPeriod);
    } finally {
      await cleanup();
    }
  });
});
