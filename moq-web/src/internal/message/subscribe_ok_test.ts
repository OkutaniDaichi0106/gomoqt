import { assertEquals } from "@std/assert";
import { SubscribeOkMessage } from "./subscribe_ok.ts";
import { SendStream } from "../webtransport/send_stream.ts";
import { ReceiveStream } from "../webtransport/receive_stream.ts";

/**
 * Create a pair of connected streams for testing encode/decode roundtrip
 */
function createTestStreams() {
	const { readable, writable } = new TransformStream<Uint8Array>();
	const writer = new SendStream({ stream: writable, transfer: undefined, streamId: 0n });
	const reader = new ReceiveStream({ stream: readable, transfer: undefined, streamId: 0n });
	return { writer, reader };
}

Deno.test("SubscribeOkMessage - encode/decode roundtrip", async (t) => {
	await t.step("should encode and decode empty message", async () => {
		const { writer, reader } = createTestStreams();

		const message = new SubscribeOkMessage({});
		const encodeErr = await message.encode(writer);
		assertEquals(encodeErr, undefined);

		// Close writer to signal end of stream
		await writer.close();

		const decodedMessage = new SubscribeOkMessage({});
		const decodeErr = await decodedMessage.decode(reader);
		assertEquals(decodeErr, undefined);
	});
});
