import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { SubscribeUpdateMessage } from './subscribe_update';
import { Writer, Reader } from '../internal/io';
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
      expect(encodeErr).toBeUndefined();

      // Close writer to signal end of stream
      await writer.close();

      const decodedMessage = new SubscribeUpdateMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      expect(decodeErr).toBeUndefined();
      expect(decodedMessage.trackPriority).toEqual(trackPriority);
      expect(decodedMessage.minGroupSequence).toEqual(minGroupSequence);
      expect(decodedMessage.maxGroupSequence).toEqual(maxGroupSequence);
    } finally {
      await cleanup();
    }
  });
});
