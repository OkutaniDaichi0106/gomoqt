import { assertEquals } from "@std/assert";
import { MockReceiveStream, MockSendStream } from "../webtransport/mock_stream_test.ts";
import { MAX_VARINT8 } from "../webtransport/len.ts";
import {
	MAX_BYTES_LENGTH,
	parseBytes,
	parseString,
	parseStringArray,
	parseVarint,
	readBytes,
	readFull,
	readString,
	readStringArray,
	readVarint,
	writeBytes,
	writeString,
	writeStringArray,
	writeVarint,
} from "./message.ts";

Deno.test("message utilities", async (t) => {
	await t.step("readFull", async (t) => {
		await t.step("reads exactly the requested bytes", async () => {
			const data = new Uint8Array([1, 2, 3, 4, 5]);
			const reader = new MockReceiveStream(0n, data);

			const result = new Uint8Array(3);
			const [n, err] = await readFull(reader, result);
			assertEquals(n, 3);
			assertEquals(err, undefined);
			assertEquals(result, new Uint8Array([1, 2, 3]));
		});

		await t.step("returns EOF when not enough data", async () => {
			const data = new Uint8Array([1, 2]);
			const reader = new MockReceiveStream(0n, data);

			const result = new Uint8Array(5);
			const [n, err] = await readFull(reader, result);
			assertEquals(n, 2);
			assertEquals(err?.name, "EOFError");
		});
	});

	await t.step("writeVarint", async (t) => {
		await t.step("writes single byte varint", async () => {
			const writer = new MockSendStream(0n);
			const [n, err] = await writeVarint(writer, 42);
			assertEquals(n, 1);
			assertEquals(err, undefined);
			assertEquals(writer.getAllWrittenData(), new Uint8Array([42]));
		});

		await t.step("writes two byte varint", async () => {
			const writer = new MockSendStream(0n);
			const [n, err] = await writeVarint(writer, 1000);
			assertEquals(n, 2);
			assertEquals(err, undefined);
			assertEquals(writer.getAllWrittenData(), new Uint8Array([0x43, 0xe8]));
		});

		await t.step("writes four byte varint", async () => {
			const writer = new MockSendStream(0n);
			const [n, err] = await writeVarint(writer, 100000);
			assertEquals(n, 4);
			assertEquals(err, undefined);
		});

		await t.step("writes eight byte varint", async () => {
			const writer = new MockSendStream(0n);
			const [n, err] = await writeVarint(writer, 4294967296); // 2^32 -> 8-byte varint
			assertEquals(n, 8);
			assertEquals(err, undefined);
		});

		await t.step("rejects negative numbers", async () => {
			const writer = new MockSendStream(0n);
			const [n, err] = await writeVarint(writer, -1);
			assertEquals(n, 0);
			assertEquals(err?.message, "Varint cannot be negative");
		});

		await t.step("rejects too large numbers", async () => {
			const writer = new MockSendStream(0n);
			const [n, err] = await writeVarint(writer, Infinity);
			assertEquals(n, 0);
			assertEquals(err?.name, "RangeError");
		});
	});

	await t.step("writeBytes", async (t) => {
		await t.step("writes bytes with length prefix", async () => {
			const writer = new MockSendStream(0n);
			const data = new Uint8Array([1, 2, 3]);
			const [n, err] = await writeBytes(writer, data);
			assertEquals(n, 4); // 1 byte length + 3 bytes data
			assertEquals(err, undefined);
		});
	});

	await t.step("writeString", async (t) => {
		await t.step("writes string with length prefix", async () => {
			const writer = new MockSendStream(0n);
			const [n, err] = await writeString(writer, "hello");
			assertEquals(n, 6); // 1 byte length + 5 bytes data
			assertEquals(err, undefined);
		});
	});

	await t.step("writeStringArray", async (t) => {
		await t.step("writes string array", async () => {
			const writer = new MockSendStream(0n);
			const arr = ["hello", "world"];
			const [n, err] = await writeStringArray(writer, arr);
			assertEquals(err, undefined);
			assertEquals(n > 0, true);
		});
	});

	await t.step("readVarint", async (t) => {
		await t.step("reads single byte varint", async () => {
			const writer = new MockSendStream(0n);
			await writeVarint(writer, 42);
			const reader = new MockReceiveStream(0n, writer.getAllWrittenData());

			const [value, n, err] = await readVarint(reader);
			assertEquals(value, 42);
			assertEquals(n, 1);
			assertEquals(err, undefined);
		});

		await t.step("reads two byte varint", async () => {
			const writer = new MockSendStream(0n);
			await writeVarint(writer, 1000);
			const reader = new MockReceiveStream(0n, writer.getAllWrittenData());

			const [value, n, err] = await readVarint(reader);
			assertEquals(value, 1000);
			assertEquals(n, 2);
			assertEquals(err, undefined);
		});

		await t.step("reads four byte varint", async () => {
			const writer = new MockSendStream(0n);
			await writeVarint(writer, 100000);
			const reader = new MockReceiveStream(0n, writer.getAllWrittenData());

			const [value, n, err] = await readVarint(reader);
			assertEquals(value, 100000);
			assertEquals(n, 4);
			assertEquals(err, undefined);
		});

		await t.step("reads eight byte varint", async () => {
			const writer = new MockSendStream(0n);
			await writeVarint(writer, 4294967296); // 2^32 -> 8-byte varint
			const reader = new MockReceiveStream(0n, writer.getAllWrittenData());

			const [value, n, err] = await readVarint(reader);
			assertEquals(value, 4294967296);
			assertEquals(n, 8);
			assertEquals(err, undefined);
		});
	});

	await t.step("readBytes", async (t) => {
		await t.step("reads bytes with length prefix", async () => {
			const writer = new MockSendStream(0n);
			const data = new Uint8Array([1, 2, 3]);
			await writeBytes(writer, data);
			const reader = new MockReceiveStream(0n, writer.getAllWrittenData());

			const [result, n, err] = await readBytes(reader);
			assertEquals(result, data);
			assertEquals(n, 4);
			assertEquals(err, undefined);
		});

		await t.step("rejects too large length", async () => {
			const writer = new MockSendStream(0n);
			// Write a varint with length > MAX_BYTES_LENGTH
			await writeVarint(writer, MAX_BYTES_LENGTH + 1);
			const reader = new MockReceiveStream(0n, writer.getAllWrittenData());

			const [result, n, err] = await readBytes(reader);
			assertEquals(result.length, 0);
			assertEquals(err?.message, "Bytes length exceeds maximum limit");
		});
	});

	await t.step("readString", async (t) => {
		await t.step("reads string with length prefix", async () => {
			const writer = new MockSendStream(0n);
			await writeString(writer, "hello");
			const reader = new MockReceiveStream(0n, writer.getAllWrittenData());

			const [result, n, err] = await readString(reader);
			assertEquals(result, "hello");
			assertEquals(n, 6);
			assertEquals(err, undefined);
		});
	});

	await t.step("readStringArray", async (t) => {
		await t.step("reads string array", async () => {
			const writer = new MockSendStream(0n);
			const arr = ["hello", "world"];
			await writeStringArray(writer, arr);
			const reader = new MockReceiveStream(0n, writer.getAllWrittenData());

			const [result, n, err] = await readStringArray(reader);
			assertEquals(result, arr);
			assertEquals(err, undefined);
		});

		await t.step("rejects too large count", async () => {
			const writer = new MockSendStream(0n);
			await writeVarint(writer, MAX_BYTES_LENGTH + 1);
			const reader = new MockReceiveStream(0n, writer.getAllWrittenData());

			const [result, n, err] = await readStringArray(reader);
			assertEquals(result.length, 0);
			assertEquals(err?.message, "String array count exceeds maximum limit");
		});
	});

	await t.step("parseVarint", async (t) => {
		await t.step("parses single byte varint", () => {
			const buf = new Uint8Array([42]);
			const [value, n] = parseVarint(buf, 0);
			assertEquals(value, 42);
			assertEquals(n, 1);
		});

		await t.step("parses two byte varint", () => {
			const buf = new Uint8Array([0x43, 0xe8]);
			const [value, n] = parseVarint(buf, 0);
			assertEquals(value, 1000);
			assertEquals(n, 2);
		});

		await t.step("parses four byte varint", () => {
			const buf = new Uint8Array([0x80, 0x01, 0x86, 0xa0]);
			const [value, n] = parseVarint(buf, 0);
			assertEquals(value, 100000);
			assertEquals(n, 4);
		});

		await t.step("parses eight byte varint", () => {
			const writer = new MockSendStream(0n);
			writeVarint(writer, 4294967296);
			const buf = writer.getAllWrittenData();
			const [value, n] = parseVarint(buf, 0);
			assertEquals(value, 4294967296);
			assertEquals(n, 8);
		});
	});

	await t.step("parseBytes", async (t) => {
		await t.step("parses bytes with length prefix", () => {
			const data = new Uint8Array([3, 1, 2, 3]);
			const [result, n] = parseBytes(data, 0);
			assertEquals(result, new Uint8Array([1, 2, 3]));
			assertEquals(n, 4);
		});
	});

	await t.step("parseString", async (t) => {
		await t.step("parses string with length prefix", () => {
			const data = new Uint8Array([5, 104, 101, 108, 108, 111]); // "hello"
			const [result, n] = parseString(data, 0);
			assertEquals(result, "hello");
			assertEquals(n, 6);
		});
	});

	await t.step("parseStringArray", async (t) => {
		await t.step("parses string array", () => {
			const data = new Uint8Array([
				2, // count
				5,
				104,
				101,
				108,
				108,
				111, // "hello"
				5,
				119,
				111,
				114,
				108,
				100, // "world"
			]);
			const [result, n] = parseStringArray(data, 0);
			assertEquals(result, ["hello", "world"]);
			assertEquals(n, 13);
		});
	});
});
