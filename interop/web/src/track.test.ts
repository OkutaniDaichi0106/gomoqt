import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { TrackReader, TrackWriter } from './track';
import { GroupReader, GroupWriter } from './group_stream';
import { Context, withCancelCause, background } from './internal/context';
import { ReceiveSubscribeStream, SendSubscribeStream, TrackConfig } from './subscribe_stream';
import { Writer, Reader } from './io';
import { BroadcastPath } from './broadcast_path';
import { Info } from './info';
import { GroupMessage } from './message';

// Mock the GroupMessage module
jest.mock('./message', () => ({
    GroupMessage: {
        encode: jest.fn()
    }
}));

describe('TrackWriter', () => {
    let mockSubscribeStream: any;
    let mockOpenUniStreamFunc: any;
    let trackWriter: TrackWriter;
    let mockWriter: any;
    let mockContext: Context;

    beforeEach(() => {
        mockContext = background();
        mockWriter = {
            writeUint8: jest.fn(),
            flush: jest.fn(),
            close: jest.fn(),
            cancel: jest.fn(),
            closed: jest.fn()
        };

        mockSubscribeStream = {
            context: mockContext,
            subscribeId: 123n,
            trackConfig: {} as TrackConfig,
            accept: jest.fn(),
            closeWithError: jest.fn(),
            close: jest.fn()
        };

        mockOpenUniStreamFunc = jest.fn();
        
        trackWriter = new TrackWriter(
            '/test/path' as BroadcastPath,
            'test-track',
            mockSubscribeStream,
            mockOpenUniStreamFunc
        );
    });

    describe('constructor', () => {
        it('should initialize with provided parameters', () => {
            expect(trackWriter.broadcastPath).toBe('/test/path');
            expect(trackWriter.trackName).toBe('test-track');
            expect(trackWriter.context).toBe(mockContext);
            expect(trackWriter.subscribeId).toBe(123n);
        });
    });

    describe('openGroup', () => {
        beforeEach(() => {
            // Reset all mocks
            jest.clearAllMocks();
            // Setup default successful returns
            mockSubscribeStream.accept.mockResolvedValue(undefined);
            mockOpenUniStreamFunc.mockResolvedValue([mockWriter, undefined]);
            (GroupMessage.encode as jest.Mock).mockImplementation(() => Promise.resolve([{}, undefined]));
        });

        it('should accept subscription and open group successfully', async () => {
            const groupId = 456n;
            
            const [groupWriter, error] = await trackWriter.openGroup(groupId);

            expect(mockSubscribeStream.accept).toHaveBeenCalledWith({
                groupOrder: 0,
                trackPriority: 0
            });
            expect(mockOpenUniStreamFunc).toHaveBeenCalled();
            expect(mockWriter.writeUint8).toHaveBeenCalled();
            expect(GroupMessage.encode).toHaveBeenCalledWith(mockWriter, 123n, groupId);
            expect(groupWriter).toBeInstanceOf(GroupWriter);
            expect(error).toBeUndefined();
        });

        it('should return error if subscription accept fails', async () => {
            const acceptError = new Error('Accept failed');
            mockSubscribeStream.accept.mockResolvedValue(acceptError);

            const [groupWriter, error] = await trackWriter.openGroup(456n);

            expect(groupWriter).toBeUndefined();
            expect(error).toBe(acceptError);
        });

        it('should return error if openUniStreamFunc fails', async () => {
            const streamError = new Error('Stream failed');
            mockOpenUniStreamFunc.mockResolvedValue([undefined, streamError]);

            const [groupWriter, error] = await trackWriter.openGroup(456n);

            expect(groupWriter).toBeUndefined();
            expect(error).toBe(streamError);
        });

        it('should skip accept if already accepted', async () => {
            // First call should accept
            await trackWriter.openGroup(456n);
            expect(mockSubscribeStream.accept).toHaveBeenCalledTimes(1);

            // Second call should not accept again
            mockSubscribeStream.accept.mockClear();
            await trackWriter.openGroup(789n);
            expect(mockSubscribeStream.accept).not.toHaveBeenCalled();
        });
    });

    describe('writeInfo', () => {
        beforeEach(() => {
            mockSubscribeStream.accept.mockResolvedValue(undefined);
        });

        it('should accept subscription with provided info', async () => {
            const info: Info = { groupOrder: 100, trackPriority: 50 };
            
            const error = await trackWriter.writeInfo(info);

            expect(mockSubscribeStream.accept).toHaveBeenCalledWith(info);
            expect(error).toBeUndefined();
        });

        it('should return error if accept fails', async () => {
            const acceptError = new Error('Accept failed');
            mockSubscribeStream.accept.mockResolvedValue(acceptError);
            const info: Info = { groupOrder: 100, trackPriority: 50 };

            const error = await trackWriter.writeInfo(info);

            expect(error).toBe(acceptError);
        });

        it('should not accept again if already accepted', async () => {
            const info: Info = { groupOrder: 100, trackPriority: 50 };
            
            // First call should accept
            await trackWriter.writeInfo(info);
            expect(mockSubscribeStream.accept).toHaveBeenCalledWith(info);

            // Second call should return early
            mockSubscribeStream.accept.mockClear();
            const error = await trackWriter.writeInfo(info);
            expect(mockSubscribeStream.accept).not.toHaveBeenCalled();
            expect(error).toBeUndefined();
        });
    });
});

describe('TrackReader', () => {
    let mockSubscribeStream: any;
    let mockAcceptFunc: any;
    let mockOnCloseFunc: any;
    let trackReader: TrackReader;
    let mockContext: Context;
    let mockReader: any;
    let mockGroupMessage: GroupMessage;

    beforeEach(() => {
        mockContext = background();
        mockReader = {
            cancel: jest.fn(),
            closed: jest.fn()
        };
        
        mockGroupMessage = {} as GroupMessage;

        mockSubscribeStream = {
            context: mockContext,
            trackConfig: {} as TrackConfig,
            update: jest.fn(),
            cancel: jest.fn()
        };

        mockAcceptFunc = jest.fn();
        mockOnCloseFunc = jest.fn();
        
        trackReader = new TrackReader(mockSubscribeStream, mockAcceptFunc, mockOnCloseFunc);
    });

    describe('acceptGroup', () => {
        it('should accept group successfully when context is valid', async () => {
            mockAcceptFunc.mockResolvedValue([mockReader, mockGroupMessage]);

            const [groupReader, error] = await trackReader.acceptGroup();

            expect(mockAcceptFunc).toHaveBeenCalled();
            expect(groupReader).toBeInstanceOf(GroupReader);
            expect(error).toBeUndefined();
        });

        it('should return context error when context is cancelled', async () => {
            const contextError = new Error('Context cancelled');
            const [ctx, cancelFunc] = withCancelCause(background());
            mockSubscribeStream.context = ctx;
            trackReader = new TrackReader(mockSubscribeStream, mockAcceptFunc, mockOnCloseFunc);

            cancelFunc(contextError);
            await new Promise(resolve => setTimeout(resolve, 10));

            const [groupReader, error] = await trackReader.acceptGroup();

            expect(mockAcceptFunc).not.toHaveBeenCalled();
            expect(groupReader).toBeUndefined();
            expect(error).toBe(contextError);
        });

        it('should return error when no group is available', async () => {
            mockAcceptFunc.mockResolvedValue(undefined);

            const [groupReader, error] = await trackReader.acceptGroup();

            expect(groupReader).toBeUndefined();
            expect(error).toBeInstanceOf(Error);
            expect(error?.message).toBe('No group available');
        });
    });

    describe('update', () => {
        it('should call subscribeStream update', async () => {
            const trackPriority = 100n;
            const minGroupSequence = 1n;
            const maxGroupSequence = 10n;
            mockSubscribeStream.update.mockResolvedValue(undefined);

            const error = await trackReader.update(trackPriority, minGroupSequence, maxGroupSequence);

            expect(mockSubscribeStream.update).toHaveBeenCalledWith(trackPriority, minGroupSequence, maxGroupSequence);
            expect(error).toBeUndefined();
        });
    });

    describe('cancel', () => {
        it('should cancel subscribeStream and call onCloseFunc', () => {
            const code = 1;
            const message = 'Test cancellation';

            trackReader.cancel(code, message);

            expect(mockSubscribeStream.cancel).toHaveBeenCalledWith(code, message);
            expect(mockOnCloseFunc).toHaveBeenCalled();
        });
    });
});
