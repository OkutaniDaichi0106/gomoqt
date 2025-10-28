import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows } from "../../deps.ts";
import { SubscribeMessage } from './subscribe.ts';
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
      assertEquals(encodeErr, undefined);

      // Close writer to signal end of stream
      await writer.close();

      const decodedMessage = new SubscribeMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      assertEquals(decodeErr, undefined);
      assertEquals(decodedMessage.subscribeId, subscribeId);
      assertEquals(decodedMessage.broadcastPath, broadcastPath);
      assertEquals(decodedMessage.trackName, trackName);
      assertEquals(decodedMessage.trackPriority, trackPriority);
      assertEquals(decodedMessage.minGroupSequence, minGroupSequence);
      assertEquals(decodedMessage.maxGroupSequence, maxGroupSequence);
    } finally {
      await cleanup();
    }
  });
});
