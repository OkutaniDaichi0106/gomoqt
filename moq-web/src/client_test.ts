import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows } from "../deps.ts";
// We'll mock Session so tests don't depend on real WebTransport streams
import { TrackMux, DefaultTrackMux } from "./track_mux.ts";

// TODO: Migrate mock to Deno compatible pattern
import { Client } from "./client.ts";
import type { MOQOptions } from "./options.ts";

// Mock WebTransport
class MockWebTransport {
    ready: Promise<void>;
    closed: Promise<void>;
    
    constructor(url: string | URL, options?: WebTransportOptions) {
        this.ready = Promise.resolve();
        this.closed = new Promise(() => {});
    }
    
    async createBidirectionalStream(): Promise<{writable: WritableStream, readable: ReadableStream}> {
        const writable = new WritableStream({
            write(chunk) {
                // Mock implementation
            }
        });
        const readable = new ReadableStream({
            start(controller) {
                // Enqueue mock data for SessionServerMessage: versions=0, extensions=0
                controller.enqueue(new Uint8Array([0x00, 0x00]));
                controller.close();
            }
        });
        return { writable, readable };
    }
    
    close() {
        // Mock implementation
    }
}

// Mock global WebTransport
(globalThis as any).WebTransport = MockWebTransport;

describe("Client", () => {
    let client: Client;
    
    beforeEach(() => {
        client = new Client();
        vi.clearAllMocks();
    });
    
    describe("constructor", () => {
        it("should create a client with default options", () => {
            const client = new Client();
            assertExists(client.options);
            assertExists(client.options.versions);
            assertInstanceOf(client.options.versions, Set);
            assertEquals(client.options.transportOptions?.allowPooling, false);
            assertEquals(client.options.transportOptions?.congestionControl, "low-latency");
            assertEquals(client.options.transportOptions?.requireUnreliable, true);
        });
        
        it("should create a client with custom options", () => {
            const customOptions: MOQOptions = {
                versions: new Set([1n]),
                transportOptions: {
                    allowPooling: true,
                    congestionControl: "throughput",
                    requireUnreliable: false,
                }
            };
            const client = new Client(customOptions);
            assertEquals(client.options.versions, new Set([1n]));
            assertEquals(client.options.transportOptions?.allowPooling, true);
            assertEquals(client.options.transportOptions?.congestionControl, "throughput");
            assertEquals(client.options.transportOptions?.requireUnreliable, false);
        });

        it("should create a client with custom mux", () => {
            const mockMux = new TrackMux();
            const client = new Client(undefined, mockMux);
            assertExists(client);
        });
    });
    
    describe("dial", () => {
        it("should create a session", async () => {
            const url = "https://example.com";
            
            const session = await client.dial(url);
            assertExists(session);
            assertEquals(session.mux, DefaultTrackMux);
        });
        
        it("should handle URL object", async () => {
            const url = new URL("https://example.com");
            
            const session = await client.dial(url);
            assertExists(session);
        });

        it("should handle custom TrackMux", async () => {
            const url = "https://example.com";
            const customMux = new TrackMux();
            
            const session = await client.dial(url, customMux);
            assertExists(session);
            assertEquals(session.mux, customMux);
        });
        
        it("should handle WebTransport connection errors", async () => {
            // Create a new mock for this specific test to avoid affecting other tests
            const FailingMockWebTransport = class extends MockWebTransport {
                constructor(url: string | URL, options?: WebTransportOptions) {
                    super(url, options);
                    this.ready = Promise.reject(new Error("Connection failed"));
                }
            };
            
            // Temporarily replace the global WebTransport
            const originalWebTransport = (globalThis as any).WebTransport;
            (globalThis as any).WebTransport = FailingMockWebTransport;
            
            try {
                const url = "https://example.com";
                await expect(client.dial(url)).rejects.toThrow("Connection failed");
            } finally {
                // Restore the original WebTransport
                (globalThis as any).WebTransport = originalWebTransport;
            }
        });
    });
    
    describe("close", () => {
        it("should close all sessions", async () => {
            const session1 = await client.dial("https://example1.com");
            const session2 = await client.dial("https://example2.com");
            
            vi.spyOn(session1, 'close').mockResolvedValue(undefined);
            vi.spyOn(session2, 'close').mockResolvedValue(undefined);
            
            await client.close();
            
            expect(session1.close).toHaveBeenCalled();
            expect(session2.close).toHaveBeenCalled();
        });

        it("should work with no sessions", async () => {
            await expect(client.close()).resolves.toBeUndefined();
        });
    });
    
    describe("abort", () => {
        it("should close all sessions", async () => {
            const session1 = await client.dial("https://example1.com");
            const session2 = await client.dial("https://example2.com");
            
            vi.spyOn(session1, 'close').mockResolvedValue(undefined);
            vi.spyOn(session2, 'close').mockResolvedValue(undefined);
            
            await client.abort();
            
            expect(session1.close).toHaveBeenCalled();
            expect(session2.close).toHaveBeenCalled();
        });

        it("should work with no sessions", async () => {
            await expect(client.abort()).resolves.toBeUndefined();
        });
    });

    describe("integration tests", () => {
        it("should handle multiple dial calls", async () => {
            const urls = [
                "https://example1.com",
                "https://example2.com", 
                "https://example3.com"
            ];

            const sessions = await Promise.all(
                urls.map(url => client.dial(url))
            );

            expect(sessions).toHaveLength(3);
            sessions.forEach(session => {
                assertExists(session);
                assertExists(session.close);
            });
        });

        it("should handle mixed URL types", async () => {
            const stringUrl = "https://example.com";
            const urlObject = new URL("https://example2.com");

            const session1 = await client.dial(stringUrl);
            const session2 = await client.dial(urlObject);

                assertExists(session1);
                assertExists(session2);
        });

        it("should handle session management lifecycle", async () => {
            // Create sessions
            const session1 = await client.dial("https://example1.com");
            const session2 = await client.dial("https://example2.com");

            vi.spyOn(session1, 'close').mockResolvedValue(undefined);
            vi.spyOn(session2, 'close').mockResolvedValue(undefined);

            // Close all sessions
            await client.close();

            expect(session1.close).toHaveBeenCalled();
            expect(session2.close).toHaveBeenCalled();

            // Should be able to create new sessions after closing
            const session3 = await client.dial("https://example3.com");
            assertExists(session3);
        });
    });

    describe("MOQ alias", () => {
        it("should export MOQ as alias for Client", async () => {
            const { MOQ } = await import("./client");
            assertEquals(MOQ, Client);
        });

        it("should be able to create instance using MOQ", async () => {
            const { MOQ } = await import("./client");
            const moqClient = new MOQ();
            assertInstanceOf(moqClient, Client);
            assertExists(moqClient.options);
        });
    });
});
