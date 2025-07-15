import { AnnouncementWriter, AnnouncementReader, Announcement } from './announce_stream';
import { Writer, Reader } from './io';
import { Context, background, withCancelCause } from './internal';
import { AnnounceMessage, AnnouncePleaseMessage } from './message';
import { TrackPrefix } from './track_prefix';
import { BroadcastPath } from './broadcast_path';

// Mock dependencies
jest.mock('./io');
jest.mock('./message');

describe('AnnouncementWriter', () => {
    let mockWriter: jest.Mocked<Writer>;
    let mockReader: jest.Mocked<Reader>;
    let mockAnnouncePlease: AnnouncePleaseMessage;
    let ctx: Context;
    let writer: AnnouncementWriter;

    beforeEach(() => {
        ctx = background();

        mockWriter = {
            writeBoolean: jest.fn(),
            writeVarint: jest.fn(),
            writeString: jest.fn(),
            writeUint8Array: jest.fn(),
            writeUint8: jest.fn(),
            flush: jest.fn().mockResolvedValue(undefined),
            close: jest.fn().mockResolvedValue(undefined),
            cancel: jest.fn().mockResolvedValue(undefined),
            closed: jest.fn().mockResolvedValue(undefined)
        } as any;

        mockReader = {
            readBoolean: jest.fn(),
            readVarint: jest.fn(),
            readString: jest.fn(),
            readUint8Array: jest.fn(),
            readUint8: jest.fn(),
            copy: jest.fn(),
            fill: jest.fn(),
            cancel: jest.fn().mockResolvedValue(undefined),
            closed: jest.fn().mockResolvedValue(Promise.resolve())
        } as any;

        mockAnnouncePlease = {
            prefix: '/test/' as TrackPrefix
        } as AnnouncePleaseMessage;

        writer = new AnnouncementWriter(ctx, mockWriter, mockReader, mockAnnouncePlease);
    });

    describe('constructor', () => {
        it('should initialize with provided parameters', () => {
            expect(writer).toBeInstanceOf(AnnouncementWriter);
            expect(writer.context).toBeDefined();
        });

        it('should validate prefix', () => {
            const invalidPrefix = 'invalid-prefix' as TrackPrefix;
            const invalidRequest = { prefix: invalidPrefix } as AnnouncePleaseMessage;

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
                ended: jest.fn().mockResolvedValue(undefined),
                fork: jest.fn().mockReturnValue({} as Announcement),
                end: jest.fn()
            } as any;
        });

        it('should send announcement when path matches prefix', async () => {
            (AnnounceMessage.encode as jest.Mock).mockResolvedValue([{}, undefined]);

            await writer.send(mockAnnouncement);

            expect(AnnounceMessage.encode).toHaveBeenCalledWith(mockWriter, 'path', true);
        });

        it('should throw error when path does not match prefix', async () => {
            const differentAnnouncement = {
                broadcastPath: '/different/path' as BroadcastPath,
                isActive: jest.fn().mockReturnValue(true),
                ended: jest.fn().mockResolvedValue(undefined),
                fork: jest.fn().mockReturnValue({} as Announcement),
                end: jest.fn()
            } as any;

            await expect(writer.send(differentAnnouncement)).rejects.toThrow('Path /different/path does not start with prefix /test/');
        });

        it('should throw error when encoding fails', async () => {
            const error = new Error('Encoding failed');
            (AnnounceMessage.encode as jest.Mock).mockResolvedValue([undefined, error]);

            await expect(writer.send(mockAnnouncement)).rejects.toThrow('Failed to write announcement: Error: Encoding failed');
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
        announcement = new Announcement('/test/path' as BroadcastPath, ctx);
    });

    afterEach(() => {
        cancelFunc(new Error('Test cleanup'));
    });

    describe('constructor', () => {
        it('should initialize with provided path and context', () => {
            expect(announcement.broadcastPath).toBe('/test/path');
        });

        it('should validate broadcast path', () => {
            expect(() => new Announcement('invalid-path' as BroadcastPath, ctx)).toThrow();
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

    describe('fork', () => {
        it('should create a forked announcement', () => {
            const forked = announcement.fork();
            expect(forked).toBeInstanceOf(Announcement);
            expect(forked.broadcastPath).toBe(announcement.broadcastPath);
            expect(forked).not.toBe(announcement);
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
    let ctx: Context;
    let reader: AnnouncementReader;

    beforeEach(() => {
        ctx = background();

        mockWriter = {
            writeBoolean: jest.fn(),
            writeVarint: jest.fn(),
            writeString: jest.fn(),
            writeUint8Array: jest.fn(),
            writeUint8: jest.fn(),
            flush: jest.fn().mockResolvedValue(undefined),
            close: jest.fn().mockResolvedValue(undefined),
            cancel: jest.fn().mockResolvedValue(undefined),
            closed: jest.fn().mockResolvedValue(undefined)
        } as any;

        mockReader = {
            readBoolean: jest.fn(),
            readVarint: jest.fn(),
            readString: jest.fn(),
            readUint8Array: jest.fn(),
            readUint8: jest.fn(),
            copy: jest.fn(),
            fill: jest.fn(),
            cancel: jest.fn().mockResolvedValue(undefined),
            closed: jest.fn().mockResolvedValue(Promise.resolve())
        } as any;

        mockAnnouncePlease = {
            prefix: '/test/' as TrackPrefix
        } as AnnouncePleaseMessage;

        reader = new AnnouncementReader(ctx, mockWriter, mockReader, mockAnnouncePlease);
    });

    describe('constructor', () => {
        it('should initialize with provided parameters', () => {
            expect(reader).toBeInstanceOf(AnnouncementReader);
        });
    });
});
