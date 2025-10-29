import { assertEquals } from "../../deps.ts";
import { SubscribeOkMessage } from './subscribe_ok.ts';
import { createIsolatedStreams } from './test-utils_test.ts';

Deno.test('SubscribeOkMessage', async (t) => {
  await t.step('should encode and decode', async () => {
    const groupPeriod = 1;

    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      const message = new SubscribeOkMessage({ groupPeriod });
      const encodeErr = await message.encode(writer);
      assertEquals(encodeErr, undefined);

      // Close writer to signal end of stream
      await writer.close();

      const decodedMessage = new SubscribeOkMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      assertEquals(decodeErr, undefined);
    } finally {
      await cleanup();
    }
  });
});
