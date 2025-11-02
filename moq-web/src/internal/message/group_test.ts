import { assertEquals } from "@std/assert";
import { GroupMessage } from "./group.ts";
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

Deno.test("GroupMessage - encode/decode roundtrip - multiple scenarios", async (t) => {
	const testCases = {
		"normal case": {
			subscribeId: 123,
			sequence: 456n,
		},
		"zero values": {
			subscribeId: 0,
			sequence: 0n,
		},
		"large numbers": {
			subscribeId: 1000000,
			sequence: 2000000n,
		},
		"single values": {
			subscribeId: 1,
			sequence: 1n,
		},
	};

	for (const [caseName, input] of Object.entries(testCases)) {
		await t.step(caseName, async () => {
			const { writer, reader } = createTestStreams();

			const msg = new GroupMessage(input);
			const encodeErr = await msg.encode(writer);
			assertEquals(encodeErr, undefined, `encode failed for ${caseName}`);

			// Close writer to signal end of stream
			await writer.close();

			const decodedMsg = new GroupMessage({});
			const decodeErr = await decodedMsg.decode(reader);
			assertEquals(decodeErr, undefined, `decode failed for ${caseName}`);
			assertEquals(decodedMsg.subscribeId, input.subscribeId, `subscribeId mismatch for ${caseName}`);
			assertEquals(decodedMsg.sequence, input.sequence, `sequence mismatch for ${caseName}`);
		});
	}
});