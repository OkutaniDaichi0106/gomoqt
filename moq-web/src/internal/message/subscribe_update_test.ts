import { assertEquals } from "@std/assert";
import { SubscribeUpdateMessage } from "./subscribe_update.ts";
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

Deno.test("SubscribeUpdateMessage - encode/decode roundtrip - multiple scenarios", async (t) => {
	const testCases = {
		"normal case": {
			trackPriority: 1,
			minGroupSequence: 2n,
			maxGroupSequence: 3n,
		},
		"zero values": {
			trackPriority: 0,
			minGroupSequence: 0n,
			maxGroupSequence: 0n,
		},
		"large sequence numbers": {
			trackPriority: 255,
			minGroupSequence: 1000000n,
			maxGroupSequence: 2000000n,
		},
		"same min and max sequence": {
			trackPriority: 10,
			minGroupSequence: 100n,
			maxGroupSequence: 100n,
		},
	};

	for (const [caseName, input] of Object.entries(testCases)) {
		await t.step(caseName, async () => {
			const { writer, reader } = createTestStreams();

			const message = new SubscribeUpdateMessage(input);
			const encodeErr = await message.encode(writer);
			assertEquals(encodeErr, undefined, `encode failed for ${caseName}`);

			// Close writer to signal end of stream
			await writer.close();

			const decodedMessage = new SubscribeUpdateMessage({});
			const decodeErr = await decodedMessage.decode(reader);
			assertEquals(decodeErr, undefined, `decode failed for ${caseName}`);
			assertEquals(decodedMessage.trackPriority, input.trackPriority, `trackPriority mismatch for ${caseName}`);
			assertEquals(decodedMessage.minGroupSequence, input.minGroupSequence, `minGroupSequence mismatch for ${caseName}`);
			assertEquals(decodedMessage.maxGroupSequence, input.maxGroupSequence, `maxGroupSequence mismatch for ${caseName}`);
		});
	}
});
