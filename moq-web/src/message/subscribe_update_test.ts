import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows } from "../../deps.ts";
import { SubscribeUpdateMessage } from './subscribe_update.ts';
import { Writer, Reader } from '../io';
import { createIsolatedStreams } from './test-utils.test';

describe('SubscribeUpdateMessage', () => {
  it('should encode and decode', async () => {
    const trackPriority = 1;
    const minGroupSequence = 2n;
    const maxGroupSequence = 3n;

    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      const message = new SubscribeUpdateMessage({
        trackPriority,
        minGroupSequence,
        maxGroupSequence
      });
      const encodeErr = await message.encode(writer);
      assertEquals(encodeErr, undefined);

      // Close writer to signal end of stream
      await writer.close();

      const decodedMessage = new SubscribeUpdateMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      assertEquals(decodeErr, undefined);
      assertEquals(decodedMessage.trackPriority, trackPriority);
      assertEquals(decodedMessage.minGroupSequence, minGroupSequence);
      assertEquals(decodedMessage.maxGroupSequence, maxGroupSequence);
    } finally {
      await cleanup();
    }
  });
});
