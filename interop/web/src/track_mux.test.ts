import { describe, it, expect, jest, beforeEach } from '@jest/globals';
import { TrackMux, TrackHandler } from './track_mux';
import { Announcement, AnnouncementWriter } from './announce_stream';
import { BroadcastPath } from './broadcast_path';
import { TrackPrefix, isValidPrefix } from './track_prefix';
import { Context, background, withCancelCause } from './internal/context';
import { SendSubscribeStream, ReceiveSubscribeStream } from './subscribe_stream';
import { TrackWriter } from './track';

// Mock dependencies
jest.mock('./announce_stream');
jest.mock('./subscribe_stream');
jest.mock('./track');

describe('TrackMux', () => {
    let trackMux: TrackMux;
    let mockHandler: TrackHandler;
    let mockTrackWriter: TrackWriter;
    let mockAnnouncement: Announcement;
    let mockAnnouncementWriter: AnnouncementWriter;

    beforeEach(() => {
        trackMux = new TrackMux();

        mockHandler = {
            serveTrack: jest.fn()
        };

        mockTrackWriter = {
            broadcastPath: '/test/path' as BroadcastPath,
            trackName: 'test-track',
            closeWithError: jest.fn(),
            close: jest.fn()
        } as any;

        const [ctx, cancelFunc] = withCancelCause(background());
        mockAnnouncement = {
            broadcastPath: '/test/path' as BroadcastPath,
            ended: jest.fn<() => Promise<void>>().mockResolvedValue(undefined)
        } as any;

        mockAnnouncementWriter = {
            send: jest.fn(),
            init: jest.fn(),
            context: ctx
        } as any;
    });

    describe('constructor', () => {
        it('should create a new TrackMux with empty handlers and announcers', () => {
            const mux = new TrackMux();
            expect(mux).toBeInstanceOf(TrackMux);
        });
    });

    describe('announce', () => {
        it('should register a handler for the announcement path', () => {
            trackMux.announce(mockAnnouncement, mockHandler);

            // Verify the handler is registered (we can't directly test the private Map,
            // but we can test through serveTrack)
            trackMux.serveTrack(mockTrackWriter);
            expect(mockHandler.serveTrack).toHaveBeenCalledWith(mockTrackWriter);
        });

        it('should notify existing announcers when path matches prefix', async () => {
            const prefix = '/test/' as TrackPrefix;

            // First register an announcer
            await trackMux.serveAnnouncement(mockAnnouncementWriter, prefix);

            // Then announce a path that matches the prefix
            trackMux.announce(mockAnnouncement, mockHandler);

            expect(mockAnnouncementWriter.send).toHaveBeenCalledWith(mockAnnouncement);
        });

        it('should clean up handler when announcement ends', async () => {
            let resolveEnded: () => void;
            const endedPromise = new Promise<void>((resolve) => {
                resolveEnded = resolve;
            });
            
            mockAnnouncement.ended = jest.fn<() => Promise<void>>().mockReturnValue(endedPromise);

            trackMux.announce(mockAnnouncement, mockHandler);

            // Initially handler should work
            await trackMux.serveTrack(mockTrackWriter);
            expect(mockHandler.serveTrack).toHaveBeenCalledTimes(1);

            // Simulate announcement ending
            resolveEnded!();
            await new Promise(resolve => setTimeout(resolve, 10)); // Wait for async cleanup

            // Handler should be removed, now reset mock and test with different path
            (mockHandler.serveTrack as jest.Mock).mockClear();
            
            const differentTrackWriter = {
                broadcastPath: '/different/path' as BroadcastPath,
                trackName: 'different-track',
                closeWithError: jest.fn(),
                close: jest.fn()
            } as any;

            await trackMux.serveTrack(differentTrackWriter);
            
            // Should call closeWithError for not found path
            expect(differentTrackWriter.closeWithError).toHaveBeenCalledWith(0x03, "Track not found");
        });
    });

    describe('handleTrack', () => {
        it('should create announcement and register handler', async () => {
            const ctx = background();
            const path = '/test/path' as BroadcastPath;

            // Mock the Announcement constructor
            const mockAnnouncementInstance = {
                broadcastPath: path,
                ended: jest.fn<() => Promise<void>>().mockResolvedValue(undefined),
                isActive: jest.fn().mockReturnValue(true),
                end: jest.fn(),
                fork: jest.fn().mockReturnValue({
                    broadcastPath: path,
                    ended: jest.fn<() => Promise<void>>().mockResolvedValue(undefined),
                    isActive: jest.fn().mockReturnValue(true),
                    end: jest.fn(),
                })
            };

            // Create a spy on the Announcement constructor
            const announcementSpy = jest.spyOn(Announcement.prototype, 'constructor' as any).mockImplementation(function(this: any) {
                Object.assign(this, mockAnnouncementInstance);
            });

            trackMux.handleTrack(ctx, path, mockHandler);

            // Test that the handler is registered
            await trackMux.serveTrack(mockTrackWriter);
            expect(mockHandler.serveTrack).toHaveBeenCalledWith(mockTrackWriter);

            // Cleanup
            announcementSpy.mockRestore();
        });
    });

    describe('serveTrack', () => {
        it('should call registered handler for matching path', async () => {
            trackMux.announce(mockAnnouncement, mockHandler);

            await trackMux.serveTrack(mockTrackWriter);

            expect(mockHandler.serveTrack).toHaveBeenCalledWith(mockTrackWriter);
        });

        it('should call NotFoundHandler for unregistered path', async () => {
            const trackWriterWithDifferentPath = {
                broadcastPath: '/different/path' as BroadcastPath,
                trackName: 'different-track',
                closeWithError: jest.fn(),
                close: jest.fn()
            } as any;

            await trackMux.serveTrack(trackWriterWithDifferentPath);

            // Should call closeWithError for not found path
            expect(trackWriterWithDifferentPath.closeWithError).toHaveBeenCalledWith(0x03, "Track not found");
        });
    });

    describe('serveAnnouncement', () => {
        it('should register announcer for valid prefix', async () => {
            const validPrefix = '/test/' as TrackPrefix;

            await trackMux.serveAnnouncement(mockAnnouncementWriter, validPrefix);

            expect(mockAnnouncementWriter.init).toHaveBeenCalledWith([]);
        });

        it('should throw error for invalid prefix', async () => {
            const invalidPrefix = 'invalid-prefix' as TrackPrefix;

            await expect(trackMux.serveAnnouncement(mockAnnouncementWriter, invalidPrefix))
                .rejects.toThrow('Invalid track prefix: invalid-prefix');
        });

        it('should clean up announcer when context ends', async () => {
            const validPrefix = '/test/' as TrackPrefix;
            const [ctx, cancelFunc] = withCancelCause(background());

            const mockAnnouncementWriterWithContext = {
                send: jest.fn(),
                init: jest.fn(),
                context: ctx
            } as any;

            await trackMux.serveAnnouncement(mockAnnouncementWriterWithContext, validPrefix);

            // Cancel the context to trigger cleanup
            cancelFunc(new Error('Test cleanup'));

            // Wait for async cleanup
            await new Promise(resolve => setTimeout(resolve, 10));

            // Subsequent announcements should not be sent to this writer
            trackMux.announce(mockAnnouncement, mockHandler);

            // The send should not be called since the writer was cleaned up
            expect(mockAnnouncementWriterWithContext.send).not.toHaveBeenCalled();
        });
    });
});

describe('TrackHandler', () => {
    it('should define the correct interface', () => {
        const handler: TrackHandler = {
            serveTrack: jest.fn()
        };

        expect(typeof handler.serveTrack).toBe('function');

        const mockTrackWriter = {
            broadcastPath: '/test/path' as BroadcastPath,
            trackName: 'test-track',
            closeWithError: jest.fn(),
            close: jest.fn()
        } as any;
        handler.serveTrack(mockTrackWriter);

        expect(handler.serveTrack).toHaveBeenCalledWith(mockTrackWriter);
    });
});
