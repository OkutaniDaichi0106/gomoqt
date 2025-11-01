import { assertEquals } from "@std/assert";
import { SubscribeUpdateMessage } from "./subscribe_update.ts";
import { createIsolatedStreams } from "./test-utils_deno_test.ts";

Deno.test("SubscribeUpdateMessage", async (t) => {
	await t.step("should encode and decode", async () => {
		const trackPriority = 1;
		const minGroupSequence = 2n;
		const maxGroupSequence = 3n;

		const { writer, reader, cleanup } = createIsolatedStreams();

		try {
			const message = new SubscribeUpdateMessage({
				trackPriority,
				minGroupSequence,
				maxGroupSequence,
			});
			const encodeErr = await message.encode(writer);
			assertEquals(encodeErr, undefined);

			// Close writer to signal end of stream
			await writer.close();

			const decodedMessage = new SubscribeUpdateMessage({});
			const decodeErr = await decodedMessage.decode(reader);
			assertEquals(decodeErr, undefined);
			assertEquals(decodedMessage.trackPriority, trackPriority);
			assertEquals(decodedMessage.minGroupSequence, minGroupSequence);
			assertEquals(decodedMessage.maxGroupSequence, maxGroupSequence);
		} finally {
			await cleanup();
		}
	});
});
Deno.test("SubscribeUpdateMessage", async (t) => {
	await t.step("should encode and decode", async () => {
		const trackPriority = 1;
		const minGroupSequence = 2n;
		const maxGroupSequence = 3n;

		const { writer, reader, cleanup } = createIsolatedStreams();

		try {
			const message = new SubscribeUpdateMessage({
				trackPriority,
				minGroupSequence,
				maxGroupSequence,
			});
			const encodeErr = await message.encode(writer);
			assertEquals(encodeErr, undefined);

			// Close writer to signal end of stream
			await writer.close();

			const decodedMessage = new SubscribeUpdateMessage({});
			const decodeErr = await decodedMessage.decode(reader);
			assertEquals(decodeErr, undefined);
			assertEquals(decodedMessage.trackPriority, trackPriority);
			assertEquals(decodedMessage.minGroupSequence, minGroupSequence);
			assertEquals(decodedMessage.maxGroupSequence, maxGroupSequence);
		} finally {
			await cleanup();
		}
	});
});
