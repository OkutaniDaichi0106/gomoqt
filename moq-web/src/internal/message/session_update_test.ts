import { assertEquals } from "@std/assert";
import { SessionUpdateMessage } from "./session_update.ts";
import { createIsolatedStreams } from "./test-utils_test.ts";

Deno.test("SessionUpdateMessage", async (t) => {
	await t.step("should encode and decode", async () => {
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
