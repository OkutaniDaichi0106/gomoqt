import { assertEquals } from "@std/assert";
import { SubscribeOkMessage } from "./subscribe_ok.ts";
import { Buffer } from "@okudai/golikejs/bytes";

Deno.test("SubscribeOkMessage - encode/decode roundtrip", async (t) => {
	await t.step("should encode and decode empty message", async () => {
		const buffer = Buffer.make(10);

		const message = new SubscribeOkMessage({});
		const encodeErr = await message.encode(buffer);
		assertEquals(encodeErr, undefined);

		const readBuffer = Buffer.make(10);
		await readBuffer.write(buffer.bytes());
		const decodedMessage = new SubscribeOkMessage({});
		const decodeErr = await decodedMessage.decode(readBuffer);
		assertEquals(decodeErr, undefined);
	});

	await t.step("messageLength should return 0", () => {
		const message = new SubscribeOkMessage({});
		assertEquals(message.len, 0);
	});

	await t.step("decode should return error when readUint16 fails", async () => {
		const buffer = Buffer.make(0); // Empty buffer causes read error
		const message = new SubscribeOkMessage({});
		const err = await message.decode(buffer);
		assertEquals(err !== undefined, true);
	});

	await t.step("decode should return error when message length mismatch", async () => {
		const buffer = Buffer.make(10);
		// Write a non-zero message length (expect 0 but got non-zero)
		await buffer.write(new Uint8Array([0x00, 0x05])); // msgLen = 5 (big-endian uint16)

		const message = new SubscribeOkMessage({});
		const err = await message.decode(buffer);
		assertEquals(err !== undefined, true);
		assertEquals(err?.message.includes("message length mismatch"), true);
	});
});
