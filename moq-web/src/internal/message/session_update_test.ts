import { assertEquals } from "@std/assert";
import { SessionUpdateMessage } from "./session_update.ts";
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

Deno.test("SessionUpdateMessage - encode/decode roundtrip - multiple scenarios", async (t) => {
	const testCases = {
		"normal bitrate": {
			bitrate: 1000,
		},
		"zero bitrate": {
			bitrate: 0,
		},
		"high bitrate": {
			bitrate: 10000000,
		},
		"low bitrate": {
			bitrate: 1,
		},
	};

	for (const [caseName, input] of Object.entries(testCases)) {
		await t.step(caseName, async () => {
			const { writer, reader } = createTestStreams();

			const message = new SessionUpdateMessage(input);
			const encodeErr = await message.encode(writer);
			assertEquals(encodeErr, undefined, `encode failed for ${caseName}`);

			// Close writer to signal end of stream
			await writer.close();

			const decodedMessage = new SessionUpdateMessage({});
			const decodeErr = await decodedMessage.decode(reader);
			assertEquals(decodeErr, undefined, `decode failed for ${caseName}`);
			assertEquals(decodedMessage.bitrate, input.bitrate, `bitrate mismatch for ${caseName}`);
		});
	}
});
