import { assertEquals, assertInstanceOf } from "@std/assert";
import type { TrackHandler } from "./track_mux.ts";
import { TrackMux } from "./track_mux.ts";
import type { AnnouncementWriter } from "./announce_stream.ts";
import { Announcement } from "./announce_stream.ts";
import type { BroadcastPath } from "./broadcast_path.ts";
import type { TrackPrefix } from "./track_prefix.ts";
import { background, withCancelCause } from "@okudai/golikejs/context";
import type { TrackWriter } from "./track.ts";
import { SubscribeErrorCode } from "./error.ts";

// Mock implementations using DI pattern
class MockTrackHandler implements TrackHandler {
	calls: Array<{ trackWriter: Partial<TrackWriter> }> = [];

	async serveTrack(trackWriter: TrackWriter): Promise<void> {
		this.calls.push({ trackWriter });
	}

	reset() {
		this.calls = [];
	}
}

function createMockTrackWriter(
	broadcastPath: BroadcastPath,
	trackName: string,
): TrackWriter & {
	closeWithErrorCalls: Array<{ code: number }>;
	closeCalls: number;
	reset: () => void;
} {
	const closeWithErrorCalls: Array<{ code: number }> = [];
	let closeCalls = 0;

	return {
		broadcastPath,
		trackName,
		closeWithErrorCalls,
		closeCalls,
		async closeWithError(code: number): Promise<void> {
			closeWithErrorCalls.push({ code });
		},
		async close(): Promise<void> {
			closeCalls++;
		},
		reset() {
			closeWithErrorCalls.length = 0;
			closeCalls = 0;
		},
	} as TrackWriter & {
		closeWithErrorCalls: Array<{ code: number }>;
		closeCalls: number;
		reset: () => void;
	};
}

function createMockAnnouncementWriter(
	context: ReturnType<typeof background>,
): AnnouncementWriter & {
	sendCalls: Announcement[];
	initCalls: Announcement[][];
	closeCalls: number;
	reset: () => void;
} {
	const sendCalls: Announcement[] = [];
	const initCalls: Announcement[][] = [];

	const mock = {
		context,
		sendCalls,
		initCalls,
		closeCalls: 0,
		async send(announcement: Announcement): Promise<Error | undefined> {
			sendCalls.push(announcement);
			return undefined;
		},
		async init(announcements: Announcement[]): Promise<Error | undefined> {
			initCalls.push(announcements);
			return undefined;
		},
		async close(): Promise<void> {
			mock.closeCalls++;
		},
		reset() {
			sendCalls.length = 0;
			initCalls.length = 0;
			mock.closeCalls = 0;
		},
	} as AnnouncementWriter & {
		sendCalls: Announcement[];
		initCalls: Announcement[][];
		closeCalls: number;
		reset: () => void;
	};

	return mock;
}

Deno.test("TrackMux - Constructor", () => {
	const mux = new TrackMux();
	assertInstanceOf(mux, TrackMux);
});

Deno.test("TrackMux - announce", async (t) => {
	await t.step("should register handler for announcement path", async () => {
		const trackMux = new TrackMux();
		const mockHandler = new MockTrackHandler();
		const mockTrackWriter = createMockTrackWriter(
			"/test/path" as BroadcastPath,
			"test-track",
		);

		const [ctx, cancelFunc] = withCancelCause(background());
		const mockAnnouncement = new Announcement("/test/path" as BroadcastPath, ctx.done());

		await trackMux.announce(mockAnnouncement, mockHandler);
		await trackMux.serveTrack(mockTrackWriter);

		assertEquals(mockHandler.calls.length, 1);
		assertEquals(mockHandler.calls[0]?.trackWriter, mockTrackWriter);

		cancelFunc(undefined);
	});

	await t.step("should notify existing announcers when path matches prefix", async () => {
		const trackMux = new TrackMux();
		const mockHandler = new MockTrackHandler();

		const [ctx, cancelFunc] = withCancelCause(background());
		const mockAnnouncement = new Announcement("/test/path" as BroadcastPath, ctx.done());
		const mockAnnouncementWriter = createMockAnnouncementWriter(ctx);

		const prefix = "/test/" as TrackPrefix;

		// First register an announcer
		const servePromise = trackMux.serveAnnouncement(mockAnnouncementWriter, prefix);

		// Then announce a path that matches the prefix
		await trackMux.announce(mockAnnouncement, mockHandler);

		// Cancel the context to complete the serveAnnouncement
		cancelFunc(undefined);
		await servePromise;

		assertEquals(mockAnnouncementWriter.sendCalls.length, 1);
		assertEquals(mockAnnouncementWriter.sendCalls[0], mockAnnouncement);
	});

	await t.step("should clean up handler when announcement ends", async () => {
		const trackMux = new TrackMux();
		const mockHandler = new MockTrackHandler();
		const mockTrackWriter = createMockTrackWriter(
			"/test/path" as BroadcastPath,
			"test-track",
		);

		const [ctx, cancelFunc] = withCancelCause(background());
		const mockAnnouncement = new Announcement("/test/path" as BroadcastPath, ctx.done());

		await trackMux.announce(mockAnnouncement, mockHandler);

		// Initially handler should work
		await trackMux.serveTrack(mockTrackWriter);
		assertEquals(mockHandler.calls.length, 1);

		// Simulate announcement ending
		cancelFunc(undefined);
		await new Promise((resolve) => setTimeout(resolve, 10)); // Wait for async cleanup

		// Handler should be removed
		mockHandler.reset();
		const differentTrackWriter = createMockTrackWriter(
			"/different/path" as BroadcastPath,
			"different-track",
		);

		await trackMux.serveTrack(differentTrackWriter);

		// Should call closeWithError for not found path
		assertEquals(differentTrackWriter.closeWithErrorCalls.length, 1);
		assertEquals(
			differentTrackWriter.closeWithErrorCalls[0]?.code,
			SubscribeErrorCode.TrackNotFound,
		);
	});
});

Deno.test("TrackMux - publish", async () => {
	const trackMux = new TrackMux();
	const mockHandler = new MockTrackHandler();
	const mockTrackWriter = createMockTrackWriter("/test/path" as BroadcastPath, "test-track");

	const [ctx, cancelFunc] = withCancelCause(background());
	const path = "/test/path" as BroadcastPath;

	await trackMux.publish(ctx.done(), path, mockHandler);

	// Test that the handler is registered
	await trackMux.serveTrack(mockTrackWriter);
	assertEquals(mockHandler.calls.length, 1);
	assertEquals(mockHandler.calls[0]?.trackWriter, mockTrackWriter);

	cancelFunc(undefined);
});

Deno.test("TrackMux - serveTrack", async (t) => {
	await t.step("should call registered handler for matching path", async () => {
		const trackMux = new TrackMux();
		const mockHandler = new MockTrackHandler();
		const mockTrackWriter = createMockTrackWriter(
			"/test/path" as BroadcastPath,
			"test-track",
		);

		const [ctx, cancelFunc] = withCancelCause(background());
		const mockAnnouncement = new Announcement("/test/path" as BroadcastPath, ctx.done());

		await trackMux.announce(mockAnnouncement, mockHandler);
		await trackMux.serveTrack(mockTrackWriter);

		assertEquals(mockHandler.calls.length, 1);
		assertEquals(mockHandler.calls[0]?.trackWriter, mockTrackWriter);

		cancelFunc(undefined);
	});

	await t.step("should call NotFoundHandler for unregistered path", async () => {
		const trackMux = new TrackMux();
		const trackWriter = createMockTrackWriter(
			"/different/path" as BroadcastPath,
			"different-track",
		);

		await trackMux.serveTrack(trackWriter);

		// Should call closeWithError for not found path
		assertEquals(trackWriter.closeWithErrorCalls.length, 1);
		assertEquals(trackWriter.closeWithErrorCalls[0]?.code, SubscribeErrorCode.TrackNotFound);
	});
});

Deno.test("TrackMux - serveAnnouncement", async (t) => {
	const prefixCases = {
		"valid prefix": { prefix: "/test/" as TrackPrefix, expectedInit: true },
		"invalid-looking prefix (no validation)": {
			prefix: "invalid-prefix" as TrackPrefix,
			expectedInit: true,
		},
	};

	for (const [name, c] of Object.entries(prefixCases)) {
		await t.step(name, async () => {
			const trackMux = new TrackMux();
			const [ctx, cancelFunc] = withCancelCause(background());
			const mockAnnouncementWriter = createMockAnnouncementWriter(ctx);

			const servePromise = trackMux.serveAnnouncement(
				mockAnnouncementWriter,
				c.prefix,
			);

			// Cancel the context to complete the serveAnnouncement
			cancelFunc(undefined);
			await servePromise;

			if (c.expectedInit) {
				assertEquals(mockAnnouncementWriter.initCalls.length, 1);
				assertEquals(mockAnnouncementWriter.initCalls[0]?.length, 0);
			}
		});
	}

	await t.step("should clean up announcer when context ends", async () => {
		const trackMux = new TrackMux();
		const validPrefix = "/test/" as TrackPrefix;
		const [ctx, cancelFunc] = withCancelCause(background());

		const mockAnnouncementWriter = createMockAnnouncementWriter(ctx);

		const servePromise = trackMux.serveAnnouncement(
			mockAnnouncementWriter,
			validPrefix,
		);

		// Cancel the context to trigger cleanup
		cancelFunc(undefined);
		await servePromise;

		// Subsequent announcements should not be sent to this writer
		const [ctx2, cancelFunc2] = withCancelCause(background());
		const mockAnnouncement = new Announcement("/test/path" as BroadcastPath, ctx2.done());
		const mockHandler = new MockTrackHandler();

		await trackMux.announce(mockAnnouncement, mockHandler);

		// The send should not be called since the writer was cleaned up
		assertEquals(mockAnnouncementWriter.sendCalls.length, 0);

		cancelFunc2(undefined);
	});
});

Deno.test("TrackMux - close", async (t) => {
	await t.step("should close all sessions", async () => {
		const trackMux = new TrackMux();
		const validPrefix = "/test" as TrackPrefix;
		const mockHandler = new MockTrackHandler();

		const [ctx, cancelFunc] = withCancelCause(background());
		const mockAnnouncementWriter = createMockAnnouncementWriter(ctx);

		// Serve an announcement to add the writer to announcers
		const servePromise = trackMux.serveAnnouncement(mockAnnouncementWriter, validPrefix);

		// Announce a track to trigger sending to the writer
		const path = "/test/path" as BroadcastPath;
		const [ctx2, cancelFunc2] = withCancelCause(background());
		await trackMux.announce(new Announcement(path, ctx2.done()), mockHandler);

		// Close the trackMux
		await trackMux.close();

		// Wait for serveAnnouncement to complete
		cancelFunc(undefined);
		await servePromise;

		// Expect the writer's close to be called
		assertEquals(mockAnnouncementWriter.closeCalls, 1);

		cancelFunc2(undefined);
	});

	await t.step("should work with no sessions", async () => {
		const trackMux = new TrackMux();
		await trackMux.close(); // Should not throw
	});
});

Deno.test("TrackHandler - Interface", () => {
	const mockHandler = new MockTrackHandler();

	assertEquals(typeof mockHandler.serveTrack, "function");

	const mockTrackWriter = createMockTrackWriter(
		"/test/path" as BroadcastPath,
		"test-track",
	);
	mockHandler.serveTrack(mockTrackWriter);

	assertEquals(mockHandler.calls.length, 1);
	assertEquals(mockHandler.calls[0]?.trackWriter, mockTrackWriter);
});

Deno.test("TrackMux - Additional Coverage", async (t) => {
	await t.step("should handle inactive announcement", async () => {
		const trackMux = new TrackMux();
		const mockHandler = new MockTrackHandler();

		// Create an announcement with already-resolved context (inactive)
		const mockAnnouncement = new Announcement(
			"/test/path" as BroadcastPath,
			Promise.resolve(),
		);

		// Wait for the announcement to become inactive
		await mockAnnouncement.ended();

		await trackMux.announce(mockAnnouncement, mockHandler);

		// Should not register the handler since announcement is inactive
		const mockTrackWriter = createMockTrackWriter(
			"/test/path" as BroadcastPath,
			"test-track",
		);
		await trackMux.serveTrack(mockTrackWriter);

		// Should call closeWithError since handler was not registered
		assertEquals(mockTrackWriter.closeWithErrorCalls.length, 1);
	});

	await t.step("should replace existing announcement with different instance", async () => {
		const trackMux = new TrackMux();
		const mockHandler1 = new MockTrackHandler();
		const mockHandler2 = new MockTrackHandler();

		const [ctx1, cancelFunc1] = withCancelCause(background());
		const mockAnnouncement1 = new Announcement("/test/path" as BroadcastPath, ctx1.done());

		await trackMux.announce(mockAnnouncement1, mockHandler1);

		// Announce again with different announcement
		const [ctx2, cancelFunc2] = withCancelCause(background());
		const mockAnnouncement2 = new Announcement("/test/path" as BroadcastPath, ctx2.done());

		await trackMux.announce(mockAnnouncement2, mockHandler2);

		// Should use the new handler
		const mockTrackWriter = createMockTrackWriter(
			"/test/path" as BroadcastPath,
			"test-track",
		);
		await trackMux.serveTrack(mockTrackWriter);

		assertEquals(mockHandler1.calls.length, 0);
		assertEquals(mockHandler2.calls.length, 1);

		cancelFunc1(undefined);
		cancelFunc2(undefined);
	});

	await t.step("should handle send error and clean up failed announcer", async () => {
		const trackMux = new TrackMux();
		const mockHandler = new MockTrackHandler();

		const [ctx, cancelFunc] = withCancelCause(background());
		const mockAnnouncementWriter = createMockAnnouncementWriter(ctx);

		// Override send to return an error
		mockAnnouncementWriter.send = async (_: Announcement): Promise<Error | undefined> => {
			return new Error("Send failed");
		};

		const prefix = "/test/" as TrackPrefix;
		const servePromise = trackMux.serveAnnouncement(mockAnnouncementWriter, prefix);

		// Announce a path that matches the prefix
		const [ctx2, cancelFunc2] = withCancelCause(background());
		const mockAnnouncement = new Announcement("/test/path" as BroadcastPath, ctx2.done());
		await trackMux.announce(mockAnnouncement, mockHandler);

		// Cancel the context to complete
		cancelFunc(undefined);
		await servePromise;

		cancelFunc2(undefined);
	});

	await t.step("should use publishFunc to register handler", async () => {
		const trackMux = new TrackMux();
		const mockTrackWriter = createMockTrackWriter(
			"/test/path" as BroadcastPath,
			"test-track",
		);

		const [ctx, cancelFunc] = withCancelCause(background());
		const path = "/test/path" as BroadcastPath;

		let handlerCalled = false;
		await trackMux.publishFunc(ctx.done(), path, async (_trackWriter) => {
			handlerCalled = true;
		});

		await trackMux.serveTrack(mockTrackWriter);

		assertEquals(handlerCalled, true);

		cancelFunc(undefined);
	});

	await t.step("should handle announcement to non-matching prefix", async () => {
		const trackMux = new TrackMux();
		const mockHandler = new MockTrackHandler();

		const [ctx, cancelFunc] = withCancelCause(background());
		const mockAnnouncementWriter = createMockAnnouncementWriter(ctx);

		const prefix = "/other/" as TrackPrefix;
		const servePromise = trackMux.serveAnnouncement(mockAnnouncementWriter, prefix);

		// Announce a path that doesn't match the prefix
		const [ctx2, cancelFunc2] = withCancelCause(background());
		const mockAnnouncement = new Announcement("/test/path" as BroadcastPath, ctx2.done());
		await trackMux.announce(mockAnnouncement, mockHandler);

		// Should not send to the writer
		assertEquals(mockAnnouncementWriter.sendCalls.length, 0);

		cancelFunc(undefined);
		await servePromise;
		cancelFunc2(undefined);
	});
});
