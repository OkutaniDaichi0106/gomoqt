import { assertEquals } from "@std/assert";
import { AnnouncePleaseMessage } from "./announce_please.ts";
import { createIsolatedStreams } from "./test-utils_test.ts";

Deno.test("AnnouncePleaseMessage", async (t) => {
	await t.step("should encode and decode", async () => {
		const prefix = "test";

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
