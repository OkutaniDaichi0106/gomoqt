import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { AnnouncementWriter, AnnouncementReader, Announcement } from './announce_stream';
import { Writer, Reader } from './io';
import { Context, background, withCancelCause } from './internal';
import { AnnounceMessage, AnnouncePleaseMessage, AnnounceInitMessage } from './message';
import { TrackPrefix } from './track_prefix';
import { BroadcastPath } from './broadcast_path';

// Mock dependencies
jest.mock('./io');
const mockAnnounceMessage = {
    encode: jest.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined)),
    decode: jest.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined))
};
const mockAnnounceInitMessage = {
    encode: jest.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined))
};
const mockAnnouncePleaseMessage = {
    prefix: '/test/' as TrackPrefix,
    messageLength: 0,
    encode: jest.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined)),
    decode: jest.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined))
} as AnnouncePleaseMessage;
jest.mock('./message', () => ({
    AnnounceMessage: jest.fn().mockImplementation(() => mockAnnounceMessage),
    AnnouncePleaseMessage: jest.fn().mockImplementation(() => mockAnnouncePleaseMessage),
    AnnounceInitMessage: jest.fn().mockImplementation(() => mockAnnounceInitMessage)
}));

// Import the mocked module to use in tests
const { AnnounceMessage: MockedAnnounceMessage } = jest.requireActual('./message') as any;

describe('AnnouncementWriter', () => {
    let mockWriter: jest.Mocked<Writer>;
    let mockReader: jest.Mocked<Reader>;
    let mockAnnouncePlease: AnnouncePleaseMessage;
    let ctx: Context;
    let writer: AnnouncementWriter;

    beforeEach(() => {
        ctx = background();

        mockWriter = {
            writeVarint: jest.fn(),
            writeBoolean: jest.fn(),
            writeBigVarint: jest.fn(),
            writeString: jest.fn(),
            writeStringArray: jest.fn(),
            writeUint8Array: jest.fn(),
            writeUint8: jest.fn(),
            flush: jest.fn<() => Promise<Error | undefined>>().mockResolvedValue(undefined),
            close: jest.fn().mockReturnValue(undefined),
            cancel: jest.fn().mockReturnValue(undefined),
            closed: jest.fn().mockReturnValue(Promise.resolve())
        } as any;

        mockReader = {
            readVarint: jest.fn(),
            readBoolean: jest.fn(),
            readBigVarint: jest.fn(),
            readString: jest.fn(),
            readStringArray: jest.fn(),
            readUint8Array: jest.fn(),
            readUint8: jest.fn(),
            copy: jest.fn(),
            fill: jest.fn(),
            cancel: jest.fn().mockReturnValue(undefined),
            closed: jest.fn().mockReturnValue(Promise.resolve())
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
                encode: jest.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined)),
                decode: jest.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined))
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
                isActive: jest.fn().mockReturnValue(true),
                ended: jest.fn().mockReturnValue(Promise.resolve()),
                fork: jest.fn().mockReturnValue({} as Announcement),
                end: jest.fn()
            } as any;
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
                isActive: jest.fn().mockReturnValue(true),
                ended: jest.fn().mockReturnValue(Promise.resolve()),
                fork: jest.fn().mockReturnValue({} as Announcement),
                end: jest.fn()
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
            expect(result?.message).toBe('Failed to write announcement: Error: Encoding failed');
        });
    });

    describe('context getter', () => {
        it('should return the internal context', () => {
            expect(writer.context).toBeDefined();
            expect(typeof writer.context.done).toBe('function');
            expect(typeof writer.context.err).toBe('function');
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

        it('should return false when context has error', () => {
            cancelFunc(new Error('Test error'));
            // Wait a bit for the context to be cancelled
            setTimeout(() => {
                expect(announcement.isActive()).toBe(false);
            }, 10);
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
    let mockWriter: jest.Mocked<Writer>;
    let mockReader: jest.Mocked<Reader>;
    let mockAnnouncePlease: AnnouncePleaseMessage;
    let mockAnnounceInit: AnnounceInitMessage;
    let ctx: Context;
    let reader: AnnouncementReader;

    beforeEach(() => {
        ctx = background();

        // Mock AnnounceMessage.decode to avoid the infinite loop in AnnouncementReader constructor
        mockAnnounceMessage.decode.mockImplementation(() => new Promise(() => {})); // Never resolves

        mockWriter = {
            writeBoolean: jest.fn(),
            writeBigVarint: jest.fn(),
            writeString: jest.fn(),
            writeStringArray: jest.fn(),
            writeUint8Array: jest.fn(),
            writeUint8: jest.fn(),
            flush: jest.fn<() => Promise<Error | undefined>>().mockResolvedValue(undefined),
            close: jest.fn().mockReturnValue(undefined),
            cancel: jest.fn().mockReturnValue(undefined),
            closed: jest.fn().mockReturnValue(Promise.resolve())
        } as any;

        mockReader = {
            readBoolean: jest.fn(),
            readBigVarint: jest.fn(),
            readString: jest.fn(),
            readStringArray: jest.fn(),
            readUint8Array: jest.fn(),
            readUint8: jest.fn(),
            copy: jest.fn(),
            fill: jest.fn(),
            cancel: jest.fn().mockReturnValue(undefined),
            closed: jest.fn().mockReturnValue(Promise.resolve())
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
        });
    });
});
