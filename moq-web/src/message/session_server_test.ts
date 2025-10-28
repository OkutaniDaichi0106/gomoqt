import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows } from "../../deps.ts";
import { SessionServerMessage } from './session_server.ts';
import { Writer, Reader } from '../io';
import { Version } from '../internal/version';
import { Extensions } from '../internal/extensions';
import { createIsolatedStreams } from './test-utils.test';

describe('SessionServerMessage', () => {
  it('should encode and decode', async () => {
    const version = 1n;
    const extensions = new Extensions();
  extensions.addBytes(1, new Uint8Array([1, 2, 3]));
  extensions.addBytes(2, new Uint8Array([4, 5, 6]));

    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      const message = new SessionServerMessage({ version, extensions });
      const encodeErr = await message.encode(writer);
      assertEquals(encodeErr, undefined);

      // Close writer to signal end of stream
      await writer.close();

      const decodedMessage = new SessionServerMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      assertEquals(decodeErr, undefined);
      assertEquals(decodedMessage.version, version);
      assertEquals(decodedMessage.extensions, extensions);
    } finally {
      await cleanup();
    }
  });
});
