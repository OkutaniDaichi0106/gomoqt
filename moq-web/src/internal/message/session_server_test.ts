import { assertEquals } from "@std/assert";
import { SessionServerMessage } from "./session_server.ts";
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

Deno.test("SessionServerMessage - encode/decode roundtrip - multiple scenarios", async (t) => {
	const testCases = {
		"with extensions": {
			version: 1,
			extensions: new Map<number, Uint8Array>([
				[1, new Uint8Array([1, 2, 3])],
				[2, new Uint8Array([4, 5, 6])],
			]),
		},
		"without extensions": {
			version: 1,
			extensions: new Map<number, Uint8Array>(),
		},
		"different version": {
			version: 2,
			extensions: new Map<number, Uint8Array>([[1, new Uint8Array([7, 8, 9])]]),
		},
		"single extension": {
			version: 1,
			extensions: new Map<number, Uint8Array>([[100, new Uint8Array([0xff, 0xfe, 0xfd])]]),
		},
	};

	for (const [caseName, input] of Object.entries(testCases)) {
		await t.step(caseName, async () => {
			const { writer, reader } = createTestStreams();

			const message = new SessionServerMessage(input);
			const encodeErr = await message.encode(writer);
			assertEquals(encodeErr, undefined, `encode failed for ${caseName}`);

			// Close writer to signal end of stream
			await writer.close();

			const decodedMessage = new SessionServerMessage({});
			const decodeErr = await decodedMessage.decode(reader);
			assertEquals(decodeErr, undefined, `decode failed for ${caseName}`);
			assertEquals(decodedMessage.version, input.version, `version mismatch for ${caseName}`);
			assertEquals(decodedMessage.extensions, input.extensions, `extensions mismatch for ${caseName}`);
		});
	}
});
