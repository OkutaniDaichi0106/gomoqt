import { assert, assertEquals } from "@std/assert";
import { SubscribeUpdateMessage } from "./subscribe_update.ts";
import { Buffer } from "@okdaichi/golikejs/bytes";

Deno.test("SubscribeUpdateMessage - encode/decode roundtrip - multiple scenarios", async (t) => {
	const testCases = {
		"normal case": {
			trackPriority: 1,
		},
		"zero values": {
			trackPriority: 0,
		},
		"max priority": {
			trackPriority: 255,
		},
		"mid priority": {
			trackPriority: 10,
		},
	};

	for (const [caseName, input] of Object.entries(testCases)) {
		await t.step(caseName, async () => {
			// Encode using Buffer
			const buffer = Buffer.make(100);
			const message = new SubscribeUpdateMessage(input);
			const encodeErr = await message.encode(buffer);
			assertEquals(encodeErr, undefined, `encode failed for ${caseName}`);

			// Decode from a new buffer with written data
			const readBuffer = Buffer.make(100);
			await readBuffer.write(buffer.bytes());
			const decodedMessage = new SubscribeUpdateMessage({});
			const decodeErr = await decodedMessage.decode(readBuffer);
			assertEquals(decodeErr, undefined, `decode failed for ${caseName}`);
			assertEquals(
				decodedMessage.trackPriority,
				input.trackPriority,
				`trackPriority mismatch for ${caseName}`,
			);
		});
	}

	await t.step(
		"decode should return error when readVarint fails for message length",
		async () => {
			const buffer = Buffer.make(0); // Empty buffer
			const message = new SubscribeUpdateMessage({});
			const err = await message.decode(buffer);
			assertEquals(err !== undefined, true);
		},
	);

	await t.step(
		"decode should return error when reading subscribeId fails",
		async () => {
			const buffer = Buffer.make(10);
			// message length = 5 (varint), but no data
			await buffer.write(new Uint8Array([0x05]));
			const message = new SubscribeUpdateMessage({});
			const err = await message.decode(buffer);
			assert(err !== undefined);
		},
	);
});
