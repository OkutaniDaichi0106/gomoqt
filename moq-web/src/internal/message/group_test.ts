import { assertEquals } from "@std/assert";
import { GroupMessage } from "./group.ts";
import { createIsolatedStreams } from "./test-utils.test.ts";

Deno.test("GroupMessage encode/decode roundtrip", async () => {
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
Deno.test("GroupMessage", async (t) => {
	await t.step("should encode and decode", async () => {
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
