import { describe, it, expect, beforeEach} from "../deps.ts";
import type { TrackHandler } from './track_mux.ts';
import { TrackMux } from './track_mux.ts';
import type { AnnouncementWriter } from './announce_stream.ts';
import { Announcement } from './announce_stream.ts';
import type { BroadcastPath } from './broadcast_path.ts';
import type { TrackPrefix} from './track_prefix.ts';
import { isValidPrefix } from './track_prefix.ts';
import { Context, background, withCancelCause } from 'golikejs/context';
import { SendSubscribeStream, ReceiveSubscribeStream } from './subscribe_stream.ts';
import type { TrackWriter } from './track.ts';
import { TrackNotFoundErrorCode } from ".";

describe('TrackMux', () => {
    let trackMux: TrackMux;
    let mockHandler: TrackHandler;
    let mockTrackWriter: TrackWriter;
    let mockAnnouncement: Announcement;
    let mockAnnouncementWriter: AnnouncementWriter;
    let cancelFunc: (err: Error | undefined) => void;

    beforeEach(() => {
        trackMux = new TrackMux();

        mockHandler = {
            serveTrack: vi.fn<(ctx: Promise<void>, trackWriter: TrackWriter) => Promise<void>>()
        };

        mockTrackWriter = {
            broadcastPath: '/test/path' as BroadcastPath,
            trackName: 'test-track',
            closeWithError: vi.fn(),
            close: vi.fn()
        } as any;

        const [ctx, _cancelFunc] = withCancelCause(background());
        cancelFunc = _cancelFunc;
        mockAnnouncement = new Announcement('/test/path' as BroadcastPath, ctx.done());

        mockAnnouncementWriter = {
            send: vi.fn(() => Promise.resolve(undefined)),
            init: vi.fn(() => Promise.resolve(undefined)),
            context: ctx,
            close: vi.fn(() => Promise.resolve(undefined))
        } as any;
    });

    describe('constructor', () => {
        it('should create a new TrackMux with empty handlers and announcers', () => {
            const mux = new TrackMux();
            assertInstanceOf(mux, TrackMux);
        });
    });

    describe('announce', () => {
        it('should register a handler for the announcement path', () => {
            trackMux.announce(mockAnnouncement, mockHandler);

            // Verify the handler is registered (we can't directly test the private Map,
            // but we can test through serveTrack)
            trackMux.serveTrack(mockTrackWriter);
            expect(mockHandler.serveTrack).toHaveBeenCalledWith(background().done(), mockTrackWriter);
        });

        it('should notify existing announcers when path matches prefix', async () => {
            const prefix = '/test/' as TrackPrefix;

            // First register an announcer
            const servePromise = trackMux.serveAnnouncement(mockAnnouncementWriter, prefix);

            // Then announce a path that matches the prefix
            trackMux.announce(mockAnnouncement, mockHandler);

            // Cancel the context to complete the serveAnnouncement
            cancelFunc(new Error('Test cleanup'));

            await servePromise;

            expect(mockAnnouncementWriter.send).toHaveBeenCalledWith(mockAnnouncement);
        });

        it('should clean up handler when announcement ends', async () => {
            trackMux.announce(mockAnnouncement, mockHandler);

            // Initially handler should work
            await trackMux.serveTrack(mockTrackWriter);
            expect(mockHandler.serveTrack).toHaveBeenCalledTimes(1);

            // Simulate announcement ending
            cancelFunc(new Error('Test cleanup'));
            await new Promise(resolve => setTimeout(resolve, 10)); // Wait for async cleanup

            // Handler should be removed, now reset mock and test with different path
            (mockHandler.serveTrack as any).mockClear();

            const differentTrackWriter = {
                broadcastPath: '/different/path' as BroadcastPath,
                trackName: 'different-track',
                closeWithError: vi.fn(),
                close: vi.fn()
            } as any;

            await trackMux.serveTrack(differentTrackWriter);

            // Should call closeWithError for not found path
            expect(differentTrackWriter.closeWithError).toHaveBeenCalledWith(TrackNotFoundErrorCode, "Track not found");
        });
    });

    describe('publish', () => {
        it('should create announcement and register handler', async () => {
            const ctx = background();
            const path = '/test/path' as BroadcastPath;

            // Mock the Announcement constructor
            const mockAnnouncementInstance = {
                broadcastPath: path,
                ended: vi.fn<() => Promise<void>>().mockResolvedValue(undefined),
                isActive: vi.fn().mockReturnValue(true),
                end: vi.fn(),
                fork: vi.fn().mockReturnValue({
                    broadcastPath: path,
                    ended: vi.fn<() => Promise<void>>().mockResolvedValue(undefined),
                    isActive: vi.fn().mockReturnValue(true),
                    end: vi.fn(),
                })
            };

            // Create a spy on the Announcement constructor
            const announcementSpy = vi.spyOn(Announcement.prototype, 'constructor' as any).mockImplementation(function(this: any) {
                Object.assign(this, mockAnnouncementInstance);
            });

                        trackMux.publish(ctx.done(), path, mockHandler);

            // Test that the handler is registered
            await trackMux.serveTrack(mockTrackWriter);
            expect(mockHandler.serveTrack).toHaveBeenCalledWith(ctx.done(), mockTrackWriter);

            // Cleanup
            announcementSpy.mockRestore();
        });
    });

    describe('serveTrack', () => {
        it('should call registered handler for matching path', async () => {
            trackMux.announce(mockAnnouncement, mockHandler);

            await trackMux.serveTrack(mockTrackWriter);

            expect(mockHandler.serveTrack).toHaveBeenCalledWith(mockAnnouncement.ended(), mockTrackWriter);
        });

        it('should call NotFoundHandler for unregistered path', async () => {
            const trackWriterWithDifferentPath = {
                broadcastPath: '/different/path' as BroadcastPath,
                trackName: 'different-track',
                closeWithError: vi.fn(),
                close: vi.fn()
            } as any;

            await trackMux.serveTrack(trackWriterWithDifferentPath);

            // Should call closeWithError for not found path
            expect(trackWriterWithDifferentPath.closeWithError).toHaveBeenCalledWith(TrackNotFoundErrorCode, "Track not found");
        });
    });

    describe('serveAnnouncement', () => {
        it('should register announcer for valid prefix', async () => {
            const validPrefix = '/test/' as TrackPrefix;

            const servePromise = trackMux.serveAnnouncement(mockAnnouncementWriter, validPrefix);

            // Cancel the context to complete the serveAnnouncement
            cancelFunc(new Error('Test cleanup'));

            await servePromise;

            expect(mockAnnouncementWriter.init).toHaveBeenCalledWith([]);
        });

        it('should allow serving announcements for invalid-looking prefix (no validation)', async () => {
            const invalidPrefix = 'invalid-prefix' as TrackPrefix;

            const servePromise = trackMux.serveAnnouncement(mockAnnouncementWriter, invalidPrefix);

            // Cancel the context to complete the serveAnnouncement
            cancelFunc(new Error('Test cleanup'));

            await servePromise;

            expect(mockAnnouncementWriter.init).toHaveBeenCalledWith([]);
        });

        it('should clean up announcer when context ends', async () => {
            const validPrefix = '/test/' as TrackPrefix;
            const [ctx, cancelFunc] = withCancelCause(background());

            const mockAnnouncementWriterWithContext = {
                send: vi.fn(),
                init: vi.fn(),
                context: ctx
            } as any;

            const servePromise = trackMux.serveAnnouncement(mockAnnouncementWriterWithContext, validPrefix);

            // Cancel the context to trigger cleanup
            cancelFunc(new Error('Test cleanup'));

            await servePromise;

            // Subsequent announcements should not be sent to this writer
            trackMux.announce(mockAnnouncement, mockHandler);

            // The send should not be called since the writer was cleaned up
            expect(mockAnnouncementWriterWithContext.send).not.toHaveBeenCalled();
        });
    });

    describe('close', () => {
        it('should close all sessions', async () => {
            const validPrefix = '/test' as TrackPrefix;
            const handler: TrackHandler = {
                serveTrack: vi.fn()
            };

            vi.spyOn(mockAnnouncementWriter, 'close').mockResolvedValue(undefined);

            // Serve an announcement to add the writer to announcers
            const servePromise = trackMux.serveAnnouncement(mockAnnouncementWriter, validPrefix);

            // Announce a track to trigger sending to the writer
            const path = '/test/path' as BroadcastPath;
            await trackMux.announce(new Announcement(path, Promise.resolve()), handler);

            // Close the trackMux
            await trackMux.close();

            // Expect the writer's close to be called
            expect(mockAnnouncementWriter.close).toHaveBeenCalled();
        });

        it('should work with no sessions', async () => {
            await expect(trackMux.close()).resolves.toBeUndefined();
        });
    });
});

describe('TrackHandler', () => {
    it('should define the correct interface', () => {
        const handler: TrackHandler = {
            serveTrack: vi.fn<(ctx: Promise<void>, trackWriter: TrackWriter) => Promise<void>>()
        };

        assertEquals(typeof handler.serveTrack, 'function');

        const mockTrackWriter = {
            broadcastPath: '/test/path' as BroadcastPath,
            trackName: 'test-track',
            closeWithError: vi.fn(),
            close: vi.fn()
        } as any;
        handler.serveTrack(background().done(), mockTrackWriter);

        expect(handler.serveTrack).toHaveBeenCalledWith(background().done(), mockTrackWriter);
    });
});
