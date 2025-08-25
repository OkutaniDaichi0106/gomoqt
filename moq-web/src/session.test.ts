import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { Session } from "./session";
import { Version, Versions } from "./internal";
import { AnnouncePleaseMessage, AnnounceInitMessage, GroupMessage,
    SessionClientMessage, SessionServerMessage,
    SubscribeMessage, SubscribeOkMessage } from "./message";
import { Writer, Reader } from "./io";
import { Extensions } from "./internal/extensions";
import { SessionStream } from "./session_stream";
import { background, Context, withPromise } from "./internal/context";
import { AnnouncementReader, AnnouncementWriter } from "./announce_stream";
import { TrackPrefix } from "./track_prefix";
import { ReceiveSubscribeStream, SendSubscribeStream, TrackConfig, SubscribeID } from "./subscribe_stream";
import { BroadcastPath } from "./broadcast_path";
import { TrackReader, TrackWriter } from "./track";
import { GroupReader, GroupWriter } from "./group_stream";
import { DefaultTrackMux, TrackMux } from "./track_mux";
import { BiStreamTypes, UniStreamTypes } from "./stream_type";
import { Queue } from "./internal/queue";
import { Info } from "./info";

// Mock WebTransport
class MockWebTransport {
    ready: Promise<void>;
    closed: Promise<void>;
    incomingBidirectionalStreams: ReadableStream;
    incomingUnidirectionalStreams: ReadableStream;

    constructor() {
        this.ready = Promise.resolve();
        this.closed = new Promise(() => {});
        this.incomingBidirectionalStreams = new ReadableStream({
            start(controller) {
                // Mock implementation
            }
        });
        this.incomingUnidirectionalStreams = new ReadableStream({
            start(controller) {
                // Mock implementation
            }
        });
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

    async createUnidirectionalStream(): Promise<WritableStream> {
        return new WritableStream({
            write(chunk) {
                // Mock implementation
            }
        });
    }

    close(options?: {closeCode?: number, reason?: string}) {
        // Mock implementation
    }
}

// Mock SessionStream
jest.mock("./session_stream", () => ({
    SessionStream: jest.fn().mockImplementation(() => ({
        context: { signal: new AbortController().signal },
        update: jest.fn()
    }))
}));

// Mock messages
jest.mock("./message", () => ({
    SessionClientMessage: {
        encode: jest.fn().mockImplementation(() => Promise.resolve([{ version: 0xffffff00n }, null]))
    },
    SessionServerMessage: {
        decode: jest.fn().mockImplementation(() => Promise.resolve([{ version: 0xffffff00n }, null]))
    },
    AnnouncePleaseMessage: {
        encode: jest.fn().mockImplementation(() => Promise.resolve([{}, null])),
        decode: jest.fn().mockImplementation(() => Promise.resolve([{}, null]))
    },
    AnnounceInitMessage: {
        decode: jest.fn().mockImplementation(() => Promise.resolve([{}, null]))
    },
    GroupMessage: {
        decode: jest.fn().mockImplementation(() => Promise.resolve([{ subscribeId: 1n }, null]))
    },
    SubscribeMessage: {
        encode: jest.fn().mockImplementation(() => Promise.resolve([{ subscribeId: 1n }, null])),
        decode: jest.fn().mockImplementation(() => Promise.resolve([{ subscribeId: 1n, broadcastPath: {}, trackName: "" }, null]))
    },
    SubscribeOkMessage: {
        decode: jest.fn().mockImplementation(() => Promise.resolve([{}, null]))
    },
}));

// Mock IO
jest.mock("./io", () => ({
    Writer: jest.fn().mockImplementation(() => ({
        writeUint8: jest.fn(),
        flush: jest.fn().mockImplementation(() => Promise.resolve(null))
    })),
    Reader: jest.fn().mockImplementation(() => ({
        readUint8: jest.fn().mockImplementation(() => Promise.resolve([1, null]))
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
    TrackConfig: jest.fn(),
    SubscribeID: jest.fn()
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
        serveTrack: jest.fn(),
        serveAnnouncement: jest.fn()
    })),
    DefaultTrackMux: {}
}));

jest.mock("./stream_type", () => ({
    BiStreamTypes: {
        SessionStreamType: 1,
        SubscribeStreamType: 2,
        AnnounceStreamType: 3
    },
    UniStreamTypes: {
        GroupStreamType: 1
    }
}));

jest.mock("./internal/queue", () => ({
    Queue: jest.fn().mockImplementation(() => ({
        enqueue: jest.fn(),
        dequeue: jest.fn()
    }))
}));

jest.mock("./info", () => ({
    Info: jest.fn()
}));

// Mock context and background
jest.mock("./internal/context", () => ({
    background: jest.fn().mockReturnValue({
        signal: new AbortController().signal
    }),
    withPromise: jest.fn().mockImplementation((ctx, promise) => ({
        signal: new AbortController().signal
    })),
    Context: jest.fn()
}));

// Mock Extensions
jest.mock("./internal/extensions", () => ({
    Extensions: jest.fn().mockImplementation(() => ({
        // Mock implementation
    }))
}));

// Mock internal index
jest.mock("./internal", () => ({
    Version: jest.fn(),
    Versions: {
        DEVELOP: 0xffffff00n
    },
    Queue: jest.fn().mockImplementation(() => ({
        enqueue: jest.fn(),
        dequeue: jest.fn()
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

        it("should create a session with custom mux", () => {
            const extensions = new Extensions();
            const mockMux = {
                serveTrack: jest.fn(),
                serveAnnouncement: jest.fn()
            };
            session = new Session(mockConn as any, new Set([Versions.DEVELOP]), extensions, mockMux as any);
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
                ready: Promise.reject(new Error("Connection failed")),
                close: jest.fn() // Add close method to prevent the error
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
            const SessionStreamMock = jest.mocked(SessionStream);
            expect(SessionStreamMock).toHaveBeenCalled();
        });

        it("should send session client message", async () => {
            await session.ready;
            const SessionClientMessageMock = jest.mocked(SessionClientMessage);
            expect(SessionClientMessageMock.encode).toHaveBeenCalled();
        });

        it("should receive session server message", async () => {
            await session.ready;
            const SessionServerMessageMock = jest.mocked(SessionServerMessage);
            expect(SessionServerMessageMock.decode).toHaveBeenCalled();
        });
    });

    describe("internal state", () => {
        beforeEach(() => {
            session = new Session(mockConn as any);
        });

        it("should initialize session successfully", () => {
            expect(session).toBeDefined();
            expect(session.ready).toBeInstanceOf(Promise);
        });

        it("should have all required methods", () => {
            expect(typeof session.update).toBe('function');
            expect(typeof session.close).toBe('function');
            expect(typeof session.closeWithError).toBe('function');
            expect(typeof session.acceptAnnounce).toBe('function');
            expect(typeof session.subscribe).toBe('function');
        });
    });

    describe("methods", () => {
        beforeEach(async () => {
            session = new Session(mockConn as any);
            await session.ready;
        });

        it("should have update method", () => {
            expect(session.update).toBeDefined();
            expect(typeof session.update).toBe('function');
        });

        it("should have acceptAnnounce method", () => {
            expect(session.acceptAnnounce).toBeDefined();
            expect(typeof session.acceptAnnounce).toBe('function');
        });

        it("should have subscribe method", () => {
            expect(session.subscribe).toBeDefined();
            expect(typeof session.subscribe).toBe('function');
        });

        it("should have close method", () => {
            expect(session.close).toBeDefined();
            expect(typeof session.close).toBe('function');
        });

        it("should have closeWithError method", () => {
            expect(session.closeWithError).toBeDefined();
            expect(typeof session.closeWithError).toBe('function');
        });

        it("should call update method without errors", async () => {
            await session.ready;
            expect(() => session.update(1000n)).not.toThrow();
        });

        it("should close connection with normal closure", () => {
            const closeSpy = jest.spyOn(mockConn, 'close');
            session.close();
            expect(closeSpy).toHaveBeenCalledWith({
                closeCode: 0x0,
                reason: "No Error"
            });
        });

        it("should close connection with error", () => {
            const closeSpy = jest.spyOn(mockConn, 'close');
            session.closeWithError(123, "Test error");
            expect(closeSpy).toHaveBeenCalledWith({
                closeCode: 123,
                reason: "Test error"
            });
        });
    });

    describe("error handling", () => {
        beforeEach(async () => {
            session = new Session(mockConn as any);
            await session.ready;
        });

        it("should handle errors gracefully during initialization", async () => {
            const errorMockConn = {
                ...mockConn,
                createBidirectionalStream: jest.fn().mockImplementation(() =>
                    Promise.reject(new Error("Stream creation failed"))
                ),
                close: jest.fn()
            };

            const errorSession = new Session(errorMockConn as any);
            await expect(errorSession.ready).rejects.toThrow();
        });

        it("should handle invalid versions", async () => {
            // Create a separate mock for version testing
            const versionMockConn = new MockWebTransport();
            versionMockConn.close = jest.fn();

            // Mock SessionServerMessage to return incompatible version
            const SessionServerMessageMock = jest.mocked(SessionServerMessage);
            const originalDecode = SessionServerMessageMock.decode;
            SessionServerMessageMock.decode.mockImplementationOnce(() =>
                Promise.resolve([{ version: 999n, extensions: {} } as any, undefined])
            );

            const versionSession = new Session(versionMockConn as any);
            await expect(versionSession.ready).rejects.toThrow("Incompatible session version");

            // Restore original mock
            SessionServerMessageMock.decode = originalDecode;
        });
    });

    describe("async operations", () => {
        let readySession: Session;

        beforeEach(async () => {
            readySession = new Session(mockConn as any);
            await readySession.ready;
        });

        it("should handle acceptAnnounce", async () => {
            const prefix = { segments: ["test"] };
            const stream = readySession.acceptAnnounce(prefix as any);
            expect(stream).toBeInstanceOf(Promise);
        });

        it("should handle subscribe", async () => {
            const path = { segments: ["test", "path"] };
            const trackName = "test-track";
            const config = { trackPriority: 1n, minGroupSequence: 0n, maxGroupSequence: 10n };

            const stream = readySession.subscribe(path as any, trackName, config);
            expect(stream).toBeInstanceOf(Promise);
        });
    });
});
