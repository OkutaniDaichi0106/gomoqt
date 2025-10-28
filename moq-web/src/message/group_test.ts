import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows } from "../../deps.ts";
import { GroupMessage } from './group.ts';
import { Writer, Reader } from '../io';
import { createIsolatedStreams } from './test-utils.test';

describe('GroupMessage', () => {
  it('should encode and decode', async () => {
    const subscribeId = 123n;
    const sequence = 456n;

    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      const msg = new GroupMessage({ subscribeId, sequence });
      const encodeErr = await msg.encode(writer);
      assertEquals(encodeErr, undefined);

      // Close writer to signal end of stream
      await writer.close();

      const decodedMsg = new GroupMessage({});
      const decodeErr = await decodedMsg.decode(reader);
      assertEquals(decodeErr, undefined);
      assertEquals(decodedMsg.subscribeId, subscribeId);
      assertEquals(decodedMsg.sequence, sequence);
    } finally {
      await cleanup();
    }
  });
});
