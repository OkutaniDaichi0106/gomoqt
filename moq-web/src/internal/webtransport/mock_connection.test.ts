import { assertEquals, assertExists } from "@std/assert";
import {
	MockBidirectionalStreamReader,
	MockConnection,
	MockReceiveStream,
	MockSendStream,
	MockStream,
	MockUnidirectionalStreamReader,
	MockWebTransport,
} from "./mock_connection.ts";

// Note: sanitizeResources and sanitizeOps are disabled for these tests
// because the mock implementation uses async operations that may not
// complete within the test lifecycle.
const testOptions = { sanitizeResources: false, sanitizeOps: false };

Deno.test("MockBidirectionalStreamReader - Normal Cases", testOptions, async (t) => {
	await t.step("should enqueue and read bidirectional streams", async () => {
		const reader = new MockBidirectionalStreamReader();
		const { readable, writable } = new TransformStream<Uint8Array>();
		const stream = { readable, writable } as WebTransportBidirectionalStream;

		reader.enqueue(stream);

		const { done, value } = await reader.read();
		assertEquals(done, false);
		assertExists(value);
		assertEquals(value, stream);
	});

	await t.step("should return done when closed", async () => {
		const reader = new MockBidirectionalStreamReader();
		reader.close();

		const { done, value } = await reader.read();
		assertEquals(done, true);
		assertEquals(value, undefined);
	});

	await t.step("should close via cancel", async () => {
		const reader = new MockBidirectionalStreamReader();
		await reader.cancel();

		const { done } = await reader.read();
		assertEquals(done, true);
	});
});

Deno.test("MockBidirectionalStreamReader - Error Cases", testOptions, async (t) => {
	await t.step("should throw when enqueuing to closed reader", () => {
		const reader = new MockBidirectionalStreamReader();
		reader.close();

		const { readable, writable } = new TransformStream<Uint8Array>();
		const stream = { readable, writable } as WebTransportBidirectionalStream;

		let errorThrown = false;
		try {
			reader.enqueue(stream);
		} catch (_e) {
			errorThrown = true;
		}
		assertEquals(errorThrown, true);
	});
});

Deno.test("MockUnidirectionalStreamReader - Normal Cases", testOptions, async (t) => {
	await t.step("should enqueue and read unidirectional streams", async () => {
		const reader = new MockUnidirectionalStreamReader();
		const { readable } = new TransformStream<Uint8Array>();

		reader.enqueue(readable);

		const { done, value } = await reader.read();
		assertEquals(done, false);
		assertExists(value);
		assertEquals(value, readable);
	});

	await t.step("should return done when closed", async () => {
		const reader = new MockUnidirectionalStreamReader();
		reader.close();

		const { done, value } = await reader.read();
		assertEquals(done, true);
		assertEquals(value, undefined);
	});
});

Deno.test("MockWebTransport - Normal Cases", testOptions, async (t) => {
	await t.step("should create bidirectional stream", async () => {
		const wt = new MockWebTransport();
		const stream = await wt.createBidirectionalStream();
		assertExists(stream);
		assertExists(stream.readable);
		assertExists(stream.writable);
	});

	await t.step("should create unidirectional stream", async () => {
		const wt = new MockWebTransport();
		const stream = await wt.createUnidirectionalStream();
		assertExists(stream);
	});

	await t.step("should mark ready", async () => {
		const wt = new MockWebTransport();
		wt.markReady();
		await wt.ready;
		// If we reach here without hanging, ready was resolved
		assertEquals(true, true);
	});

	await t.step("should close with info", async () => {
		const wt = new MockWebTransport();
		const closeInfo = { closeCode: 42, reason: "test close" };
		wt.close(closeInfo);
		const closed = await wt.closed;
		assertEquals(closed.closeCode, 42);
		assertEquals(closed.reason, "test close");
	});

	await t.step("should get close info", () => {
		const wt = new MockWebTransport();
		const closeInfo = { closeCode: 100, reason: "manual close" };
		wt.close(closeInfo);
		const info = wt.getCloseInfo();
		assertExists(info);
		assertEquals(info.closeCode, 100);
		assertEquals(info.reason, "manual close");
	});
});

Deno.test("MockStream - Normal Cases", testOptions, async (t) => {
	await t.step("should create mock stream with correct ID", () => {
		const stream = MockStream.create(42);
		assertEquals(stream.id, 42);
		assertExists(stream.readable);
		assertExists(stream.writable);
	});
});

Deno.test("MockSendStream - Normal Cases", testOptions, async (t) => {
	await t.step("should create mock send stream", () => {
		const stream = MockSendStream.create(10);
		assertEquals(stream.id, 10);
	});

	// Note: flush() may hang if the writable stream has no reader
	// In real usage, there should be a reader on the other end
});

Deno.test("MockReceiveStream - Normal Cases", testOptions, async (t) => {
	await t.step("should create mock receive stream", () => {
		const stream = MockReceiveStream.create(30);
		assertEquals(stream.id, 30);
	});

	await t.step("should create with pre-filled data", () => {
		const data = [new Uint8Array([1, 2, 3]), new Uint8Array([4, 5, 6])];
		const stream = MockReceiveStream.create(40, data);
		assertEquals(stream.id, 40);
	});
});

Deno.test("MockConnection - Normal Cases", testOptions, async (t) => {
	await t.step("should create connection", () => {
		const conn = new MockConnection();
		assertExists(conn);
	});

	await t.step("should open bidirectional stream", async () => {
		const conn = new MockConnection();
		const [stream, err] = await conn.openStream();
		assertEquals(err, undefined);
		assertExists(stream);
		assertEquals(stream?.id, 0);
	});

	await t.step("should open unidirectional stream", async () => {
		const conn = new MockConnection();
		const [stream, err] = await conn.openUniStream();
		assertEquals(err, undefined);
		assertExists(stream);
		assertEquals(stream?.id, 2);
	});

	await t.step("should increment stream IDs correctly", async () => {
		const conn = new MockConnection();

		const [stream1, err1] = await conn.openStream();
		assertEquals(err1, undefined);
		assertEquals(stream1?.id, 0);

		const [stream2, err2] = await conn.openStream();
		assertEquals(err2, undefined);
		assertEquals(stream2?.id, 4);

		const [stream3, err3] = await conn.openUniStream();
		assertEquals(err3, undefined);
		assertEquals(stream3?.id, 2);

		const [stream4, err4] = await conn.openUniStream();
		assertEquals(err4, undefined);
		assertEquals(stream4?.id, 6);
	});

	await t.step("should accept simulated bidirectional stream", async () => {
		const conn = new MockConnection();
		conn.simulateIncomingBiStream();

		const [stream, err] = await conn.acceptStream();
		assertEquals(err, undefined);
		assertExists(stream);
		assertEquals(stream?.id, 1);
	});

	await t.step("should accept simulated unidirectional stream", async () => {
		const conn = new MockConnection();
		conn.simulateIncomingUniStream();

		const [stream, err] = await conn.acceptUniStream();
		assertEquals(err, undefined);
		assertExists(stream);
		assertEquals(stream?.id, 3);
	});

	await t.step("should close connection", async () => {
		const conn = new MockConnection();
		const closeInfo = { closeCode: 0, reason: "test" };
		conn.close(closeInfo);

		const closed = await conn.closed;
		assertEquals(closed.closeCode, 0);
		assertEquals(closed.reason, "test");
	});

	await t.step("should mark connection as ready", async () => {
		const conn = new MockConnection();
		conn.markReady();
		await conn.ready;
		// If we reach here, ready was resolved successfully
		assertEquals(true, true);
	});
});

Deno.test("MockConnection - Error Cases", testOptions, async (t) => {
	await t.step("should return error when accepting from closed stream", async () => {
		const conn = new MockConnection();
		const wt = conn.getWebTransport();
		wt.getBiReader().close();

		const [stream, err] = await conn.acceptStream();
		assertEquals(stream, undefined);
		assertExists(err);
		assertEquals(err?.message, "Failed to accept stream");
	});

	await t.step("should return error when accepting uni from closed stream", async () => {
		const conn = new MockConnection();
		const wt = conn.getWebTransport();
		wt.getUniReader().close();

		const [stream, err] = await conn.acceptUniStream();
		assertEquals(stream, undefined);
		assertExists(err);
		assertEquals(err?.message, "Failed to accept unidirectional stream");
	});
});

Deno.test("MockConnection - Integration Scenarios", testOptions, async (t) => {
	await t.step("should handle multiple stream operations", async () => {
		const conn = new MockConnection();

		// Open multiple streams
		const [stream1] = await conn.openStream();
		const [stream2] = await conn.openUniStream();
		const [stream3] = await conn.openStream();

		// Simulate incoming streams
		conn.simulateIncomingBiStream();
		conn.simulateIncomingUniStream();

		// Accept incoming streams
		const [stream4] = await conn.acceptStream();
		const [stream5] = await conn.acceptUniStream();

		// Verify IDs are correct
		assertEquals(stream1?.id, 0);
		assertEquals(stream2?.id, 2);
		assertEquals(stream3?.id, 4);
		assertEquals(stream4?.id, 1);
		assertEquals(stream5?.id, 3);
	});

	await t.step("should handle ready and close lifecycle", async () => {
		const conn = new MockConnection();

		// Mark as ready
		conn.markReady();
		await conn.ready;

		// Open a stream
		const [stream] = await conn.openStream();
		assertExists(stream);

		// Close connection
		conn.close({ closeCode: 100, reason: "done" });
		const closed = await conn.closed;
		assertEquals(closed.closeCode, 100);
	});
});
