import { assertEquals } from "@std/assert";
import { GroupMessage } from "./group.ts";
import { createIsolatedStreams } from "./test-utils_test.ts";

Deno.test("GroupMessage - encode/decode roundtrip - multiple scenarios", async (t) => {
	const testCases = {
		"normal case": {
			subscribeId: 123n,
			sequence: 456n,
		},
		"zero values": {
			subscribeId: 0n,
			sequence: 0n,
		},
		"large numbers": {
			subscribeId: 1000000n,
			sequence: 2000000n,
		},
		"single values": {
			subscribeId: 1n,
			sequence: 1n,
		},
	};

	for (const [caseName, input] of Object.entries(testCases)) {
		await t.step(caseName, async () => {
			const { writer, reader, cleanup } = createIsolatedStreams();

			try {
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
			} finally {
				await cleanup();
			}
		});
	}
});