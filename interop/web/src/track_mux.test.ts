import { TrackMux, TrackHandler } from './track_mux';
import { Announcement, AnnouncementWriter } from './announce_stream';
import { Publication } from './publication';
import { BroadcastPath } from './broadcast_path';
import { TrackPrefix } from './track_prefix';
import { Context, background, withCancelCause } from './internal/context';
import { PublishController } from './subscribe_stream';
import { TrackWriter } from './track';

// Mock dependencies
jest.mock('./announce_stream');
jest.mock('./subscribe_stream');
jest.mock('./track');

describe('TrackMux', () => {
    let trackMux: TrackMux;
    let mockHandler: TrackHandler;
    let mockPublication: Publication;
    let mockAnnouncement: Announcement;
    let mockAnnouncementWriter: AnnouncementWriter;

    beforeEach(() => {
        trackMux = new TrackMux();
        
        mockHandler = {
            serveTrack: jest.fn()
        };

        mockPublication = {
            broadcastPath: '/test/path' as BroadcastPath,
            trackName: 'test-track',
            controller: {} as PublishController,
            trackWriter: {} as TrackWriter
        };

        const [ctx, cancelFunc] = withCancelCause(background());
        mockAnnouncement = {
            broadcastPath: '/test/path' as BroadcastPath,
            ended: jest.fn().mockResolvedValue(undefined)
        } as any;

        mockAnnouncementWriter = {
            send: jest.fn(),
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
            trackMux.serveTrack(mockPublication);
            expect(mockHandler.serveTrack).toHaveBeenCalledWith(mockPublication);
        });

        it('should notify existing announcers when path matches prefix', () => {
            const prefix = '/test/' as TrackPrefix;
            
            // First register an announcer
            trackMux.serveAnnouncement(mockAnnouncementWriter, prefix);
            
            // Then announce a path that matches the prefix
            trackMux.announce(mockAnnouncement, mockHandler);
            
            expect(mockAnnouncementWriter.send).toHaveBeenCalledWith(mockAnnouncement);
        });

        it('should clean up handler when announcement ends', async () => {
            // Mock the ended promise to resolve immediately
            const resolveEnded = jest.fn();
            mockAnnouncement.ended = jest.fn().mockImplementation(() => {
                return new Promise(resolve => {
                    resolveEnded.mockImplementation(resolve);
                });
            });

            trackMux.announce(mockAnnouncement, mockHandler);
            
            // Initially handler should work
            trackMux.serveTrack(mockPublication);
            expect(mockHandler.serveTrack).toHaveBeenCalledTimes(1);

            // Simulate announcement ending
            resolveEnded();
            await new Promise(resolve => setTimeout(resolve, 10)); // Wait for async cleanup

            // Handler should be removed, so NotFoundHandler should be used
            const notFoundSpy = jest.fn();
            mockHandler.serveTrack = notFoundSpy;
            trackMux.serveTrack(mockPublication);
            
            // Should not call our handler anymore
            expect(notFoundSpy).not.toHaveBeenCalled();
        });
    });

    describe('handlerTrack', () => {
        it('should create announcement and register handler', () => {
            const ctx = background();
            const path = '/test/path' as BroadcastPath;
            
            // Mock Announcement constructor
            const mockAnnouncementConstructor = jest.fn().mockReturnValue(mockAnnouncement);
            (Announcement as jest.MockedClass<typeof Announcement>).mockImplementation(mockAnnouncementConstructor);
            
            trackMux.handlerTrack(ctx, path, mockHandler);
            
            expect(mockAnnouncementConstructor).toHaveBeenCalledWith(path, ctx);
        });
    });

    describe('serveTrack', () => {
        it('should call registered handler for matching path', async () => {
            trackMux.announce(mockAnnouncement, mockHandler);
            
            await trackMux.serveTrack(mockPublication);
            
            expect(mockHandler.serveTrack).toHaveBeenCalledWith(mockPublication);
        });

        it('should call NotFoundHandler for unregistered path', async () => {
            const publicationWithDifferentPath = {
                ...mockPublication,
                broadcastPath: '/different/path' as BroadcastPath
            };
            
            // Since we can't easily spy on the private NotFoundHandler,
            // we just ensure no error is thrown
            await expect(trackMux.serveTrack(publicationWithDifferentPath)).resolves.not.toThrow();
        });
    });

    describe('serveAnnouncement', () => {
        it('should register announcer for valid prefix', async () => {
            const validPrefix = '/test/' as TrackPrefix;
            
            await expect(trackMux.serveAnnouncement(mockAnnouncementWriter, validPrefix))
                .resolves.not.toThrow();
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
        
        const mockPublication = {} as Publication;
        handler.serveTrack(mockPublication);
        
        expect(handler.serveTrack).toHaveBeenCalledWith(mockPublication);
    });
});
