import { assertEquals, assertInstanceOf, assertNotEquals, assertThrows } from "../deps.ts";
import { BytesFrame } from "./frame.ts";

Deno.test("frame - BytesFrame basic operations", async (t) => {
	await t.step("byteLength reports correctly", () => {
		const data = new Uint8Array([1, 2, 3]);
		const f = new BytesFrame(data);
		assertEquals(f.byteLength, 3);
	});

	await t.step("copyTo into Uint8Array", () => {
		const data = new Uint8Array([10, 20, 30]);
		const f = new BytesFrame(data);
		const dest = new Uint8Array(3);
		f.copyTo(dest);
		assertEquals(dest, data);
	});

	await t.step("copyTo into ArrayBuffer", () => {
		const data = new Uint8Array([7, 8, 9]);
		const f = new BytesFrame(data);
		const destBuf = new ArrayBuffer(3);
		f.copyTo(destBuf);
		assertEquals(new Uint8Array(destBuf), data);
	});

	await t.step("copyTo throws on unsupported dest type", () => {
		const data = new Uint8Array([1]);
		const f = new BytesFrame(data);
		// @ts-expect-error: intentionally passing unsupported type for testing
		assertThrows(() => f.copyTo(123), Error);
	});

	await t.step("copyFrom and clone behavior", () => {
		const srcData = new Uint8Array([5, 6, 7]);
		const src = new BytesFrame(srcData);
		const dest = new BytesFrame(new Uint8Array(3));
		dest.copyFrom(src);
		assertEquals(dest.bytes, srcData);

		const f = new BytesFrame(new Uint8Array([1, 2, 3]));
		const cloned = f.clone();
		assertInstanceOf(cloned, BytesFrame);
		assertEquals(cloned.bytes, f.bytes);
		assertNotEquals(cloned.bytes, f.bytes); // different reference
	});
});
