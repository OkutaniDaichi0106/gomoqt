import { TrackReader, TrackWriter } from './track';
import { GroupReader, GroupWriter } from './group_stream';
import { Context, withCancelCause, background } from './internal/context';

describe('TrackWriter', () => {
    let ctx: Context;
    let mockOpenGroupFunc: (trackCtx: Context, groupId: bigint) => Promise<[GroupWriter?, Error?]>;
    let trackWriter: TrackWriter;

    beforeEach(() => {
        ctx = background();
        mockOpenGroupFunc = jest.fn();
        trackWriter = new TrackWriter(ctx, mockOpenGroupFunc);
    });

    describe('constructor', () => {
        it('should initialize with provided context and open group function', () => {
            expect(trackWriter.context).toBe(ctx);
        });
    });

    describe('context getter', () => {
        it('should return the context', () => {
            expect(trackWriter.context).toBe(ctx);
        });
    });

    describe('openGroup', () => {
        it('should call openGroupFunc with context and groupId', async () => {
            const groupId = 123n;
            const mockGroupWriter = {} as GroupWriter;
            mockOpenGroupFunc = jest.fn().mockResolvedValue([mockGroupWriter, undefined]);
            trackWriter = new TrackWriter(ctx, mockOpenGroupFunc);

            const [groupWriter, error] = await trackWriter.openGroup(groupId);

            expect(mockOpenGroupFunc).toHaveBeenCalledWith(ctx, groupId);
            expect(groupWriter).toBe(mockGroupWriter);
            expect(error).toBeUndefined();
        });

        it('should return error from openGroupFunc', async () => {
            const groupId = 123n;
            const mockError = new Error('Failed to open group');
            mockOpenGroupFunc = jest.fn().mockResolvedValue([undefined, mockError]);
            trackWriter = new TrackWriter(ctx, mockOpenGroupFunc);

            const [groupWriter, error] = await trackWriter.openGroup(groupId);

            expect(mockOpenGroupFunc).toHaveBeenCalledWith(ctx, groupId);
            expect(groupWriter).toBeUndefined();
            expect(error).toBe(mockError);
        });
    });
});

describe('TrackReader', () => {
    let ctx: Context;
    let cancelFunc: (err: Error | null) => void;
    let mockAcceptFunc: () => Promise<[GroupReader?, Error?]>;
    let trackReader: TrackReader;

    beforeEach(() => {
        [ctx, cancelFunc] = withCancelCause(background());
        mockAcceptFunc = jest.fn();
        trackReader = new TrackReader(ctx, mockAcceptFunc);
    });

    afterEach(() => {
        cancelFunc(new Error('Test cleanup'));
    });

    describe('constructor', () => {
        it('should initialize with provided context and accept function', () => {
            expect(trackReader.context).toBe(ctx);
        });
    });

    describe('context getter', () => {
        it('should return the context', () => {
            expect(trackReader.context).toBe(ctx);
        });
    });

    describe('acceptGroup', () => {
        it('should call acceptFunc when context has no error', async () => {
            const mockGroupReader = {} as GroupReader;
            mockAcceptFunc = jest.fn().mockResolvedValue([mockGroupReader, undefined]);
            trackReader = new TrackReader(ctx, mockAcceptFunc);

            const [groupReader, error] = await trackReader.acceptGroup();

            expect(mockAcceptFunc).toHaveBeenCalled();
            expect(groupReader).toBe(mockGroupReader);
            expect(error).toBeUndefined();
        });

        it('should return error from acceptFunc', async () => {
            const mockError = new Error('Failed to accept group');
            mockAcceptFunc = jest.fn().mockResolvedValue([undefined, mockError]);
            trackReader = new TrackReader(ctx, mockAcceptFunc);

            const [groupReader, error] = await trackReader.acceptGroup();

            expect(mockAcceptFunc).toHaveBeenCalled();
            expect(groupReader).toBeUndefined();
            expect(error).toBe(mockError);
        });

        it('should return context error when context is cancelled', async () => {
            const contextError = new Error('Context cancelled');
            cancelFunc(contextError); // Cancel the context

            // Wait a bit for the context to be cancelled
            await new Promise(resolve => setTimeout(resolve, 10));

            const [groupReader, error] = await trackReader.acceptGroup();

            expect(mockAcceptFunc).not.toHaveBeenCalled();
            expect(groupReader).toBeUndefined();
            expect(error).toBeDefined();
            expect(error?.message).toContain('cancelled');
        });
    });
});
