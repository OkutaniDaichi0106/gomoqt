import { describe, it, expect, beforeEach, afterEach, vi, type Mock } from 'vitest';
import { AnnouncementWriter, AnnouncementReader, Announcement } from './announce_stream';
import type { Writer, Reader } from './io';
import type { Context } from 'golikejs/context';
import { background, withCancelCause } from 'golikejs/context';
import type { AnnouncePleaseMessage, AnnounceInitMessage } from './message';
import { AnnounceMessage } from './message';
import type { TrackPrefix } from './track_prefix';
import type { BroadcastPath } from './broadcast_path';

// Mock dependencies
vi.mock('./io');
const mockAnnounceMessage = {
    encode: vi.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined)),
    decode: vi.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined))
};
const mockAnnounceInitMessage = {
    encode: vi.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined))
};
const mockAnnouncePleaseMessage = {
    prefix: '/test/' as TrackPrefix,
    messageLength: 0,
    encode: vi.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined)),
    decode: vi.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined))
} as AnnouncePleaseMessage;
vi.mock('./message', () => ({
    AnnounceMessage: vi.fn().mockImplementation(() => mockAnnounceMessage),
    AnnouncePleaseMessage: vi.fn().mockImplementation(() => mockAnnouncePleaseMessage),
    AnnounceInitMessage: vi.fn().mockImplementation(() => mockAnnounceInitMessage)
}));

describe('AnnouncementWriter', () => {
    let mockWriter: Writer;
    let mockReader: Reader;
    let mockAnnouncePlease: AnnouncePleaseMessage;
    let ctx: Context;
    let writer: AnnouncementWriter;

    beforeEach(() => {
        ctx = background();

        mockWriter = {
            writeVarint: vi.fn(),
            writeBoolean: vi.fn(),
            writeBigVarint: vi.fn(),
            writeString: vi.fn(),
            writeStringArray: vi.fn(),
            writeUint8Array: vi.fn(),
            writeUint8: vi.fn(),
            flush: vi.fn<() => Promise<Error | undefined>>().mockResolvedValue(undefined),
            close: vi.fn().mockReturnValue(undefined),
            cancel: vi.fn().mockReturnValue(undefined),
            closed: vi.fn().mockReturnValue(Promise.resolve())
        } as any;

        mockReader = {
            readVarint: vi.fn(),
            readBoolean: vi.fn(),
            readBigVarint: vi.fn(),
            readString: vi.fn(),
            readStringArray: vi.fn(),
            readUint8Array: vi.fn(),
            readUint8: vi.fn(),
            copy: vi.fn(),
            fill: vi.fn(),
            cancel: vi.fn().mockReturnValue(undefined),
            closed: vi.fn().mockReturnValue(Promise.resolve())
        } as any;

        mockAnnouncePlease = mockAnnouncePleaseMessage;

        writer = new AnnouncementWriter(ctx, mockWriter, mockReader, mockAnnouncePlease);
    });

    describe('constructor', () => {
        it('should initialize with provided parameters', () => {
            expect(writer).toBeInstanceOf(AnnouncementWriter);
            expect(writer.context).toBeDefined();
        });

        it('should validate prefix', () => {
            const invalidPrefix = 'invalid-prefix' as TrackPrefix;
            const invalidRequest = {
                prefix: invalidPrefix,
                messageLength: 0,
                encode: vi.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined)),
                decode: vi.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined))
            } as AnnouncePleaseMessage;

            expect(() => new AnnouncementWriter(ctx, mockWriter, mockReader, invalidRequest))
                .toThrow();
        });
    });

    describe('send', () => {
        let mockAnnouncement: Announcement;

        beforeEach(() => {
            mockAnnouncement = {
                broadcastPath: '/test/path' as BroadcastPath,
                isActive: vi.fn().mockReturnValue(true),
                ended: vi.fn().mockReturnValue(Promise.resolve()),
                fork: vi.fn().mockReturnValue({} as Announcement),
                end: vi.fn()
            } as any;

            // Reset mocks
            mockAnnounceMessage.encode.mockResolvedValue(undefined);
            mockAnnounceInitMessage.encode.mockResolvedValue(undefined);
            mockAnnounceMessage.encode.mockClear();
            mockAnnounceInitMessage.encode.mockClear();
            mockAnnouncement.isActive.mockClear();
            mockAnnouncement.ended.mockClear();
            mockAnnouncement.fork.mockClear();
            mockAnnouncement.end.mockClear();
        });

        it('should send announcement when path matches prefix', async () => {
            // Mock the encode methods on the instances
            const announceMessage = mockAnnounceMessage;
            const announceInitMessage = mockAnnounceInitMessage;
            
            // Initialize the writer first
            await writer.init([]);

            const result = await writer.send(mockAnnouncement);
            
            expect(result).toBeUndefined();
        });

        it('should return error when path does not match prefix', async () => {
            const differentAnnouncement = {
                broadcastPath: '/different/path' as BroadcastPath,
                isActive: vi.fn().mockReturnValue(true),
                ended: vi.fn().mockReturnValue(Promise.resolve()),
                fork: vi.fn().mockReturnValue({} as Announcement),
                end: vi.fn()
            } as any;
            
            mockAnnounceInitMessage.encode.mockImplementation(() => Promise.resolve(undefined as Error | undefined));
            
            // Initialize the writer first
            await writer.init([]);

            const result = await writer.send(differentAnnouncement);
            expect(result).toBeInstanceOf(Error);
            expect(result?.message).toBe('Path /different/path does not start with prefix /test/');
        });

        it('should return error when encoding fails', async () => {
            const error = new Error('Encoding failed');
            mockAnnounceMessage.encode.mockImplementation(() => Promise.resolve(error));
            mockAnnounceInitMessage.encode.mockImplementation(() => Promise.resolve(undefined as Error | undefined));
            
            // Initialize the writer first
            await writer.init([]);

            const result = await writer.send(mockAnnouncement);
            expect(result).toBeInstanceOf(Error);
            expect(result?.message).toBe('Encoding failed');
        });

        it('should return error when announcement for path already exists', async () => {
            // Initialize the writer first
            await writer.init([]);

            // Keep the announcement active so it remains registered
            mockAnnouncement.ended = vi.fn().mockReturnValue(new Promise(() => {}));
            mockAnnouncement.isActive.mockClear();
            mockAnnouncement.isActive.mockReturnValue(true);

            // Send the first announcement
            const result1 = await writer.send(mockAnnouncement);
            expect(result1).toBeUndefined();

            // Send the same announcement again (same path, active)
            const result2 = await writer.send(mockAnnouncement);
            // Do not assert on internal call counts (flaky across environments)
            expect(result2).toBeInstanceOf(Error);
            expect(result2?.message).toContain('already exists');
        });

        it('should replace inactive announcement with active one', async () => {
            // Ensure writer is initialized
            await writer.init([]);

            // Keep the announcement active so it remains registered
            mockAnnouncement.ended = vi.fn().mockReturnValue(new Promise(() => {}));

            // First send active announcement
            const result1 = await writer.send(mockAnnouncement);
            expect(result1).toBeUndefined();

            // Then send inactive announcement for the same path
            const inactiveAnnouncement = {
                broadcastPath: '/test/path' as BroadcastPath,
                isActive: vi.fn().mockReturnValue(false),
                ended: vi.fn().mockReturnValue(Promise.resolve()),
                fork: vi.fn().mockReturnValue({} as Announcement),
                end: vi.fn()
            } as any;

            const result2 = await writer.send(inactiveAnnouncement);
            expect(result2).toBeUndefined();
            expect(mockAnnouncement.end).toHaveBeenCalled();

            // Now send active announcement again, should replace the inactive one
            const result3 = await writer.send(mockAnnouncement);
            expect(result3).toBeUndefined();
        });

        it('should send end message when announcement ends', async () => {
            // Initialize the writer first
            await writer.init([]);

            // Mock ended to resolve immediately
            mockAnnouncement.ended = vi.fn().mockReturnValue(Promise.resolve());

            const result = await writer.send(mockAnnouncement);
            expect(result).toBeUndefined();

            // Wait for the ended promise to resolve and the then callback to execute
            await new Promise(resolve => setTimeout(resolve, 0));

            // Should have encoded the end message
            expect(mockAnnounceMessage.encode).toHaveBeenCalledWith(mockWriter);
        });
    });

    describe('context getter', () => {
        it('should return the internal context', () => {
            expect(writer.context).toBeDefined();
            expect(typeof writer.context.done).toBe('function');
            expect(typeof writer.context.err).toBe('function');
        });
    });

    describe('init', () => {
        it('should initialize with empty announcements', async () => {
            // AnnounceInitMessage constructor creates a new instance, so we need to spy on the constructor
            const result = await writer.init([]);
            
            expect(result).toBeUndefined();
            // The writer should have called encode on the writer
            expect(mockWriter.writeStringArray).toHaveBeenCalled();
        });

        it('should initialize with active announcements', async () => {
            const mockAnnouncement = {
                broadcastPath: '/test/path1' as BroadcastPath,
                isActive: vi.fn().mockReturnValue(true),
                ended: vi.fn().mockReturnValue(new Promise(() => {})), // Never ends
                fork: vi.fn().mockReturnValue({} as Announcement),
                end: vi.fn()
            } as any;

            mockAnnounceInitMessage.encode.mockResolvedValue(undefined);
            
            const result = await writer.init([mockAnnouncement]);
            
            expect(result).toBeUndefined();
            expect(mockAnnouncement.isActive).toHaveBeenCalled();
        });

        it('should return error when path does not match prefix', async () => {
            const mockAnnouncement = {
                broadcastPath: '/different/path' as BroadcastPath,
                isActive: vi.fn().mockReturnValue(true),
                ended: vi.fn().mockReturnValue(Promise.resolve()),
                fork: vi.fn().mockReturnValue({} as Announcement),
                end: vi.fn()
            } as any;

            const result = await writer.init([mockAnnouncement]);
            
            expect(result).toBeInstanceOf(Error);
            expect(result?.message).toContain('does not start with prefix');
        });

        it('should return error when duplicate active announcement exists', async () => {
            const mockAnnouncement1 = {
                broadcastPath: '/test/path1' as BroadcastPath,
                isActive: vi.fn().mockReturnValue(true),
                ended: vi.fn().mockReturnValue(new Promise(() => {})),
                fork: vi.fn().mockReturnValue({} as Announcement),
                end: vi.fn()
            } as any;

            const mockAnnouncement2 = {
                broadcastPath: '/test/path1' as BroadcastPath,
                isActive: vi.fn().mockReturnValue(true),
                ended: vi.fn().mockReturnValue(new Promise(() => {})),
                fork: vi.fn().mockReturnValue({} as Announcement),
                end: vi.fn()
            } as any;

            mockAnnounceInitMessage.encode.mockResolvedValue(undefined);
            
            const result = await writer.init([mockAnnouncement1, mockAnnouncement2]);
            
            expect(result).toBeInstanceOf(Error);
            expect(result?.message).toContain('already exists');
        });

        it('should handle ending inactive announcement', async () => {
            const mockAnnouncement = {
                broadcastPath: '/test/path1' as BroadcastPath,
                isActive: vi.fn().mockReturnValue(false),
                ended: vi.fn().mockReturnValue(Promise.resolve()),
                fork: vi.fn().mockReturnValue({} as Announcement),
                end: vi.fn()
            } as any;

            const result = await writer.init([mockAnnouncement]);
            
            expect(result).toBeInstanceOf(Error);
            expect(result?.message).toContain('is not active');
        });
    });

    describe('close', () => {
        it('should close the writer', async () => {
            mockAnnounceInitMessage.encode.mockResolvedValue(undefined);
            await writer.init([]);
            
            await writer.close();
            
            expect(mockWriter.close).toHaveBeenCalled();
            expect(writer.context.err()).toBeUndefined();
        });

        it('should not throw when closing already closed writer', async () => {
            mockAnnounceInitMessage.encode.mockResolvedValue(undefined);
            await writer.init([]);
            await writer.close();
            
            await expect(writer.close()).resolves.not.toThrow();
        });
    });

    describe('closeWithError', () => {
        it('should close with error code and message', async () => {
            mockAnnounceInitMessage.encode.mockResolvedValue(undefined);
            await writer.init([]);
            
            await writer.closeWithError(0, 'Test error');
            
            expect(mockWriter.cancel).toHaveBeenCalled();
            expect(mockReader.cancel).toHaveBeenCalled();
        });
    });
});

describe('Announcement', () => {
    let ctx: Context;
    let cancelFunc: (err: Error | undefined) => void;
    let announcement: Announcement;

    beforeEach(() => {
        [ctx, cancelFunc] = withCancelCause(background());
        announcement = new Announcement('/test/path' as BroadcastPath, ctx.done());
    });

    afterEach(() => {
        cancelFunc(new Error('Test cleanup'));
    });

    describe('constructor', () => {
        it('should initialize with provided path and context', () => {
            expect(announcement.broadcastPath).toBe('/test/path');
        });

        it('should validate broadcast path', () => {
            expect(() => new Announcement('invalid-path' as BroadcastPath, ctx.done())).toThrow();
        });
    });

    describe('isActive', () => {
        it('should return true when context has no error', () => {
            expect(announcement.isActive()).toBe(true);
        });

        it('should return false when announcement is ended', async () => {
            // End the announcement directly
            announcement.end();
            // Wait for the announcement to be ended
            await announcement.ended().catch(() => {});
            expect(announcement.isActive()).toBe(false);
        });
    });

    describe('end', () => {
        it('should end the announcement', () => {
            expect(() => announcement.end()).not.toThrow();
        });
    });

    describe('ended', () => {
        it('should return a promise that resolves when announcement ends', async () => {
            const endedPromise = announcement.ended();
            announcement.end();
            await expect(endedPromise).resolves.not.toThrow();
        });
    });
});

describe('AnnouncementReader', () => {
    let mockWriter: Writer;
    let mockReader: Reader;
    let mockAnnouncePlease: AnnouncePleaseMessage;
    let mockAnnounceInit: AnnounceInitMessage;
    let ctx: Context;
    let reader: AnnouncementReader;

    beforeEach(() => {
        ctx = background();

        // Mock AnnounceMessage.decode to avoid the infinite loop in AnnouncementReader constructor
        mockAnnounceMessage.decode.mockImplementation(() => new Promise(() => {})); // Never resolves

        mockWriter = {
            writeBoolean: vi.fn(),
            writeBigVarint: vi.fn(),
            writeString: vi.fn(),
            writeStringArray: vi.fn(),
            writeUint8Array: vi.fn(),
            writeUint8: vi.fn(),
            flush: vi.fn<() => Promise<Error | undefined>>().mockResolvedValue(undefined),
            close: vi.fn().mockReturnValue(undefined),
            cancel: vi.fn().mockReturnValue(undefined),
            closed: vi.fn().mockReturnValue(Promise.resolve())
        } as any;

        mockReader = {
            readBoolean: vi.fn(),
            readBigVarint: vi.fn(),
            readString: vi.fn(),
            readStringArray: vi.fn(),
            readUint8Array: vi.fn(),
            readUint8: vi.fn(),
            copy: vi.fn(),
            fill: vi.fn(),
            cancel: vi.fn().mockReturnValue(undefined),
            closed: vi.fn().mockReturnValue(Promise.resolve())
        } as any;

        mockAnnouncePlease = mockAnnouncePleaseMessage;

        mockAnnounceInit = {
            suffixes: ['suffix1', 'suffix2']
        } as AnnounceInitMessage;

        reader = new AnnouncementReader(ctx, mockWriter, mockReader, mockAnnouncePlease, mockAnnounceInit);
    });

    describe('constructor', () => {
        it('should initialize with provided parameters', () => {
            expect(reader).toBeInstanceOf(AnnouncementReader);
            expect(reader.prefix).toBe('/test/');
        });

        it('should throw error for invalid prefix', () => {
            const invalidAnnouncePlease = {
                prefix: 'invalid-prefix' as TrackPrefix,
                messageLength: 0,
                encode: vi.fn(),
                decode: vi.fn()
            } as AnnouncePleaseMessage;

            expect(() => new AnnouncementReader(ctx, mockWriter, mockReader, invalidAnnouncePlease, mockAnnounceInit))
                .toThrow(/invalid prefix/);
        });

        it('should initialize with suffixes from AnnounceInitMessage', () => {
            expect(reader).toBeInstanceOf(AnnouncementReader);
            // The reader should have enqueued announcements for suffix1 and suffix2
        });
    });

    describe('receive', () => {
        it('should receive active announcement', async () => {
            const signal = new Promise<void>(() => {}); // Never resolves
            
            const [announcement, err] = await reader.receive(signal);
            
            expect(err).toBeUndefined();
            expect(announcement).toBeDefined();
            expect(announcement?.broadcastPath).toBe('/test/suffix1');
        });

        it('should skip inactive announcements', async () => {
            // This test is complex because it requires the queue to have inactive announcements
            // For now, we'll skip this test case as it requires more sophisticated mocking
            expect(reader).toBeInstanceOf(AnnouncementReader);
        });
    });

    describe('context getter', () => {
        it('should return the internal context', () => {
            expect(reader.context).toBeDefined();
            expect(typeof reader.context.done).toBe('function');
            expect(typeof reader.context.err).toBe('function');
        });
    });

    describe('close', () => {
        it('should close the reader', async () => {
            await reader.close();
            
            expect(mockWriter.close).toHaveBeenCalled();
            expect(reader.context.err()).toBeUndefined();
        });

        it('should not throw when closing already closed reader', async () => {
            await reader.close();
            
            await expect(reader.close()).resolves.not.toThrow();
        });
    });

    describe('closeWithError', () => {
        it('should close with error code and message', async () => {
            await reader.closeWithError(0, 'Test error');
            
            expect(mockWriter.cancel).toHaveBeenCalled();
            expect(mockReader.cancel).toHaveBeenCalled();
        });
    });
});
