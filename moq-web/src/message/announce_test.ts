import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows } from "../../deps.ts";
import { AnnounceMessage } from './announce.ts';
import { Writer, Reader } from '../io';
import { createIsolatedStreams } from './test-utils.test';

describe('AnnounceMessage', () => {
  it('should encode and decode', async () => {
    const suffix = 'test';
    const active = true;

    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      const message = new AnnounceMessage({ suffix, active });
      const encodeErr = await message.encode(writer);
      assertEquals(encodeErr, undefined);

      // Close writer to signal end of stream
      await writer.close();

      const decodedMessage = new AnnounceMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      assertEquals(decodeErr, undefined);
      assertEquals(decodedMessage.suffix, suffix);
      assertEquals(decodedMessage.active, active);
    } finally {
      await cleanup();
    }
  });
});
