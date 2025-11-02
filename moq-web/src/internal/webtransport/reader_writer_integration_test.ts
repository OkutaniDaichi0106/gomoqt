import { assertEquals } from "@std/assert";
import { SendStream } from "./send_stream.ts";
import { ReceiveStream } from "./receive_stream.ts";

Deno.test("SendStream/ReceiveStream - integration", async (t) => {
	await t.step("varint single byte roundtrip", async () => {
		const testValue = 42n;
		const { readable, writable } = new TransformStream<Uint8Array, Uint8Array>();
		const writer = new SendStream({ stream: writable, streamId: 0n });
		const reader = new ReceiveStream({ stream: readable, streamId: 0n });

		const writePromise = (async () => {
			writer.writeBigVarint(testValue);
			const flushError = await writer.flush();
			assertEquals(flushError, undefined);
			await writer.close();
		})();

		const [readValue, err] = await reader.readBigVarint();
		assertEquals(err, undefined);
		assertEquals(readValue, testValue);

		await writePromise;
	});

	await t.step("varint two byte roundtrip", async () => {
		const testValue = 300n;
		const { readable, writable } = new TransformStream<Uint8Array, Uint8Array>();
		const writer = new SendStream({ stream: writable, streamId: 0n });
		const reader = new ReceiveStream({ stream: readable, streamId: 0n });

		const writePromise = (async () => {
			writer.writeBigVarint(testValue);
			const flushError = await writer.flush();
			assertEquals(flushError, undefined);
			await writer.close();
		})();

		const [readValue, err] = await reader.readBigVarint();
		assertEquals(err, undefined);
		assertEquals(readValue, testValue);

		await writePromise;
	});

	await t.step("string array roundtrip", async () => {
		const testValue = ["hello", "world", "test"];
		const { readable, writable } = new TransformStream<Uint8Array, Uint8Array>();
		const writer = new SendStream({ stream: writable, streamId: 0n });
		const reader = new ReceiveStream({ stream: readable, streamId: 0n });

		const writePromise = (async () => {
			writer.writeStringArray(testValue);
			const flushError = await writer.flush();
			assertEquals(flushError, undefined);
			await writer.close();
		})();

		const [readValue, err] = await reader.readStringArray();
		assertEquals(err, undefined);
		assertEquals(readValue, testValue);

		await writePromise;
	});

	await t.step("empty string array roundtrip", async () => {
		const testValue: string[] = [];
		const { readable, writable } = new TransformStream<Uint8Array, Uint8Array>();
		const writer = new SendStream({ stream: writable, streamId: 0n });
		const reader = new ReceiveStream({ stream: readable, streamId: 0n });

		const writePromise = (async () => {
			writer.writeStringArray(testValue);
			const flushError = await writer.flush();
			assertEquals(flushError, undefined);
			await writer.close();
		})();

		const [readValue, err] = await reader.readStringArray();
		assertEquals(err, undefined);
		assertEquals(readValue, testValue);

		await writePromise;
	});

	await t.step("string roundtrip", async () => {
		const testValue = "hello world";
		const { readable, writable } = new TransformStream<Uint8Array, Uint8Array>();
		const writer = new SendStream({ stream: writable, streamId: 0n });
		const reader = new ReceiveStream({ stream: readable, streamId: 0n });

		const writePromise = (async () => {
			writer.writeString(testValue);
			const flushError = await writer.flush();
			assertEquals(flushError, undefined);
			await writer.close();
		})();

		const [readValue, err] = await reader.readString();
		assertEquals(err, undefined);
		assertEquals(readValue, testValue);

		await writePromise;
	});

	await t.step("uint8 roundtrip", async () => {
		const testValue = 123;
		const { readable, writable } = new TransformStream<Uint8Array, Uint8Array>();
		const writer = new SendStream({ stream: writable, streamId: 0n });
		const reader = new ReceiveStream({ stream: readable, streamId: 0n });

		const writePromise = (async () => {
			writer.writeUint8(testValue);
			const flushError = await writer.flush();
			assertEquals(flushError, undefined);
			await writer.close();
		})();

		const [readValue, err] = await reader.readUint8();
		assertEquals(err, undefined);
		assertEquals(readValue, testValue);

		await writePromise;
	});

	await t.step("boolean roundtrip", async () => {
		const testValue = true;
		const { readable, writable } = new TransformStream<Uint8Array, Uint8Array>();
		const writer = new SendStream({ stream: writable, streamId: 0n });
		const reader = new ReceiveStream({ stream: readable, streamId: 0n });

		const writePromise = (async () => {
			writer.writeBoolean(testValue);
			const flushError = await writer.flush();
			assertEquals(flushError, undefined);
			await writer.close();
		})();

		const [readValue, err] = await reader.readBoolean();
		assertEquals(err, undefined);
		assertEquals(readValue, testValue);

		await writePromise;
	});

	await t.step("uint8 array roundtrip", async () => {
		const testValue = new Uint8Array([1, 2, 3, 4, 5]);
		const { readable, writable } = new TransformStream<Uint8Array, Uint8Array>();
		const writer = new SendStream({ stream: writable, streamId: 0n });
		const reader = new ReceiveStream({ stream: readable, streamId: 0n });

		const writePromise = (async () => {
			writer.writeUint8Array(testValue);
			const flushError = await writer.flush();
			assertEquals(flushError, undefined);
			await writer.close();
		})();

		const [readValue, err] = await reader.readUint8Array();
		assertEquals(err, undefined);
		assertEquals(readValue, testValue);

		await writePromise;
	});

	await t.step("mixed types sequence", async () => {
		const { readable, writable } = new TransformStream<Uint8Array, Uint8Array>();
		const writer = new SendStream({ stream: writable, streamId: 0n });
		const reader = new ReceiveStream({ stream: readable, streamId: 0n });

		const writePromise = (async () => {
			writer.writeBoolean(true);
			writer.writeBigVarint(123n);
			writer.writeString("test");
			writer.writeUint8Array(new Uint8Array([1, 2, 3]));
			writer.writeStringArray(["a", "b", "c"]);
			const flushError = await writer.flush();
			assertEquals(flushError, undefined);
			await writer.close();
		})();

		const [bool1, err1] = await reader.readBoolean();
		assertEquals(err1, undefined);
		assertEquals(bool1, true);

		const [varint1, err2] = await reader.readBigVarint();
		assertEquals(err2, undefined);
		assertEquals(varint1, 123n);

		const [string1, err3] = await reader.readString();
		assertEquals(err3, undefined);
		assertEquals(string1, "test");

		const [bytes1, err4] = await reader.readUint8Array();
		assertEquals(err4, undefined);
		assertEquals(bytes1, new Uint8Array([1, 2, 3]));

		const [strArray1, err5] = await reader.readStringArray();
		assertEquals(err5, undefined);
		assertEquals(strArray1, ["a", "b", "c"]);

		await writePromise;
	});
});
