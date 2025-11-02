import { assertEquals, assertInstanceOf, assertThrows } from "@std/assert";
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
		assertEquals(cloned.bytes, f.bytes); // values should be equal
		// Verify it's a different instance by checking if modifying one doesn't affect the other
		f.bytes[0] = 99;
		assertEquals(cloned.bytes[0], 1); // cloned should still have original value
	});

	await t.step("data getter returns bytes", () => {
		const data = new Uint8Array([42, 43, 44]);
		const f = new BytesFrame(data);
		assertEquals(f.data, data);
	});

	await t.step("clone with provided buffer", () => {
		const data = new Uint8Array([1, 2, 3]);
		const f = new BytesFrame(data);
		const buffer = new Uint8Array(5);
		const cloned = f.clone(buffer);
		assertEquals(cloned.bytes, data);
		assertEquals(cloned.bytes.buffer, buffer.buffer); // should use the provided buffer
	});

	await t.step("copyFrom resizes when src is larger", () => {
		const srcData = new Uint8Array([1, 2, 3, 4, 5]);
		const src = new BytesFrame(srcData);
		const dest = new BytesFrame(new Uint8Array(2)); // smaller than src
		dest.copyFrom(src);
		assertEquals(dest.bytes, srcData);
		assertEquals(dest.byteLength, 5);
	});
});