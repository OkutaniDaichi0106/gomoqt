import { assertEquals } from "@std/assert";
import { AnnounceInitMessage } from "./announce_init.ts";
import { SendStream } from "../webtransport/send_stream.ts";
import { ReceiveStream } from "../webtransport/receive_stream.ts";

/**
 * Helper to create connected send/receive streams for testing encode/decode
 */
function createTestStreams() {
	const { writable, readable } = new TransformStream<Uint8Array>();
	const writer = new SendStream({ stream: writable, transfer: undefined, streamId: 0n });
	const reader = new ReceiveStream({ stream: readable, transfer: undefined, streamId: 0n });
	return { writer, reader };
}

Deno.test("AnnounceInitMessage", async (t) => {
	await t.step("should encode and decode with suffixes", async () => {
		const suffixes = ["suffix1", "suffix2", "suffix3"];

		const { writer, reader } = createTestStreams();

		const message = new AnnounceInitMessage({ suffixes });
		const encodeErr = await message.encode(writer);
		assertEquals(encodeErr, undefined);

		// Close writer to signal end of stream
		await writer.close();

		const decodedMessage = new AnnounceInitMessage({});
		const decodeErr = await decodedMessage.decode(reader);
		assertEquals(decodeErr, undefined);
		assertEquals(decodedMessage.suffixes, suffixes);
	});

	await t.step("should encode and decode without suffixes", async () => {
		const { writer, reader } = createTestStreams();

		const message = new AnnounceInitMessage({});
		const encodeErr = await message.encode(writer);
		assertEquals(encodeErr, undefined);

		// Close writer to signal end of stream
		await writer.close();

		const decodedMessage = new AnnounceInitMessage({});
		const decodeErr = await decodedMessage.decode(reader);
		assertEquals(decodeErr, undefined);
		assertEquals(decodedMessage.suffixes, []);
	});
});
