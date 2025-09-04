import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { SubscribeMessage } from './subscribe';
import { Writer, Reader } from '../io';
import { createIsolatedStreams } from './test-utils.test';

describe('SubscribeMessage', () => {
  it('should encode and decode', async () => {
    const subscribeId = 123n;
    const broadcastPath = 'path';
    const trackName = 'track';
    const trackPriority = 1;
    const minGroupSequence = 2n;
    const maxGroupSequence = 3n;

    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      const message = new SubscribeMessage({
        subscribeId,
        broadcastPath,
        trackName,
        trackPriority,
        minGroupSequence,
        maxGroupSequence
      });
      const encodeErr = await message.encode(writer);
      expect(encodeErr).toBeUndefined();

      // Close writer to signal end of stream
      await writer.close();

      const decodedMessage = new SubscribeMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      expect(decodeErr).toBeUndefined();
      expect(decodedMessage.subscribeId).toEqual(subscribeId);
      expect(decodedMessage.broadcastPath).toEqual(broadcastPath);
      expect(decodedMessage.trackName).toEqual(trackName);
      expect(decodedMessage.trackPriority).toEqual(trackPriority);
      expect(decodedMessage.minGroupSequence).toEqual(minGroupSequence);
      expect(decodedMessage.maxGroupSequence).toEqual(maxGroupSequence);
    } finally {
      await cleanup();
    }
  });
});
