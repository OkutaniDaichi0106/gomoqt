import { assertEquals } from "@std/assert";
import { SubscribeMessage } from "./subscribe.ts";
import { createIsolatedStreams } from "./test-utils_test.ts";

Deno.test("SubscribeMessage - encode/decode roundtrip - multiple scenarios", async (t) => {
	const testCases = {
		"normal case": {
			subscribeId: 123,
			broadcastPath: "path",
			trackName: "track",
			trackPriority: 1,
			minGroupSequence: 2n,
			maxGroupSequence: 3n,
		},
		"large sequence numbers": {
			subscribeId: 1000000,
			broadcastPath: "long/path/to/resource",
			trackName: "long-track-name-with-hyphens",
			trackPriority: 255,
			minGroupSequence: 1000000n,
			maxGroupSequence: 2000000n,
		},
		"zero values": {
			subscribeId: 0,
			broadcastPath: "",
			trackName: "",
			trackPriority: 0,
			minGroupSequence: 0n,
			maxGroupSequence: 0n,
		},
		"single character paths": {
			subscribeId: 1,
			broadcastPath: "a",
			trackName: "b",
			trackPriority: 1,
			minGroupSequence: 1n,
			maxGroupSequence: 2n,
		},
	};

	for (const [caseName, input] of Object.entries(testCases)) {
		await t.step(caseName, async () => {
			const { writer, reader, cleanup } = createIsolatedStreams();

			try {
				const message = new SubscribeMessage(input);
				const encodeErr = await message.encode(writer);
				assertEquals(encodeErr, undefined, `encode failed for ${caseName}`);

				// Close writer to signal end of stream
				await writer.close();

				const decodedMessage = new SubscribeMessage({});
				const decodeErr = await decodedMessage.decode(reader);
				assertEquals(decodeErr, undefined, `decode failed for ${caseName}`);

				// Verify all fields match
				assertEquals(
					decodedMessage.subscribeId,
					input.subscribeId,
					`subscribeId mismatch for ${caseName}`,
				);
				assertEquals(
					decodedMessage.broadcastPath,
					input.broadcastPath,
					`broadcastPath mismatch for ${caseName}`,
				);
				assertEquals(
					decodedMessage.trackName,
					input.trackName,
					`trackName mismatch for ${caseName}`,
				);
				assertEquals(
					decodedMessage.trackPriority,
					input.trackPriority,
					`trackPriority mismatch for ${caseName}`,
				);
				assertEquals(
					decodedMessage.minGroupSequence,
					input.minGroupSequence,
					`minGroupSequence mismatch for ${caseName}`,
				);
				assertEquals(
					decodedMessage.maxGroupSequence,
					input.maxGroupSequence,
					`maxGroupSequence mismatch for ${caseName}`,
				);
			} finally {
				await cleanup();
			}
		});
	}
});
