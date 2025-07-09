import { Client } from "./client";
import { Session } from "./session";
import { MOQOptions } from "./options";

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
(global as any).WebTransport = MockWebTransport;

// Mock Session
jest.mock("./session", () => ({
    Session: jest.fn().mockImplementation(() => ({
        close: jest.fn(),
    }))
}));

describe("Client", () => {
    let client: Client;
    
    beforeEach(() => {
        client = new Client();
        jest.clearAllMocks();
    });
    
    describe("constructor", () => {
        it("should create a client with default transport options", () => {
            const client = new Client();
            expect(client.transportOptions).toBeDefined();
            expect(client.transportOptions.allowPooling).toBe(false);
            expect(client.transportOptions.congestionControl).toBe("low-latency");
            expect(client.transportOptions.requireUnreliable).toBe(true);
        });
        
        it("should create a client with custom transport options", () => {
            const customOptions: WebTransportOptions = {
                allowPooling: true,
                congestionControl: "throughput",
                requireUnreliable: false,
            };
            const client = new Client(customOptions);
            expect(client.transportOptions).toEqual(customOptions);
        });
    });
    
    describe("dial", () => {
        it("should create a session and add it to sessions", async () => {
            const url = "https://example.com";
            const options: MOQOptions = {};
            
            const session = await client.dial(url, options);
            expect(session).toBeDefined();
            expect(Session).toHaveBeenCalledWith(expect.any(MockWebTransport));
            expect(client.sessions.has(session)).toBe(true);
        });
        
        it("should handle URL object", async () => {
            const url = new URL("https://example.com");
            
            const session = await client.dial(url);
            expect(session).toBeDefined();
            expect(client.sessions.has(session)).toBe(true);
        });
        
        it("should handle WebTransport connection errors", async () => {
            const mockTransport = MockWebTransport as any;
            mockTransport.prototype.ready = Promise.reject(new Error("Connection failed"));
            
            const url = "https://example.com";
            await expect(client.dial(url)).rejects.toThrow("Connection failed");
        });
    });
    
    describe("close", () => {
        it("should close all sessions and clear sessions set", async () => {
            const session1 = await client.dial("https://example1.com");
            const session2 = await client.dial("https://example2.com");
            
            client.close();
            
            expect(session1.close).toHaveBeenCalled();
            expect(session2.close).toHaveBeenCalled();
            expect(client.sessions.size).toBe(0);
        });
    });
    
    describe("abort", () => {
        it("should close all sessions and clear sessions set", async () => {
            const session1 = await client.dial("https://example1.com");
            const session2 = await client.dial("https://example2.com");
            
            client.abort();
            
            expect(session1.close).toHaveBeenCalled();
            expect(session2.close).toHaveBeenCalled();
            expect(client.sessions.size).toBe(0);
        });
    });
});
