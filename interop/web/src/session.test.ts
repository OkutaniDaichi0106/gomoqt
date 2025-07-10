import { Session } from "./session";
import { Version, Versions, Queue } from "./internal";
import { AnnouncePleaseMessage, GroupMessage,
    SessionClientMessage, SessionServerMessage,
    SubscribeMessage, SubscribeOkMessage } from "./message";
import { Writer, Reader } from "./io";
import { Extensions } from "./internal/extensions";
import { SessionStream } from "./session_stream";
import { background, Context, withPromise } from "./internal/context";
import { AnnouncementReader, AnnouncementWriter } from "./announce_stream";
import { TrackPrefix } from "./track_prefix";
import { ReceiveSubscribeStream, SendSubscribeStream, SubscribeConfig, SubscribeID } from "./subscribe_stream";
import { Subscription as Subscription } from "./subscription";
import { BroadcastPath } from "./broadcast_path";
import { TrackReader, TrackWriter } from "./track";
import { GroupReader, GroupWriter } from "./group_stream";
import { TrackMux } from "./track_mux";

// Mock WebTransport
class MockWebTransport {
    ready: Promise<void>;
    closed: Promise<void>;
    
    constructor() {
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

// Mock SessionStream
jest.mock("./session_stream", () => ({
    SessionStream: jest.fn().mockImplementation(() => ({
        // Mock implementation
    }))
}));

// Mock messages
jest.mock("./message", () => ({
    SessionClientMessage: {
        encode: jest.fn().mockResolvedValue([{}, null])
    },
    SessionServerMessage: {
        decode: jest.fn().mockResolvedValue([{}, null])
    },
    AnnouncePleaseMessage: jest.fn(),
    GroupMessage: jest.fn(),
    SubscribeMessage: jest.fn(),
    SubscribeOkMessage: jest.fn(),
}));

// Mock IO
jest.mock("./io", () => ({
    Writer: jest.fn().mockImplementation(() => ({
        // Mock implementation
    })),
    Reader: jest.fn().mockImplementation(() => ({
        // Mock implementation
    }))
}));

// Mock other dependencies
jest.mock("./announce_stream", () => ({
    AnnouncementReader: jest.fn(),
    AnnouncementWriter: jest.fn()
}));

jest.mock("./track_prefix", () => ({
    TrackPrefix: jest.fn()
}));

jest.mock("./subscribe_stream", () => ({
    ReceiveSubscribeStream: jest.fn(),
    SendSubscribeStream: jest.fn(),
    SubscribeConfig: jest.fn(),
    SubscribeID: jest.fn()
}));

jest.mock("./subscription", () => ({
    Subscription: jest.fn()
}));

jest.mock("./broadcast_path", () => ({
    BroadcastPath: jest.fn()
}));

jest.mock("./track", () => ({
    TrackReader: jest.fn(),
    TrackWriter: jest.fn()
}));

jest.mock("./group_stream", () => ({
    GroupReader: jest.fn(),
    GroupWriter: jest.fn()
}));

jest.mock("./track_mux", () => ({
    TrackMux: jest.fn().mockImplementation(() => ({
        // Mock implementation
    }))
}));

describe("Session", () => {
    let mockConn: MockWebTransport;
    let session: Session;
    
    beforeEach(() => {
        mockConn = new MockWebTransport();
        jest.clearAllMocks();
    });
    
    describe("constructor", () => {
        it("should create a session with default parameters", () => {
            session = new Session(mockConn as any);
            expect(session).toBeDefined();
            expect(session.ready).toBeDefined();
        });
        
        it("should create a session with custom versions", () => {
            const versions = new Set([Versions.DEVELOP]);
            session = new Session(mockConn as any, versions);
            expect(session).toBeDefined();
            expect(session.ready).toBeDefined();
        });
        
        it("should create a session with custom extensions", () => {
            const extensions = new Extensions();
            session = new Session(mockConn as any, new Set([Versions.DEVELOP]), extensions);
            expect(session).toBeDefined();
            expect(session.ready).toBeDefined();
        });
    });
    
    describe("ready", () => {
        it("should resolve when connection is ready", async () => {
            session = new Session(mockConn as any);
            await expect(session.ready).resolves.toBeUndefined();
        });
        
        it("should handle connection setup errors", async () => {
            const mockConnWithError = {
                ...mockConn,
                ready: Promise.reject(new Error("Connection failed"))
            };
            
            session = new Session(mockConnWithError as any);
            await expect(session.ready).rejects.toThrow("Connection failed");
        });
    });
    
    describe("session initialization", () => {
        beforeEach(() => {
            session = new Session(mockConn as any);
        });
        
        it("should initialize session stream when ready", async () => {
            await session.ready;
            // Verify that SessionStream was created
            expect(require("./session_stream").SessionStream).toHaveBeenCalled();
        });
        
        it("should send session client message", async () => {
            await session.ready;
            expect(require("./message").SessionClientMessage.encode).toHaveBeenCalled();
        });
        
        it("should receive session server message", async () => {
            await session.ready;
            expect(require("./message").SessionServerMessage.decode).toHaveBeenCalled();
        });
    });
    
    describe("internal state", () => {
        beforeEach(() => {
            session = new Session(mockConn as any);
        });
        
        it("should initialize with empty subscribings map", () => {
            expect((session as any).subscribings).toBeInstanceOf(Map);
            expect((session as any).subscribings.size).toBe(0);
        });
        
        it("should initialize with TrackMux", () => {
            expect(require("./track_mux").TrackMux).toHaveBeenCalled();
        });
        
        it("should initialize with id counter at 0", () => {
            expect((session as any).idCounter).toBe(0n);
        });
    });
});
