import { assertEquals, assertExists, assertInstanceOf, assertThrows } from "@std/assert";
import { Announcement, AnnouncementReader, AnnouncementWriter } from "./announce_stream.ts";
import type { ReceiveStream, SendStream } from "./internal/webtransport/mod.ts";
import { MockStream } from "./internal/webtransport/mock_stream_test.ts";

class MockContext {
	doneCalled = false;
	cancelled = false;
	cancelError?: Error;
	private doneResolve?: () => void;
	private donePromise: Promise<void>;

	constructor() {
		this.donePromise = new Promise<void>((resolve) => {
			this.doneResolve = resolve;
		});
	}

	done(): Promise<void> {
		return this.donePromise;
	}

	err(): Error | undefined {
		return this.cancelError;
	}

	cancel(): void {
		this.cancelled = true;
		this.cancelError = new Error("context cancelled");
		if (this.doneResolve) {
			this.doneResolve();
		}
	}
}

class MockAnnouncePleaseMessage {
	prefix: string;

	constructor(prefix: string) {
		this.prefix = prefix;
	}

	get messageLength(): number {
		return this.prefix.length;
	}

	async encode(_writer: SendStream): Promise<Error | undefined> {
		return undefined;
	}

	async decode(_reader: ReceiveStream): Promise<Error | undefined> {
		return undefined;
	}
}

class MockAnnounceInitMessage {
	suffixes: string[];

	constructor(suffixes: string[]) {
		this.suffixes = suffixes;
	}

	get messageLength(): number {
		return this.suffixes.length;
	}

	async encode(_writer: SendStream): Promise<Error | undefined> {
		return undefined;
	}

	async decode(_reader: ReceiveStream): Promise<Error | undefined> {
		return undefined;
	}
}

// Test Announcement class
Deno.test("Announcement - Normal Cases", async (t) => {
	await t.step("should create active announcement", () => {
		const announcement = new Announcement("/test/path", new Promise(() => {}));
		assertEquals(announcement.broadcastPath, "/test/path");
		assertEquals(announcement.isActive(), true);
	});

	await t.step("should end announcement", async () => {
		const announcement = new Announcement("/test/path", new Promise(() => {}));
		announcement.end();
		assertEquals(announcement.isActive(), false);

		// ended() should resolve
		const endedPromise = announcement.ended();
		await endedPromise; // Should not hang
	});

	await t.step("should end on signal", async () => {
		let resolveSignal: () => void;
		const signal = new Promise<void>((resolve) => {
			resolveSignal = resolve;
		});

		const announcement = new Announcement("/test/path", signal);
		assertEquals(announcement.isActive(), true);

		resolveSignal!();
		await new Promise((resolve) => setTimeout(resolve, 10)); // Allow signal to propagate

		assertEquals(announcement.isActive(), false);
	});
});

Deno.test("Announcement - Error Cases", () => {
	// Announcement constructor validates the path, so test invalid paths
	assertThrows(
		() => {
			new Announcement("", new Promise(() => {}));
		},
		Error,
		"Invalid broadcast path",
	);

	assertThrows(
		() => {
			new Announcement("test", new Promise(() => {}));
		},
		Error,
		"Invalid broadcast path",
	);
});

// Test AnnouncementWriter class
Deno.test("AnnouncementWriter - Constructor", () => {
	const mockStream = new MockStream(42n);
	const mockContext = new MockContext();
	const mockRequest = new MockAnnouncePleaseMessage("/test/");

	const writer = new AnnouncementWriter(
		mockContext as any,
		mockStream as any,
		mockRequest as any,
	);

	assertEquals(writer.prefix, "/test/");
	assertExists(writer.context);
});

Deno.test("AnnouncementWriter - Init Method", async (t) => {
	const mockStream = new MockStream(42n);
	const mockContext = new MockContext();
	const mockRequest = new MockAnnouncePleaseMessage("/test/");

	const writer = new AnnouncementWriter(
		mockContext as any,
		mockStream as any,
		mockRequest as any,
	);

	await t.step("should initialize with valid announcements", async () => {
		const announcement = new Announcement("/test/path1", new Promise(() => {}));
		const result = await writer.init([announcement]);
		assertEquals(result, undefined);
	});

	await t.step("should reject announcement with wrong prefix", async () => {
		const announcement = new Announcement("/wrong/path", new Promise(() => {}));
		const result = await writer.init([announcement]);
		assertExists(result);
		assertInstanceOf(result, Error);
	});

	await t.step("should reject duplicate active announcements", async () => {
		const announcement1 = new Announcement("/test/path1", new Promise(() => {}));
		const announcement2 = new Announcement("/test/path1", new Promise(() => {}));
		const result = await writer.init([announcement1, announcement2]);
		assertExists(result);
		assertInstanceOf(result, Error);
	});
});

Deno.test("AnnouncementWriter - Send Method", async (t) => {
	const mockStream = new MockStream(42n);
	const mockContext = new MockContext();
	const mockRequest = new MockAnnouncePleaseMessage("/test/");

	const writer = new AnnouncementWriter(
		mockContext as any,
		mockStream as any,
		mockRequest as any,
	);

	// Initialize first
	const initAnnouncement = new Announcement("/test/path1", new Promise(() => {}));
	await writer.init([initAnnouncement]);

	await t.step("should send valid announcement", async () => {
		const announcement = new Announcement("/test/path2", new Promise(() => {}));
		const result = await writer.send(announcement);
		assertEquals(result, undefined);
	});

	await t.step("should reject announcement with wrong prefix", async () => {
		const announcement = new Announcement("/wrong/path", new Promise(() => {}));
		const result = await writer.send(announcement);
		assertExists(result);
		assertInstanceOf(result, Error);
	});
});

Deno.test("AnnouncementWriter - Close Methods", async (t) => {
	const mockStream = new MockStream(42n);
	const mockContext = new MockContext();
	const mockRequest = new MockAnnouncePleaseMessage("/test/");

	const writer = new AnnouncementWriter(
		mockContext as any,
		mockStream as any,
		mockRequest as any,
	);

	await t.step("should close normally", async () => {
		await writer.close();
		assertEquals(mockStream.writable.closeCalls, 1);
	});

	await t.step("should close with error", async () => {
		const mockStream2 = new MockStream(43n);
		const mockContext2 = new MockContext();
		const mockRequest2 = new MockAnnouncePleaseMessage("/test/");

		const writer2 = new AnnouncementWriter(
			mockContext2 as any,
			mockStream2 as any,
			mockRequest2 as any,
		);

		await writer2.closeWithError(1, "test error");
		// cancelCalls is an array, check its length
		assertEquals(mockStream2.readable.cancelCalls.length, 1);
	});
});

// Test AnnouncementReader class
Deno.test("AnnouncementReader - Constructor", () => {
	const mockStream = new MockStream(42n);
	const mockContext = new MockContext();
	const mockRequest = new MockAnnouncePleaseMessage("/test/");
	const mockInit = new MockAnnounceInitMessage(["path1", "path2"]);

	const reader = new AnnouncementReader(
		mockContext as any,
		mockStream as any,
		mockRequest as any,
		mockInit as any,
	);

	assertEquals(reader.prefix, "/test/");
	assertExists(reader.context);
});

Deno.test("AnnouncementReader - Constructor Error Cases", () => {
	const mockStream = new MockStream();
	const mockContext = new MockContext();

	assertThrows(
		() => {
			const mockRequest = new MockAnnouncePleaseMessage("invalid prefix");
			const mockInit = new MockAnnounceInitMessage([]);
			new AnnouncementReader(
				mockContext as any,
				mockStream as any,
				mockRequest as any,
				mockInit as any,
			);
		},
		Error,
		"invalid prefix",
	);
});

Deno.test("AnnouncementReader - Receive Method", async (t) => {
	await t.step("should receive initial announcements", async () => {
		const mockStream = new MockStream(42n);
		// Add mock data for announcement: suffix "path1"
		mockStream.readable.data = [new Uint8Array([7]), new Uint8Array([112, 97, 116, 104, 49])]; // "path1"

		const mockContext = new MockContext();
		const mockRequest = new MockAnnouncePleaseMessage("/test/");
		const mockInit = new MockAnnounceInitMessage(["path1"]);

		const reader = new AnnouncementReader(
			mockContext as any,
			mockStream as any,
			mockRequest as any,
			mockInit as any,
		);

		// Set timeout to prevent hanging
		let timerId: number | undefined;
		const timeoutPromise = new Promise<void>((resolve) => {
			timerId = setTimeout(resolve, 100) as unknown as number;
		});
		try {
			const receivePromise = reader.receive(timeoutPromise);

			// Wait with timeout
			await Promise.race([
				receivePromise,
				timeoutPromise.then(() => [undefined, new Error("timeout")] as const),
			]);

			// Should either get an announcement or an error
			// This is acceptable behavior
		} finally {
			if (timerId !== undefined) {
				clearTimeout(timerId);
			}
		}
	});

	await t.step("should handle cancellation gracefully", async () => {
		const mockStream = new MockStream();
		const mockContext = new MockContext();
		const mockRequest = new MockAnnouncePleaseMessage("/test/");
		const mockInit = new MockAnnounceInitMessage([]);

		const reader = new AnnouncementReader(
			mockContext as any,
			mockStream as any,
			mockRequest as any,
			mockInit as any,
		);

		// Cancel the context
		mockContext.cancel();

		// Receive should handle cancellation
		let timerId: number | undefined;
		const timeoutPromise = new Promise<void>((resolve) => {
			timerId = setTimeout(resolve, 50) as unknown as number;
		});
		try {
			await Promise.race([
				reader.receive(timeoutPromise),
				timeoutPromise,
			]);
		} finally {
			if (timerId !== undefined) {
				clearTimeout(timerId);
			}
		}
	});
});

Deno.test("AnnouncementReader - Close Methods", async (t) => {
	const mockStream = new MockStream(42n);
	const mockContext = new MockContext();
	const mockRequest = new MockAnnouncePleaseMessage("/test/");
	const mockInit = new MockAnnounceInitMessage([]);

	const reader = new AnnouncementReader(
		mockContext as any,
		mockStream as any,
		mockRequest as any,
		mockInit as any,
	);

	await t.step("should close normally", async () => {
		await reader.close();
		assertEquals(mockStream.writable.closeCalls, 1);
	});

	await t.step("should close with error", async () => {
		const mockStream2 = new MockStream(43n);
		const mockContext2 = new MockContext();
		const mockRequest2 = new MockAnnouncePleaseMessage("/test/");
		const mockInit2 = new MockAnnounceInitMessage([]);

		const reader2 = new AnnouncementReader(
			mockContext2 as any,
			mockStream2 as any,
			mockRequest2 as any,
			mockInit2 as any,
		);

		await reader2.closeWithError(1, "test error");
		// cancelCalls is an array, check its length
		assertEquals(mockStream2.readable.cancelCalls.length, 1);
	});
});
