import { assertEquals } from "../../deps.ts";
import { AnnounceMessage } from './announce.ts';
import { createIsolatedStreams } from './test-utils_test.ts';

Deno.test('AnnounceMessage', async (t) => {
  await t.step('should encode and decode', async () => {
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
