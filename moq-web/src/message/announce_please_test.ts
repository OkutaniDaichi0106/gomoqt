import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows } from "../../deps.ts";
import { AnnouncePleaseMessage } from './announce_please.ts';
import { Writer, Reader } from '../io';
import { createIsolatedStreams } from './test-utils.test';

describe('AnnouncePleaseMessage', () => {
  it('should encode and decode', async () => {
    const prefix = 'test';

    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      const message = new AnnouncePleaseMessage({ prefix });
      const encodeErr = await message.encode(writer);
      assertEquals(encodeErr, undefined);

      // Close writer to signal end of stream
      await writer.close();

      const decodedMessage = new AnnouncePleaseMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      assertEquals(decodeErr, undefined);
      assertEquals(decodedMessage.prefix, prefix);
    } finally {
      await cleanup();
    }
  });
});
