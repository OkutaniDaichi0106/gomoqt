import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { SubscribeUpdateMessage } from './subscribe_update';
import { Writer, Reader } from '../io';
import { createIsolatedStreams } from './test-utils.test';

describe('SubscribeUpdateMessage', () => {
  it('should encode and decode', async () => {
    const trackPriority = 1n;
    const minGroupSequence = 2n;
    const maxGroupSequence = 3n;

    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      const [encodedMessage, encodeErr] = await SubscribeUpdateMessage.encode(
        writer,
        trackPriority,
        minGroupSequence,
        maxGroupSequence
      );
      expect(encodeErr).toBeUndefined();
      expect(encodedMessage).toBeDefined();
      expect(encodedMessage?.trackPriority).toEqual(trackPriority);
      expect(encodedMessage?.minGroupSequence).toEqual(minGroupSequence);
      expect(encodedMessage?.maxGroupSequence).toEqual(maxGroupSequence);

      // Close writer to signal end of stream
      await writer.close();

      const [decodedMessage, decodeErr] = await SubscribeUpdateMessage.decode(reader);
      expect(decodeErr).toBeUndefined();
      expect(decodedMessage).toBeDefined();
      expect(decodedMessage?.trackPriority).toEqual(trackPriority);
      expect(decodedMessage?.minGroupSequence).toEqual(minGroupSequence);
      expect(decodedMessage?.maxGroupSequence).toEqual(maxGroupSequence);
    } finally {
      await cleanup();
    }
  });
});
