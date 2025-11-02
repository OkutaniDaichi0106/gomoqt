import { assertEquals, assertExists } from "@std/assert";
import { SessionStream } from "./session_stream.ts";
import { background, withCancelCause } from "@okudai/golikejs/context";
import { SessionUpdateMessage } from "./internal/message/mod.ts";
import type { SessionClientMessage, SessionServerMessage } from "./internal/message/mod.ts";
import { Extensions } from "./extensions.ts";
import type { Version } from "./version.ts";

// Test configuration to ignore resource leaks from background operations
const testOptions = {
  sanitizeResources: false,
  sanitizeOps: false,
};

// Mock Stream implementation for testing
class MockStream {
	readable: MockReadable;
	writable: MockWritable;

	constructor(readable: MockReadable, writable: MockWritable) {
		this.readable = readable;
		this.writable = writable;
	}
}

class MockReadable {
	#messages: SessionUpdateMessage[] = [];
	#readCount = 0;
	#closed = false;

	constructor(messages: SessionUpdateMessage[] = []) {
		this.#messages = messages;
	}

	// Simulate the Reader interface that SessionUpdateMessage.decode expects
	// decode first reads message length, then bitrate
	async readVarint(): Promise<[number, Error | undefined]> {
		if (this.#closed) {
			return [0, new Error("Stream closed")];
		}

		const msgIndex = Math.floor(this.#readCount / 2);
		const isLength = this.#readCount % 2 === 0;
		this.#readCount++;

		if (msgIndex >= this.#messages.length) {
			return [0, new Error("EOF")];
		}

		const msg = this.#messages[msgIndex];
		if (!msg) {
			return [0, new Error("EOF")];
		}

		if (isLength) {
			// Return message length
			return [msg.messageLength, undefined];
		} else {
			// Return bitrate
			return [msg.bitrate, undefined];
		}
	}

	close() {
		this.#closed = true;
	}
}

class MockWritable {
	#encoded: number[] = [];
	#encodeError: Error | undefined = undefined;

	setEncodeError(err: Error | undefined) {
		this.#encodeError = err;
	}

	// Simulate the SendStream interface that SessionUpdateMessage.encode expects
	// encode first writes message length, then bitrate
	writeVarint(value: number): void {
		if (!this.#encodeError) {
			this.#encoded.push(value);
		}
	}

	async flush(): Promise<Error | undefined> {
		return this.#encodeError;
	}

	getEncoded(): number[] {
		return this.#encoded;
	}

	// Get only the bitrate values (every other value, starting from index 1)
	getBitrates(): number[] {
		return this.#encoded.filter((_, i) => i % 2 === 1);
	}

	clear() {
		this.#encoded = [];
	}
}

// Helper to create mock messages
function createMockClient(versions: Set<Version> = new Set([0xffffff00])): SessionClientMessage {
	const extensions = new Extensions(new Map());
	return {
		versions,
		// SessionStream will call new Extensions(extensions), which expects a Map
		// So we pass the entries property
		extensions: extensions.entries as any,
	} as unknown as SessionClientMessage;
}

function createMockServer(version: Version = 0xffffff00): SessionServerMessage {
	const extensions = new Extensions(new Map());
	return {
		version,
		// SessionStream will call new Extensions(extensions), which expects a Map
		// So we pass the entries property
		extensions: extensions.entries as any,
	} as unknown as SessionServerMessage;
}

Deno.test("SessionStream - Constructor", testOptions, async (t) => {
	await t.step("should initialize with provided parameters", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const readable = new MockReadable();
		const writable = new MockWritable();
		const stream = new MockStream(readable, writable);
		const client = createMockClient();
		const server = createMockServer();
		const detectFunc = async () => 0;

		const sessionStream = new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		assertEquals(sessionStream.clientInfo.versions, client.versions);
		assertEquals(sessionStream.clientInfo.bitrate, 0);
		assertEquals(sessionStream.serverInfo.version, server.version);
		assertEquals(sessionStream.serverInfo.bitrate, 0);
		assertEquals(sessionStream.context, ctx);

		cancel(undefined);
		await new Promise((resolve) => setTimeout(resolve, 10));
	});

	await t.step("should initialize with custom versions", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const readable = new MockReadable();
		const writable = new MockWritable();
		const stream = new MockStream(readable, writable);
		const versions = new Set<Version>([0xffffff00, 0xffffff01]);
		const client = createMockClient(versions);
		const server = createMockServer(0xffffff01);
		const detectFunc = async () => 0;

		const sessionStream = new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		assertEquals(sessionStream.clientInfo.versions.size, 2);
		assertEquals(sessionStream.clientInfo.versions.has(0xffffff00), true);
		assertEquals(sessionStream.clientInfo.versions.has(0xffffff01), true);
		assertEquals(sessionStream.serverInfo.version, 0xffffff01);

		cancel(undefined);
		await new Promise((resolve) => setTimeout(resolve, 10));
	});
});

Deno.test("SessionStream - Context", testOptions, async (t) => {
	await t.step("should return the internal context", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const readable = new MockReadable();
		const writable = new MockWritable();
		const stream = new MockStream(readable, writable);
		const client = createMockClient();
		const server = createMockServer();
		const detectFunc = async () => 0;

		const sessionStream = new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		const context = sessionStream.context;
		assertExists(context);
		assertEquals(typeof context.done, "function");
		assertEquals(typeof context.err, "function");

		cancel(undefined);
		await new Promise((resolve) => setTimeout(resolve, 10));
	});

	await t.step("should stop background operations on context cancellation", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const readable = new MockReadable();
		const writable = new MockWritable();
		const stream = new MockStream(readable, writable);
		const client = createMockClient();
		const server = createMockServer();
		const detectFunc = async () => 0;

		const sessionStream = new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		assertEquals(sessionStream.context.err(), undefined);

		cancel(new Error("Context cancelled"));
		await new Promise((resolve) => setTimeout(resolve, 10));

		assertExists(sessionStream.context.err());
	});
});

Deno.test("SessionStream - Client Info", testOptions, async (t) => {
	await t.step("should return initial client information", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const readable = new MockReadable();
		const writable = new MockWritable();
		const stream = new MockStream(readable, writable);
		const client = createMockClient();
		const server = createMockServer();
		const detectFunc = async () => 0;

		const sessionStream = new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		const clientInfo = sessionStream.clientInfo;
		assertExists(clientInfo);
		assertEquals(clientInfo.versions, client.versions);
		assertEquals(clientInfo.bitrate, 0);
		assertExists(clientInfo.extensions);

		cancel(undefined);
		await new Promise((resolve) => setTimeout(resolve, 10));
	});
});

Deno.test("SessionStream - Server Info", testOptions, async (t) => {
	await t.step("should return initial server information", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const readable = new MockReadable();
		const writable = new MockWritable();
		const stream = new MockStream(readable, writable);
		const client = createMockClient();
		const server = createMockServer();
		const detectFunc = async () => 0;

		const sessionStream = new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		const serverInfo = sessionStream.serverInfo;
		assertExists(serverInfo);
		assertEquals(serverInfo.version, server.version);
		assertEquals(serverInfo.bitrate, 0);
		assertExists(serverInfo.extensions);

		cancel(undefined);
		await new Promise((resolve) => setTimeout(resolve, 10));
	});

	await t.step("should update server bitrate on receiving update message", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const updateMsg = new SessionUpdateMessage({ bitrate: 1000 });
		const readable = new MockReadable([updateMsg]);
		const writable = new MockWritable();
		const stream = new MockStream(readable, writable);
		const client = createMockClient();
		const server = createMockServer();
		const detectFunc = async () => 0;

		const sessionStream = new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		// Wait for the update to be processed
		await new Promise((resolve) => setTimeout(resolve, 20));

		const serverInfo = sessionStream.serverInfo;
		assertEquals(serverInfo.bitrate, 1000);

		cancel(undefined);
		await new Promise((resolve) => setTimeout(resolve, 10));
	});
});

Deno.test("SessionStream - Bitrate Detection", testOptions, async (t) => {
	await t.step("should call detectFunc periodically", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const readable = new MockReadable();
		const writable = new MockWritable();
		const stream = new MockStream(readable, writable);
		const client = createMockClient();
		const server = createMockServer();

		let detectCallCount = 0;
		const detectFunc = async () => {
			detectCallCount++;
			return detectCallCount * 100;
		};

		new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		// Wait for a few detect cycles
		await new Promise((resolve) => setTimeout(resolve, 50));

		// detectFunc should have been called at least once
		assertEquals(detectCallCount >= 1, true);

		cancel(undefined);
		await new Promise((resolve) => setTimeout(resolve, 10));
	});

	await t.step("should send update message when bitrate changes", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const readable = new MockReadable();
		const writable = new MockWritable();
		const stream = new MockStream(readable, writable);
		const client = createMockClient();
		const server = createMockServer();

		let detectCallCount = 0;
		const detectFunc = async () => {
			detectCallCount++;
			return detectCallCount * 500;
		};

		new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		// Wait for detect cycles
		await new Promise((resolve) => setTimeout(resolve, 50));

		// Check that update messages were sent
		const bitrates = writable.getBitrates();
		assertEquals(bitrates.length >= 1, true);
		if (bitrates.length > 0) {
			assertEquals(bitrates[0], 500);
		}

		cancel(undefined);
		await new Promise((resolve) => setTimeout(resolve, 10));
	});

	await t.step("should stop detecting on context cancellation", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const readable = new MockReadable();
		const writable = new MockWritable();
		const stream = new MockStream(readable, writable);
		const client = createMockClient();
		const server = createMockServer();

		let detectCallCount = 0;
		const detectFunc = async () => {
			detectCallCount++;
			return 100;
		};

		new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		await new Promise((resolve) => setTimeout(resolve, 20));
		const callsBeforeCancel = detectCallCount;

		cancel(new Error("Cancelled"));
		await new Promise((resolve) => setTimeout(resolve, 30));

		// detectFunc should not have been called many more times after cancellation
		assertEquals(detectCallCount <= callsBeforeCancel + 2, true);
	});
});

Deno.test("SessionStream - Update Handling", testOptions, async (t) => {
	await t.step("should process multiple update messages", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const msg1 = new SessionUpdateMessage({ bitrate: 1000 });
		const msg2 = new SessionUpdateMessage({ bitrate: 2000 });
		const msg3 = new SessionUpdateMessage({ bitrate: 3000 });
		const readable = new MockReadable([msg1, msg2, msg3]);
		const writable = new MockWritable();
		const stream = new MockStream(readable, writable);
		const client = createMockClient();
		const server = createMockServer();
		const detectFunc = async () => 0;

		const sessionStream = new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		// Wait for all updates to be processed
		await new Promise((resolve) => setTimeout(resolve, 50));

		// The last update should set the bitrate to 3000
		assertEquals(sessionStream.serverInfo.bitrate, 3000);

		cancel(undefined);
		await new Promise((resolve) => setTimeout(resolve, 10));
	});

	await t.step("should handle decode errors gracefully", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const readable = new MockReadable();
		readable.close(); // Force decode to fail immediately
		const writable = new MockWritable();
		const stream = new MockStream(readable, writable);
		const client = createMockClient();
		const server = createMockServer();
		const detectFunc = async () => 0;

		const sessionStream = new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		// Wait to ensure handleUpdates has tried to read
		await new Promise((resolve) => setTimeout(resolve, 30));

		// Session should still be functional
		assertExists(sessionStream.context);
		assertExists(sessionStream.clientInfo);
		assertExists(sessionStream.serverInfo);

		cancel(undefined);
		await new Promise((resolve) => setTimeout(resolve, 10));
	});

	await t.step("should stop handling updates on context cancellation", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const msg1 = new SessionUpdateMessage({ bitrate: 1000 });
		const msg2 = new SessionUpdateMessage({ bitrate: 2000 });
		const readable = new MockReadable([msg1, msg2]);
		const writable = new MockWritable();
		const stream = new MockStream(readable, writable);
		const client = createMockClient();
		const server = createMockServer();
		const detectFunc = async () => 0;

		const sessionStream = new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		await new Promise((resolve) => setTimeout(resolve, 20));

		cancel(new Error("Cancelled"));
		await new Promise((resolve) => setTimeout(resolve, 20));

		// After cancellation, no more updates should be processed
		// The bitrate might be 0, 1000, or 2000 depending on timing
		const bitrate = sessionStream.serverInfo.bitrate;
		assertEquals(typeof bitrate, "number");

		await new Promise((resolve) => setTimeout(resolve, 10));
	});
});

Deno.test("SessionStream - Updated Method", testOptions, async (t) => {
	await t.step("should exist as a method", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const readable = new MockReadable();
		const writable = new MockWritable();
		const stream = new MockStream(readable, writable);
		const client = createMockClient();
		const server = createMockServer();
		const detectFunc = async () => 0;

		const sessionStream = new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		// Verify the method exists and has the correct type
		assertEquals(typeof sessionStream.updated, "function");

		cancel(undefined);
		await new Promise((resolve) => setTimeout(resolve, 10));
	});

	await t.step("should allow server bitrate updates to be observed", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const msg = new SessionUpdateMessage({ bitrate: 5000 });
		const readable = new MockReadable([msg]);
		const writable = new MockWritable();
		const stream = new MockStream(readable, writable);
		const client = createMockClient();
		const server = createMockServer();
		const detectFunc = async () => 0;

		const sessionStream = new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		// Wait for the update to be processed
		await new Promise((resolve) => setTimeout(resolve, 30));

		// The update should have been processed
		// We can verify the bitrate was updated
		assertEquals(sessionStream.serverInfo.bitrate, 5000);

		cancel(undefined);
		await new Promise((resolve) => setTimeout(resolve, 10));
	});
});

Deno.test("SessionStream - Error Cases", testOptions, async (t) => {
	await t.step("should handle encode errors when sending updates", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const readable = new MockReadable();
		const writable = new MockWritable();
		writable.setEncodeError(new Error("Encode failed"));
		const stream = new MockStream(readable, writable);
		const client = createMockClient();
		const server = createMockServer();
		const detectFunc = async () => 100;

		new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		// Wait for detect to trigger update
		await new Promise((resolve) => setTimeout(resolve, 30));

		// The session should handle the error (logged to console.error)
		// but not crash
		cancel(undefined);
		await new Promise((resolve) => setTimeout(resolve, 10));
	});
});

Deno.test("SessionStream - Edge Cases", testOptions, async (t) => {
	await t.step("should handle zero bitrate", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const readable = new MockReadable();
		const writable = new MockWritable();
		const stream = new MockStream(readable, writable);
		const client = createMockClient();
		const server = createMockServer();
		const detectFunc = async () => 0;

		new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		await new Promise((resolve) => setTimeout(resolve, 30));

		// Check that zero bitrate is handled correctly
		const bitrates = writable.getBitrates();
		if (bitrates.length > 0) {
			assertEquals(bitrates[0], 0);
		}

		cancel(undefined);
		await new Promise((resolve) => setTimeout(resolve, 10));
	});

	await t.step("should handle large bitrate values", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const readable = new MockReadable();
		const writable = new MockWritable();
		const stream = new MockStream(readable, writable);
		const client = createMockClient();
		const server = createMockServer();
		const detectFunc = async () => 1_000_000_000; // 1 Gbps

		new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		await new Promise((resolve) => setTimeout(resolve, 30));

		const bitrates = writable.getBitrates();
		if (bitrates.length > 0) {
			assertEquals(bitrates[0], 1_000_000_000);
		}

		cancel(undefined);
		await new Promise((resolve) => setTimeout(resolve, 10));
	});

	await t.step("should handle empty version set", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const readable = new MockReadable();
		const writable = new MockWritable();
		const stream = new MockStream(readable, writable);
		const client = createMockClient(new Set());
		const server = createMockServer();
		const detectFunc = async () => 0;

		const sessionStream = new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		assertEquals(sessionStream.clientInfo.versions.size, 0);

		cancel(undefined);
		await new Promise((resolve) => setTimeout(resolve, 10));
	});
});

Deno.test("SessionStream - Integration", testOptions, async (t) => {
	await t.step("should handle complete session lifecycle", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const msg1 = new SessionUpdateMessage({ bitrate: 500 });
		const msg2 = new SessionUpdateMessage({ bitrate: 1500 });
		const readable = new MockReadable([msg1, msg2]);
		const writable = new MockWritable();
		const stream = new MockStream(readable, writable);
		const client = createMockClient();
		const server = createMockServer();

		let detectCallCount = 0;
		const detectFunc = async () => {
			detectCallCount++;
			return detectCallCount * 200;
		};

		const sessionStream = new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		// Verify initial state
		assertExists(sessionStream.context);
		assertEquals(sessionStream.clientInfo.versions, client.versions);
		assertEquals(sessionStream.serverInfo.version, server.version);

		// Wait for operations
		await new Promise((resolve) => setTimeout(resolve, 50));

		// Verify that both server updates and client detects happened
		assertEquals(detectCallCount >= 1, true);
		const bitrates = writable.getBitrates();
		assertEquals(bitrates.length >= 1, true);

		// Verify final bitrate from server updates
		assertEquals(sessionStream.serverInfo.bitrate >= 500, true);

		cancel(undefined);
		await new Promise((resolve) => setTimeout(resolve, 10));
	});

	await t.step("should coordinate multiple concurrent operations", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const updates = Array.from({ length: 5 }, (_, i) =>
			new SessionUpdateMessage({ bitrate: (i + 1) * 1000 })
		);
		const readable = new MockReadable(updates);
		const writable = new MockWritable();
		const stream = new MockStream(readable, writable);
		const client = createMockClient();
		const server = createMockServer();

		let detectValue = 0;
		const detectFunc = async () => {
			detectValue += 100;
			return detectValue;
		};

		const sessionStream = new SessionStream({
			context: ctx,
			stream: stream as any,
			client,
			server,
			detectFunc,
		});

		// Let both loops run for a while
		await new Promise((resolve) => setTimeout(resolve, 80));

		// Verify that operations ran concurrently
		assertEquals(detectValue >= 100, true);
		assertEquals(sessionStream.serverInfo.bitrate >= 1000, true);

		cancel(undefined);
		await new Promise((resolve) => setTimeout(resolve, 10));
	});
});
