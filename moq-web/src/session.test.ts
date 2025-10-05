import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import type { MockedFunction } from 'vitest';
import { Session } from "./session";
import { Version, Versions } from "./internal";
import { AnnouncePleaseMessage, AnnounceInitMessage, GroupMessage,
    SessionClientMessage, SessionServerMessage,
    SubscribeMessage, SubscribeOkMessage } from "./message";
import { Writer, Reader } from "./io";
import { Extensions } from "./internal/extensions";
import { SessionStream } from "./session_stream";
import { background, Context, watchPromise } from "golikejs/context";
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
vi.mock("./session_stream", () => ({
    SessionStream: vi.fn().mockImplementation(() => ({
        context: { 
            done: vi.fn().mockReturnValue(Promise.resolve()),
            err: vi.fn().mockReturnValue(undefined)
        },
        update: vi.fn()
    }))
}));

// Mock messages
vi.mock("./message", () => ({
    SessionClientMessage: vi.fn().mockImplementation((init: any = {}) => ({
        versions: init.versions ?? new Set([0xffffff00n]),
        extensions: init.extensions ?? {},
        encode: vi.fn().mockImplementation(() => Promise.resolve(undefined)),
        decode: vi.fn().mockImplementation(() => Promise.resolve(undefined))
    })),
    SessionServerMessage: vi.fn().mockImplementation((init: any = {}) => ({
        version: init.version ?? 0xffffff00n,
        extensions: init.extensions ?? {},
        encode: vi.fn().mockImplementation(() => Promise.resolve(undefined)),
        decode: vi.fn().mockImplementation(function(this: any) {
            this.version = this.version ?? 0xffffff00n;
            return Promise.resolve(undefined);
        })
    })),
    AnnouncePleaseMessage: vi.fn().mockImplementation((init: any = {}) => ({
        prefix: init.prefix ?? "",
        encode: vi.fn().mockImplementation(() => Promise.resolve(undefined)),
        decode: vi.fn().mockImplementation(() => Promise.resolve(undefined))
    })),
    AnnounceInitMessage: vi.fn().mockImplementation((init: any = {}) => ({
        suffixes: init.suffixes ?? [],
        encode: vi.fn().mockImplementation(() => Promise.resolve(undefined)),
        decode: vi.fn().mockImplementation(() => Promise.resolve(undefined))
    })),
    GroupMessage: vi.fn().mockImplementation((init: any = {}) => ({
        subscribeId: init.subscribeId ?? 1n,
        sequence: init.sequence ?? 0n,
        encode: vi.fn().mockImplementation(() => Promise.resolve(undefined)),
        decode: vi.fn().mockImplementation(() => Promise.resolve(undefined))
    })),
    SubscribeMessage: vi.fn().mockImplementation((init: any = {}) => ({
        subscribeId: init.subscribeId ?? 1n,
        broadcastPath: init.broadcastPath ?? {},
        trackName: init.trackName ?? "",
        encode: vi.fn().mockImplementation(() => Promise.resolve(undefined)),
        decode: vi.fn().mockImplementation(() => Promise.resolve(undefined))
    })),
    SubscribeOkMessage: vi.fn().mockImplementation((init: any = {}) => ({
        groupPeriod: init.groupPeriod ?? 0,
        encode: vi.fn().mockImplementation(() => Promise.resolve(undefined)),
        decode: vi.fn().mockImplementation(() => Promise.resolve(undefined))
    }))
}));

// Mock IO
vi.mock("./io", () => ({
    Writer: vi.fn().mockImplementation(() => ({
        writeUint8: vi.fn(),
        flush: vi.fn().mockImplementation(() => Promise.resolve(null))
    })),
    Reader: vi.fn().mockImplementation(() => ({
        readUint8: vi.fn().mockImplementation(() => Promise.resolve([1, null]))
    }))
}));

// Mock other dependencies
vi.mock("./announce_stream", () => ({
    AnnouncementReader: vi.fn(),
    AnnouncementWriter: vi.fn()
}));

vi.mock("./track_prefix", () => ({
    TrackPrefix: vi.fn()
}));

vi.mock("./subscribe_stream", () => ({
    ReceiveSubscribeStream: vi.fn(),
    SendSubscribeStream: vi.fn(),
    TrackConfig: vi.fn(),
    SubscribeID: vi.fn()
}));

vi.mock("./broadcast_path", () => ({
    BroadcastPath: vi.fn()
}));

vi.mock("./track", () => ({
    TrackReader: vi.fn(),
    TrackWriter: vi.fn()
}));

vi.mock("./group_stream", () => ({
    GroupReader: vi.fn(),
    GroupWriter: vi.fn()
}));

vi.mock("./track_mux", () => ({
    TrackMux: vi.fn().mockImplementation(() => ({
        serveTrack: vi.fn(),
        serveAnnouncement: vi.fn()
    })),
    DefaultTrackMux: {}
}));

vi.mock("./stream_type", () => ({
    BiStreamTypes: {
        SessionStreamType: 1,
        SubscribeStreamType: 2,
        AnnounceStreamType: 3
    },
    UniStreamTypes: {
        GroupStreamType: 1
    }
}));

vi.mock("./internal/queue", () => ({
    Queue: vi.fn().mockImplementation(() => ({
        enqueue: vi.fn(),
        dequeue: vi.fn()
    }))
}));

vi.mock("./info", () => ({
    Info: vi.fn()
}));

// Mock protocol module
vi.mock("./protocol", () => ({
    TrackName: vi.fn(),
    SubscribeID: vi.fn()
}));

// Mock context and background
vi.mock("golikejs/context", () => ({
    background: vi.fn().mockReturnValue({
        done: vi.fn().mockReturnValue(Promise.resolve()),
        err: vi.fn().mockReturnValue(undefined)
    }),
    watchPromise: vi.fn().mockImplementation((ctx, promise) => ({
        done: vi.fn().mockReturnValue(Promise.resolve()),
        err: vi.fn().mockReturnValue(undefined)
    })),
    Context: vi.fn()
}));

// Mock Extensions
vi.mock("./internal/extensions", () => ({
    Extensions: vi.fn().mockImplementation(() => ({
        // Mock implementation
    }))
}));

// Mock internal index
vi.mock("./internal", () => ({
    Version: vi.fn(),
    Versions: {
        DEVELOP: 0xffffff00n
    },
    DEFAULT_CLIENT_VERSIONS: new Set([0xffffff00n]),
    Queue: vi.fn().mockImplementation(() => ({
        enqueue: vi.fn(),
        dequeue: vi.fn()
    }))
}));

describe("Session", () => {
    let mockConn: MockWebTransport;
    let session: Session;

    beforeEach(() => {
        mockConn = new MockWebTransport();
        vi.clearAllMocks();
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
                serveTrack: vi.fn(),
                serveAnnouncement: vi.fn()
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
                close: vi.fn() // Add close method to prevent the error
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
            const SessionStreamMock = vi.mocked(SessionStream);
            expect(SessionStreamMock).toHaveBeenCalled();
        });

        it("should send session client message", async () => {
            await session.ready;
            const SessionClientMessageMock = vi.mocked(SessionClientMessage);
            expect(SessionClientMessageMock).toHaveBeenCalled();
        });

        it("should receive session server message", async () => {
            await session.ready;
            const SessionServerMessageMock = vi.mocked(SessionServerMessage);
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
            const closeSpy = vi.spyOn(mockConn, 'close');
            
            // Mock Promise.allSettled to resolve immediately, avoiding background task wait
            const originalAllSettled = Promise.allSettled;
            vi.spyOn(Promise, 'allSettled').mockResolvedValue([]);
            
            try {
                await session.close();
                expect(closeSpy).toHaveBeenCalledWith({
                    closeCode: 0x0,
                    reason: "No Error"
                });
            } finally {
                // Restore the original Promise.allSettled
                (Promise.allSettled as any).mockRestore();
            }
        });

        it("should close connection with error", async () => {
            const closeSpy = vi.spyOn(mockConn, 'close');
            
            // Mock Promise.allSettled to resolve immediately, avoiding background task wait
            const originalAllSettled = Promise.allSettled;
            vi.spyOn(Promise, 'allSettled').mockResolvedValue([]);
            
            try {
                await session.closeWithError(123, "Test error");
                expect(closeSpy).toHaveBeenCalledWith({
                    closeCode: 123,
                    reason: "Test error"
                });
            } finally {
                // Restore the original Promise.allSettled
                (Promise.allSettled as any).mockRestore();
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
                createBidirectionalStream: vi.fn().mockImplementation(() =>
                    Promise.reject(new Error("Stream creation failed"))
                ),
                close: vi.fn()
            };

            const errorSession = new Session({conn: errorMockConn as any});
            await expect(errorSession.ready).rejects.toThrow();
        });

        it("should handle invalid versions", async () => {
            // Create a separate mock for version testing
            const versionMockConn = new MockWebTransport();
            versionMockConn.close = vi.fn();

            // Mock SessionServerMessage to return incompatible version
            const SessionServerMessageMock = vi.mocked(SessionServerMessage);
            // Mock the instance decode method to set the version property
            const mockDecode = vi.fn().mockImplementation(function(this: any) {
                this.version = 999n; // Set incompatible version on the instance
                return Promise.resolve(undefined);
            });
            SessionServerMessageMock.mockImplementationOnce(() => ({
                version: 0xffffff00n, // Initial version (will be overwritten by decode)
                extensions: new Extensions(),
                encode: vi.fn().mockImplementation(() => Promise.resolve(undefined)),
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

        it("should handle incoming bidirectional subscribe and announce streams", async () => {
            // Prepare a mock connection where incomingBidirectionalStreams.getReader
            // returns one stream (subscribe) and then closes.
            const subscribeReadable = new ReadableStream({ start(controller) { controller.close(); } });
            const announceReadable = new ReadableStream({ start(controller) { controller.close(); } });

            const subscribeStream = {
                writable: new WritableStream({ write() {} }),
                readable: subscribeReadable
            } as any;

            const announceStream = {
                writable: new WritableStream({ write() {} }),
                readable: announceReadable
            } as any;

            let readCallCount = 0;
            const biReader = {
                read: vi.fn().mockImplementation(async () => {
                    readCallCount++;
                    if (readCallCount === 1) return { done: false, value: subscribeStream };
                    if (readCallCount === 2) return { done: false, value: announceStream };
                    return { done: true };
                }),
                releaseLock: vi.fn()
            } as any;

            mockConn.incomingBidirectionalStreams = {
                getReader: () => biReader
            } as any;

            // Make Reader instances return SubscribeStreamType or AnnounceStreamType depending
            // on the stream object passed in by the code under test.
            const ReaderMock = vi.mocked(Reader);
            ReaderMock.mockImplementation((opts: any) => {
                const s = opts?.stream;
                if (s === subscribeStream.readable) {
                    return { readUint8: async () => [2, null] } as any;
                }
                if (s === announceStream.readable) {
                    return { readUint8: async () => [3, null] } as any;
                }
                return { readUint8: async () => [1, null] } as any;
            });

            // Provide a mock mux so we can assert serveTrack/serveAnnouncement were called
            const mockMux = {
                serveTrack: vi.fn(),
                serveAnnouncement: vi.fn()
            } as any;

            // Create a session which will start listeners
            const listenSession = new Session({ conn: mockConn as any, mux: mockMux });
            await listenSession.ready;

            // Give the background listeners a short tick to run through our mocked reads
            await new Promise((res) => setTimeout(res, 10));

            // Verify that mux.serveTrack and serveAnnouncement were called
            expect(mockMux.serveTrack).toHaveBeenCalled();
            expect(mockMux.serveAnnouncement).toHaveBeenCalled();
        });

        it("should handle incoming unidirectional group stream and enqueue message", async () => {
            // Prepare a mock unidirectional stream value that will be passed to Reader
            const uniValue = {} as any;

            let uniReadCount = 0;
            const uniReaderObj = {
                read: vi.fn().mockImplementation(async () => {
                    uniReadCount++;
                    if (uniReadCount === 1) return { done: false, value: uniValue };
                    return { done: true };
                }),
                releaseLock: vi.fn()
            } as any;

            mockConn.incomingUnidirectionalStreams = {
                getReader: () => uniReaderObj
            } as any;

            // Ensure Reader.readUint8 returns GroupStreamType for the uni stream
            const ReaderMock2 = vi.mocked(Reader);
            ReaderMock2.mockImplementationOnce(() => ({ readUint8: async () => [1, null] } as any));

            // Prepare a session and subscribe to create an enqueue function
            const s = new Session({ conn: mockConn as any });
            await s.ready;

            // Call subscribe to register an enqueue function
            const [trackReader, subErr] = await s.subscribe({ segments: ['a'] } as any, 't' as any, { trackPriority: 0 } as any);
            expect(subErr).toBeUndefined();

            // Give the listeners a tick
            await new Promise((res) => setTimeout(res, 10));
        });
    });
});
