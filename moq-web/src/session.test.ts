import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import type { MockedFunction } from 'jest-mock';
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
import { ReceiveSubscribeStream, SendSubscribeStream, TrackConfig } from "./subscribe_stream";
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
        context: { 
            done: jest.fn().mockReturnValue(Promise.resolve()),
            err: jest.fn().mockReturnValue(undefined)
        },
        update: jest.fn()
    }))
}));

// Mock messages
jest.mock("./message", () => ({
    SessionClientMessage: jest.fn().mockImplementation((init: any = {}) => ({
        versions: init.versions ?? new Set([0xffffff00n]),
        extensions: init.extensions ?? {},
        encode: jest.fn().mockImplementation(() => Promise.resolve(undefined)),
        decode: jest.fn().mockImplementation(() => Promise.resolve(undefined))
    })),
    SessionServerMessage: jest.fn().mockImplementation((init: any = {}) => ({
        version: init.version ?? 0xffffff00n,
        extensions: init.extensions ?? {},
        encode: jest.fn().mockImplementation(() => Promise.resolve(undefined)),
        decode: jest.fn().mockImplementation(function(this: any) {
            this.version = this.version ?? 0xffffff00n;
            return Promise.resolve(undefined);
        })
    })),
    AnnouncePleaseMessage: jest.fn().mockImplementation((init: any = {}) => ({
        prefix: init.prefix ?? "",
        encode: jest.fn().mockImplementation(() => Promise.resolve(undefined)),
        decode: jest.fn().mockImplementation(() => Promise.resolve(undefined))
    })),
    AnnounceInitMessage: jest.fn().mockImplementation((init: any = {}) => ({
        suffixes: init.suffixes ?? [],
        encode: jest.fn().mockImplementation(() => Promise.resolve(undefined)),
        decode: jest.fn().mockImplementation(() => Promise.resolve(undefined))
    })),
    GroupMessage: jest.fn().mockImplementation((init: any = {}) => ({
        subscribeId: init.subscribeId ?? 1n,
        sequence: init.sequence ?? 0n,
        encode: jest.fn().mockImplementation(() => Promise.resolve(undefined)),
        decode: jest.fn().mockImplementation(() => Promise.resolve(undefined))
    })),
    SubscribeMessage: jest.fn().mockImplementation((init: any = {}) => ({
        subscribeId: init.subscribeId ?? 1n,
        broadcastPath: init.broadcastPath ?? {},
        trackName: init.trackName ?? "",
        encode: jest.fn().mockImplementation(() => Promise.resolve(undefined)),
        decode: jest.fn().mockImplementation(() => Promise.resolve(undefined))
    })),
    SubscribeOkMessage: jest.fn().mockImplementation((init: any = {}) => ({
        groupPeriod: init.groupPeriod ?? 0,
        encode: jest.fn().mockImplementation(() => Promise.resolve(undefined)),
        decode: jest.fn().mockImplementation(() => Promise.resolve(undefined))
    }))
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

// Mock protocol module
jest.mock("./protocol", () => ({
    TrackName: jest.fn(),
    SubscribeID: jest.fn()
}));

// Mock context and background
jest.mock("./internal/context", () => ({
    background: jest.fn().mockReturnValue({
        done: jest.fn().mockReturnValue(Promise.resolve()),
        err: jest.fn().mockReturnValue(undefined)
    }),
    withPromise: jest.fn().mockImplementation((ctx, promise) => ({
        done: jest.fn().mockReturnValue(Promise.resolve()),
        err: jest.fn().mockReturnValue(undefined)
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
    DEFAULT_CLIENT_VERSIONS: new Set([0xffffff00n]),
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

    afterEach(() => {
        // No cleanup needed since console.error is handled globally
    });

    describe("constructor", () => {
        it("should create a session with default parameters", () => {
            session = new Session({conn: mockConn as any});
            expect(session).toBeDefined();
            expect(session.ready).toBeDefined();
        });

        it("should create a session with custom versions", () => {
            const versions = new Set([Versions.DEVELOP]);
            session = new Session({conn: mockConn as any, versions});
            expect(session).toBeDefined();
            expect(session.ready).toBeDefined();
        });

        it("should create a session with custom extensions", () => {
            const extensions = new Extensions();
            session = new Session({conn: mockConn as any, versions: new Set([Versions.DEVELOP]), extensions});
            expect(session).toBeDefined();
            expect(session.ready).toBeDefined();
        });

        it("should create a session with custom mux", () => {
            const extensions = new Extensions();
            const mockMux = {
                serveTrack: jest.fn(),
                serveAnnouncement: jest.fn()
            };
            session = new Session({conn: mockConn as any, versions: new Set([Versions.DEVELOP]), extensions, mux: mockMux as any});
            expect(session).toBeDefined();
            expect(session.ready).toBeDefined();
        });
    });

    describe("ready", () => {
        it("should resolve when connection is ready", async () => {
            session = new Session({conn: mockConn as any});
            await expect(session.ready).resolves.toBeUndefined();
        });

        it("should handle connection setup errors", async () => {
            const mockConnWithError = {
                ...mockConn,
                ready: Promise.reject(new Error("Connection failed")),
                close: jest.fn() // Add close method to prevent the error
            };

            session = new Session({conn: mockConnWithError as any});
            await expect(session.ready).rejects.toThrow("Connection failed");
        });
    });

    describe("session initialization", () => {
        beforeEach(() => {
            session = new Session({conn: mockConn as any});
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
            expect(SessionClientMessageMock).toHaveBeenCalled();
        });

        it("should receive session server message", async () => {
            await session.ready;
            const SessionServerMessageMock = jest.mocked(SessionServerMessage);
            expect(SessionServerMessageMock).toHaveBeenCalled();
        });
    });

    describe("internal state", () => {
        beforeEach(() => {
            session = new Session({conn: mockConn as any});
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
            session = new Session({conn: mockConn as any});
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

        it("should close connection with normal closure", async () => {
            const closeSpy = jest.spyOn(mockConn, 'close');
            
            // Mock Promise.allSettled to resolve immediately, avoiding background task wait
            const originalAllSettled = Promise.allSettled;
            jest.spyOn(Promise, 'allSettled').mockResolvedValue([]);
            
            try {
                await session.close();
                expect(closeSpy).toHaveBeenCalledWith({
                    closeCode: 0x0,
                    reason: "No Error"
                });
            } finally {
                // Restore the original Promise.allSettled
                (Promise.allSettled as jest.Mock).mockRestore();
            }
        });

        it("should close connection with error", async () => {
            const closeSpy = jest.spyOn(mockConn, 'close');
            
            // Mock Promise.allSettled to resolve immediately, avoiding background task wait
            const originalAllSettled = Promise.allSettled;
            jest.spyOn(Promise, 'allSettled').mockResolvedValue([]);
            
            try {
                await session.closeWithError(123, "Test error");
                expect(closeSpy).toHaveBeenCalledWith({
                    closeCode: 123,
                    reason: "Test error"
                });
            } finally {
                // Restore the original Promise.allSettled
                (Promise.allSettled as jest.Mock).mockRestore();
            }
        });
    });

    describe("error handling", () => {
        beforeEach(async () => {
            session = new Session({conn: mockConn as any});
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

            const errorSession = new Session({conn: errorMockConn as any});
            await expect(errorSession.ready).rejects.toThrow();
        });

        it("should handle invalid versions", async () => {
            // Create a separate mock for version testing
            const versionMockConn = new MockWebTransport();
            versionMockConn.close = jest.fn();

            // Mock SessionServerMessage to return incompatible version
            const SessionServerMessageMock = jest.mocked(SessionServerMessage);
            // Mock the instance decode method to set the version property
            const mockDecode = jest.fn().mockImplementation(function(this: any) {
                this.version = 999n; // Set incompatible version on the instance
                return Promise.resolve(undefined);
            });
            SessionServerMessageMock.mockImplementationOnce(() => ({
                version: 0xffffff00n, // Initial version (will be overwritten by decode)
                extensions: new Extensions(),
                encode: jest.fn().mockImplementation(() => Promise.resolve(undefined)),
                decode: mockDecode
            } as any));

            const versionSession = new Session({conn: versionMockConn as any});
            await expect(versionSession.ready).rejects.toThrow("Incompatible session version");

            // Verify the decode method was called
            expect(mockDecode).toHaveBeenCalled();
        });
    });

    describe("async operations", () => {
        let readySession: Session;

        beforeEach(async () => {
            readySession = new Session({conn: mockConn as any});
            await readySession.ready;
        });

        it("should handle acceptAnnounce", async () => {
            const prefix = { segments: ["test"] };
            const result = await readySession.acceptAnnounce(prefix as any);
            expect(result).toBeInstanceOf(Array);
            expect(result).toHaveLength(2);
            // Should return [AnnouncementReader?, Error?]
            const [reader, error] = result;
            expect(error).toBeUndefined();
        });

        it("should handle subscribe", async () => {
            const path = { segments: ["test", "path"] };
            const trackName = "test-track";
            const config = { trackPriority: 1, minGroupSequence: 0n, maxGroupSequence: 10n };

            const result = await readySession.subscribe(path as any, trackName, config);
            expect(result).toBeInstanceOf(Array);
            expect(result).toHaveLength(2);
            // Should return [TrackReader?, Error?]
            const [trackReader, error] = result;
            expect(error).toBeUndefined();
        });
    });
});
