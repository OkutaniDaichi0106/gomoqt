import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows } from "../../deps.ts";
import { SessionUpdateMessage } from './session_update.ts';
import { Writer, Reader } from '../io';
import { createIsolatedStreams } from './test-utils.test';

describe('SessionUpdateMessage', () => {
  it('should encode and decode', async () => {
    const bitrate = 1000;

    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      const message = new SessionUpdateMessage({ bitrate });
      const encodeErr = await message.encode(writer);
      assertEquals(encodeErr, undefined);

      // Close writer to signal end of stream
      await writer.close();

      const decodedMessage = new SessionUpdateMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      assertEquals(decodeErr, undefined);
      assertEquals(decodedMessage.bitrate, bitrate);
    } finally {
      await cleanup();
    }
  });
});
