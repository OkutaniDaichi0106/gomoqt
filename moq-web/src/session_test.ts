import { assertEquals, assertExists, assertRejects } from "@std/assert";
import { Session } from "./session.ts";
import { Extensions } from "./extensions.ts";
import { DefaultTrackMux, TrackMux } from "./track_mux.ts";
import type { BroadcastPath } from "./broadcast_path.ts";
import type { TrackPrefix } from "./track_prefix.ts";
import type { TrackName } from "./alias.ts";
import { MockWebTransport } from "./internal/webtransport/mock_connection_test.ts";

Deno.test({
	name: "Session - Constructor",
	sanitizeResources: false,
	sanitizeOps: false,
}, async (t) => {
	await t.step("should create session with default parameters", () => {
		const mockTransport = new MockWebTransport();
		mockTransport.markReady();
		const conn = mockTransport;
		const session = new Session({ conn });
		assertExists(session);
		assertExists(session.ready);
		assertEquals(session.mux, DefaultTrackMux);
		// Catch initialization errors to prevent dangling promises
		session.ready.catch(() => {});
	});

	await t.step("should create session with custom version", () => {
		const mockTransport = new MockWebTransport();
		mockTransport.markReady();
		const conn = mockTransport;
		const versions = new Set([0xffffff00]); // Using valid Version number
		const session = new Session({ conn, versions });
		assertExists(session);
		// Catch initialization errors to prevent dangling promises
		session.ready.catch(() => {});
	});

	await t.step("should create session with custom extensions", () => {
		const mockTransport = new MockWebTransport();
		mockTransport.markReady();
		const conn = mockTransport;
		const extensions = new Extensions();
		const session = new Session({ conn, extensions });
		assertExists(session);
		// Catch initialization errors to prevent dangling promises
		session.ready.catch(() => {});
	});

	await t.step("should create session with custom mux", () => {
		const mockTransport = new MockWebTransport();
		mockTransport.markReady();
		const conn = mockTransport;
		const mux = new TrackMux();
		const session = new Session({ conn, mux });
		assertExists(session);
		assertEquals(session.mux, mux);
		// Catch initialization errors to prevent dangling promises
		session.ready.catch(() => {});
	});
});

Deno.test({
	name: "Session - Ready",
	sanitizeResources: false,
	sanitizeOps: false,
}, async (t) => {
	await t.step("should have ready promise property", () => {
		const mockTransport = new MockWebTransport();
		mockTransport.markReady();
		const conn = mockTransport;
		const session = new Session({ conn });
		assertExists(session.ready);
		assertEquals(typeof session.ready.then, "function");
		// Catch initialization errors to prevent dangling promises
		session.ready.catch(() => {});
	});

	await t.step("should handle setup error on connection failure", async () => {
		const mockTransport = new MockWebTransport();
		mockTransport.failReady(new Error("Connection failed"));
		const conn = mockTransport;
		const session = new Session({ conn });
		await assertRejects(
			() => session.ready,
			Error,
			"Connection failed",
		);
	});
});

Deno.test({
	name: "Session - Methods",
	sanitizeResources: false,
	sanitizeOps: false,
}, async (t) => {
	await t.step("should have all required methods", () => {
		const mockTransport = new MockWebTransport();
		mockTransport.markReady();
		const conn = mockTransport;
		const session = new Session({ conn });
		assertEquals(typeof session.acceptAnnounce, "function");
		assertEquals(typeof session.subscribe, "function");
		assertEquals(typeof session.close, "function");
		assertEquals(typeof session.closeWithError, "function");
		// Catch initialization errors to prevent dangling promises
		session.ready.catch(() => {});
	});

	await t.step("should have acceptAnnounce method", () => {
		const mockTransport = new MockWebTransport();
		mockTransport.markReady();
		const conn = mockTransport;
		const session = new Session({ conn });
		assertExists(session.acceptAnnounce);
		assertEquals(typeof session.acceptAnnounce, "function");
		// Catch initialization errors to prevent dangling promises
		session.ready.catch(() => {});
	});

	await t.step("should have subscribe method", () => {
		const mockTransport = new MockWebTransport();
		mockTransport.markReady();
		const conn = mockTransport;
		const session = new Session({ conn });
		assertExists(session.subscribe);
		assertEquals(typeof session.subscribe, "function");
		// Catch initialization errors to prevent dangling promises
		session.ready.catch(() => {});
	});

	await t.step("should have close method", () => {
		const mockTransport = new MockWebTransport();
		mockTransport.markReady();
		const conn = mockTransport;
		const session = new Session({ conn });
		assertExists(session.close);
		assertEquals(typeof session.close, "function");
		// Catch initialization errors to prevent dangling promises
		session.ready.catch(() => {});
	});

	await t.step("should have closeWithError method", () => {
		const mockTransport = new MockWebTransport();
		mockTransport.markReady();
		const conn = mockTransport;
		const session = new Session({ conn });
		assertExists(session.closeWithError);
		// Catch initialization errors to prevent dangling promises
		session.ready.catch(() => {});
		assertEquals(typeof session.closeWithError, "function");
	});
});

Deno.test({
	name: "Session - Close Operations",
	sanitizeResources: false,
	sanitizeOps: false,
}, async (t) => {
	await t.step("should have close method", () => {
		const mockTransport = new MockWebTransport();
		mockTransport.markReady();
		const conn = mockTransport;
		const session = new Session({ conn });
		// Catch initialization errors to prevent dangling promises
		session.ready.catch(() => {});

		assertExists(session.close);
		assertEquals(typeof session.close, "function");
	});

	await t.step("should have closeWithError method", () => {
		const mockTransport = new MockWebTransport();
		mockTransport.markReady();
		const conn = mockTransport;
		const session = new Session({ conn });
		// Catch initialization errors to prevent dangling promises
		session.ready.catch(() => {});

		assertExists(session.closeWithError);
		assertEquals(typeof session.closeWithError, "function");
	});
});

Deno.test({
	name: "Session - Error Handling",
	sanitizeResources: false,
	sanitizeOps: false,
}, async (t) => {
	await t.step("should handle connection errors", async () => {
		const mockTransport = new MockWebTransport();
		const error = new Error("Connection error");
		mockTransport.failReady(error);
		const conn = mockTransport;
		const session = new Session({ conn });

		await assertRejects(
			() => session.ready,
			Error,
			"Connection error",
		);
	});

	await t.step("should create session with custom versions", () => {
		const mockTransport = new MockWebTransport();
		mockTransport.markReady();
		const conn = mockTransport;
		const versions = new Set([999]); // Custom version number
		const session = new Session({ conn, versions });
		assertExists(session);
		// Catch initialization errors to prevent dangling promises
		session.ready.catch(() => {});
	});
});

Deno.test({
	name: "Session - Async Operations",
	sanitizeResources: false,
	sanitizeOps: false,
}, async (t) => {
	await t.step("acceptAnnounce should be async function", () => {
		const mockTransport = new MockWebTransport();
		mockTransport.markReady();
		const conn = mockTransport;
		const session = new Session({ conn });
		// Catch initialization errors to prevent dangling promises
		session.ready.catch(() => {});

		assertExists(session.acceptAnnounce);
		assertEquals(typeof session.acceptAnnounce, "function");
		// Check it returns a promise
		const result = session.acceptAnnounce("/" as TrackPrefix);
		assertExists(result);
		assertEquals(typeof result.then, "function");
		// Catch any errors to prevent dangling promises
		result.catch(() => {});
	});

	await t.step("subscribe should be async function", () => {
		const mockTransport = new MockWebTransport();
		mockTransport.markReady();
		const conn = mockTransport;
		const session = new Session({ conn });
		// Catch initialization errors to prevent dangling promises
		session.ready.catch(() => {});

		assertExists(session.subscribe);
		assertEquals(typeof session.subscribe, "function");
		// Check it returns a promise
		const result = session.subscribe(
			"/test" as BroadcastPath,
			"video" as TrackName,
		);
		assertExists(result);
		assertEquals(typeof result.then, "function");
		// Catch any errors to prevent dangling promises
		result.catch(() => {});
	});
});
