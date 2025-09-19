import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { Client } from "./client";
import { Session } from "./session";
import type { MOQOptions } from "./options";
import { TrackMux } from "./track_mux";

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
                // Mock implementation
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

// Mock Session
jest.mock("./session", () => ({
    Session: jest.fn().mockImplementation((transport: any) => ({
        ready: transport?.ready || Promise.resolve(),
        close: jest.fn(),
    }))
}));

// Mock TrackMux
jest.mock("./track_mux", () => ({
    TrackMux: jest.fn().mockImplementation(() => ({
        // Mock implementation
    })),
    DefaultTrackMux: {}
}));

describe("Client", () => {
    let client: Client;
    
    beforeEach(() => {
        client = new Client();
        jest.clearAllMocks();
    });
    
    describe("constructor", () => {
        it("should create a client with default options", () => {
            const client = new Client();
            expect(client.options).toBeDefined();
            expect(client.options.transport?.allowPooling).toBe(false);
            expect(client.options.transport?.congestionControl).toBe("low-latency");
            expect(client.options.transport?.requireUnreliable).toBe(true);
        });
        
        it("should create a client with custom options", () => {
            const customOptions: MOQOptions = {
                transport: {
                    allowPooling: true,
                    congestionControl: "throughput",
                    requireUnreliable: false,
                }
            };
            const client = new Client(customOptions);
            expect(client.options.transport?.allowPooling).toBe(true);
            expect(client.options.transport?.congestionControl).toBe("throughput");
            expect(client.options.transport?.requireUnreliable).toBe(false);
        });

        it("should create a client with custom mux", () => {
            const mockMux = new TrackMux();
            const client = new Client(undefined, mockMux);
            expect(client).toBeDefined();
        });
    });
    
    describe("dial", () => {
        it("should create a session", async () => {
            const url = "https://example.com";
            
            const session = await client.dial(url);
            expect(session).toBeDefined();
            expect(Session).toHaveBeenCalledWith(expect.any(MockWebTransport), undefined, undefined, expect.anything());
        });
        
        it("should handle URL object", async () => {
            const url = new URL("https://example.com");
            
            const session = await client.dial(url);
            expect(session).toBeDefined();
        });

        it("should handle custom TrackMux", async () => {
            const url = "https://example.com";
            const customMux = new TrackMux();
            
            const session = await client.dial(url, customMux);
            expect(session).toBeDefined();
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
            
            client.close();
            
            expect(session1.close).toHaveBeenCalled();
            expect(session2.close).toHaveBeenCalled();
        });

        it("should work with no sessions", () => {
            expect(() => client.close()).not.toThrow();
        });
    });
    
    describe("abort", () => {
        it("should close all sessions", async () => {
            const session1 = await client.dial("https://example1.com");
            const session2 = await client.dial("https://example2.com");
            
            client.abort();
            
            expect(session1.close).toHaveBeenCalled();
            expect(session2.close).toHaveBeenCalled();
        });

        it("should work with no sessions", () => {
            expect(() => client.abort()).not.toThrow();
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
                expect(session).toBeDefined();
                expect(session.close).toBeDefined();
            });
        });

        it("should handle mixed URL types", async () => {
            const stringUrl = "https://example.com";
            const urlObject = new URL("https://example2.com");

            const session1 = await client.dial(stringUrl);
            const session2 = await client.dial(urlObject);

            expect(session1).toBeDefined();
            expect(session2).toBeDefined();
        });

        it("should handle session management lifecycle", async () => {
            // Create sessions
            const session1 = await client.dial("https://example1.com");
            const session2 = await client.dial("https://example2.com");

            // Close all sessions
            client.close();

            expect(session1.close).toHaveBeenCalled();
            expect(session2.close).toHaveBeenCalled();

            // Should be able to create new sessions after closing
            const session3 = await client.dial("https://example3.com");
            expect(session3).toBeDefined();
        });
    });

    describe("MOQ alias", () => {
        it("should export MOQ as alias for Client", async () => {
            const { MOQ } = await import("./client");
            expect(MOQ).toBe(Client);
        });

        it("should be able to create instance using MOQ", async () => {
            const { MOQ } = await import("./client");
            const moqClient = new MOQ();
            expect(moqClient).toBeInstanceOf(Client);
            expect(moqClient.options).toBeDefined();
        });
    });
});
