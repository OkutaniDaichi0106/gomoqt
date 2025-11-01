/**
 * Tests for mock stream implementations
 */

import { assertEquals, assertExists } from "@std/assert";
import { MockStream, MockSendStream, MockReceiveStream } from "./mock_stream_test.ts";

Deno.test("MockSendStream - basic functionality", async () => {
	const mock = new MockSendStream();

	// Test write operations
	mock.writeVarint(42);
	mock.writeString("test");
	mock.writeBoolean(true);

	// Test flush
	await mock.flush();
	assertEquals(mock.flushCalls, 1);

	// Test close
	await mock.close();
	assertEquals(mock.closeCalls, 1);

	// Test reset
	mock.reset();
	assertEquals(mock.flushCalls, 0);
	assertEquals(mock.closeCalls, 0);
});

Deno.test("MockSendStream - error simulation", async () => {
	const mock = new MockSendStream();

	// Simulate flush error
	mock.flushError = new Error("Test error");
	const err = await mock.flush();

	assertExists(err);
	assertEquals(err.message, "Test error");
	assertEquals(mock.flushCalls, 1);
});

Deno.test("MockReceiveStream - basic functionality", async () => {
	const mock = new MockReceiveStream();

	// Setup data
	mock.data = [
		new Uint8Array([5]), // varint value
		new Uint8Array([104, 101, 108, 108, 111]), // "hello"
	];

	// Read varint
	const [value, err1] = await mock.readVarint();
	assertEquals(err1, undefined);
	assertEquals(value, 5);

	// Read string
	const [str, err2] = await mock.readString();
	assertEquals(err2, undefined);
	assertEquals(str, "hello");
});

Deno.test("MockReceiveStream - custom implementations", async () => {
	const mock = new MockReceiveStream();

	// Custom readVarint implementation
	let callCount = 0;
	mock.readVarintImpl = async () => {
		callCount++;
		return [callCount * 10, undefined];
	};

	const [val1] = await mock.readVarint();
	assertEquals(val1, 10);

	const [val2] = await mock.readVarint();
	assertEquals(val2, 20);
});

Deno.test("MockStream - bidirectional stream", async () => {
	const mock = new MockStream(123n);

	// Verify stream ID
	assertEquals(mock.streamId, 123n);
	assertEquals(mock.writable.streamId, 123n);
	assertEquals(mock.readable.streamId, 123n);

	// Test writable side
	await mock.writable.flush();
	assertEquals(mock.writable.flushCalls, 1);

	// Test readable side
	mock.readable.data = [new Uint8Array([42])];
	const [value] = await mock.readable.readVarint();
	assertEquals(value, 42);

	// Test reset
	mock.reset();
	assertEquals(mock.writable.flushCalls, 0);
	// Note: reset() resets the data index, not the data array itself
	// The data array should be manually cleared if needed
});

Deno.test("MockStream - writable and readable are independent", async () => {
	const mock = new MockStream(456n);

	// Write operations don't affect read operations
	mock.writable.writeString("test");
	await mock.writable.flush();

	// Read operations are independent
	mock.readable.data = [new Uint8Array([1, 2, 3])];
	const [data] = await mock.readable.readUint8Array();

	assertEquals(mock.writable.flushCalls, 1);
	assertEquals(data?.length, 3);
});
