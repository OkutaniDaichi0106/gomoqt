import { assertEquals } from "@std/assert";
import { AnnounceMessage } from "./announce.ts";
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

Deno.test("AnnounceMessage - encode/decode roundtrip - multiple scenarios", async (t) => {
	const testCases = {
		"normal case with active true": {
			suffix: "test",
			active: true,
		},
		"normal case with active false": {
			suffix: "test",
			active: false,
		},
		"empty suffix": {
			suffix: "",
			active: true,
		},
		"long suffix": {
			suffix: "very/long/path/to/broadcast/suffix",
			active: true,
		},
	};

	for (const [caseName, input] of Object.entries(testCases)) {
		await t.step(caseName, async () => {
			const { writer, reader } = createTestStreams();

			const message = new AnnounceMessage(input);
			const encodeErr = await message.encode(writer);
			assertEquals(encodeErr, undefined, `encode failed for ${caseName}`);

			// Close writer to signal end of stream
			await writer.close();

			const decodedMessage = new AnnounceMessage({});
			const decodeErr = await decodedMessage.decode(reader);
			assertEquals(decodeErr, undefined, `decode failed for ${caseName}`);
			assertEquals(decodedMessage.suffix, input.suffix, `suffix mismatch for ${caseName}`);
			assertEquals(decodedMessage.active, input.active, `active mismatch for ${caseName}`);
		});
	}
});
