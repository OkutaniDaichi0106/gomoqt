import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { TrackReader, TrackWriter } from './track';
import { GroupReader, GroupWriter } from './group_stream';
import type { Context} from './internal/context';
import { withCancelCause, background,ContextCancelledError } from './internal/context';
import type { TrackConfig } from './subscribe_stream';
import { ReceiveSubscribeStream, SendSubscribeStream } from './subscribe_stream';
import { Writer, Reader } from './io';
import type { BroadcastPath } from './broadcast_path';
import type { Info } from './info';
import { GroupMessage } from './message';

// Mock the GroupMessage module
const mockGroupMessage = {
    encode: jest.fn()
};
jest.mock('./message', () => ({
    GroupMessage: jest.fn().mockImplementation(() => mockGroupMessage)
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
            writeInfo: jest.fn().mockImplementation(() => Promise.resolve(undefined)),
            closeWithError: jest.fn(),
            close: jest.fn(),
            update: jest.fn(),
            info: {} as any,
            _infoWritten: false
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
            mockGroupMessage.encode.mockImplementation((writer: any) => Promise.resolve(undefined));
        });

        it('should accept subscription and open group successfully', async () => {
            const groupId = 456n;

            const [groupWriter, error] = await trackWriter.openGroup(groupId);

            expect(mockSubscribeStream.writeInfo).toHaveBeenCalledWith();
            expect(mockOpenUniStreamFunc).toHaveBeenCalled();
            expect(mockWriter.writeUint8).toHaveBeenCalled();
            expect(GroupMessage).toHaveBeenCalledWith({ subscribeId: 123n, sequence: groupId });
            expect(mockGroupMessage.encode).toHaveBeenCalledWith(mockWriter);
            expect(groupWriter).toBeInstanceOf(GroupWriter);
            expect(error).toBeUndefined();
        });

        it('should return error if subscription accept fails', async () => {
            const acceptError = new Error('Accept failed');
            mockSubscribeStream.writeInfo.mockResolvedValue(acceptError);

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
            expect(mockSubscribeStream.writeInfo).toHaveBeenCalledTimes(1);

            // Second call should still call writeInfo (but it returns early)
            mockSubscribeStream.writeInfo.mockClear();
            await trackWriter.openGroup(789n);
            expect(mockSubscribeStream.writeInfo).toHaveBeenCalledTimes(1);
        });
    });

    describe('writeInfo', () => {
        beforeEach(() => {
            mockSubscribeStream.writeInfo.mockImplementation(() => Promise.resolve(undefined));
        });

        it('should accept subscription with provided info', async () => {
            const info: Info = { groupPeriod: 100 };
            mockSubscribeStream.writeInfo.mockImplementation(() => Promise.resolve(undefined));

            const error = await trackWriter.writeInfo(info);

            expect(mockSubscribeStream.writeInfo).toHaveBeenCalledWith(info);
            expect(error).toBeUndefined();
        });

        it('should return error if accept fails', async () => {
            const acceptError = new Error('Accept failed');
            mockSubscribeStream.writeInfo.mockImplementation(() => Promise.resolve(acceptError));
            const info: Info = { groupPeriod: 100 };

            const error = await trackWriter.writeInfo(info);

            expect(error).toBe(acceptError);
        });

        it('should not accept again if already accepted', async () => {
            const info: Info = { groupPeriod: 100 };

            // First call should accept
            await trackWriter.writeInfo(info);
            expect(mockSubscribeStream.writeInfo).toHaveBeenCalledWith(info);

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
            config: {} as TrackConfig,
            update: jest.fn(),
            cancel: jest.fn(),
            closeWithError: jest.fn(),
            info: {} as any
        };

        mockAcceptFunc = jest.fn();
        mockOnCloseFunc = jest.fn();

        trackReader = new TrackReader(mockSubscribeStream, mockAcceptFunc, mockOnCloseFunc);
    });

    describe('acceptGroup', () => {
        it('should accept group successfully when context is valid', async () => {
            mockAcceptFunc.mockResolvedValue([mockReader, mockGroupMessage]);

            const [groupReader, error] = await trackReader.acceptGroup(Promise.resolve());

            expect(mockAcceptFunc).toHaveBeenCalled();
            expect(groupReader).toBeInstanceOf(GroupReader);
            expect(error).toBeUndefined();
        });

        it('should return context error when context is cancelled', async () => {
            const contextError = ContextCancelledError;
            const [ctx, cancelFunc] = withCancelCause(background());
            mockSubscribeStream.context = ctx;
            trackReader = new TrackReader(mockSubscribeStream, mockAcceptFunc, mockOnCloseFunc);

            cancelFunc(contextError);
            await new Promise(resolve => setTimeout(resolve, 10));

            const [groupReader, error] = await trackReader.acceptGroup(Promise.resolve());

            expect(mockAcceptFunc).not.toHaveBeenCalled();
            expect(groupReader).toBeUndefined();
            expect(error).toBe(contextError);
        });
    });

    describe('update', () => {
        it('should call subscribeStream update', async () => {
            const trackPriority = 100;
            const minGroupSequence = 1n;
            const maxGroupSequence = 10n;
            const config: TrackConfig = { trackPriority, minGroupSequence, maxGroupSequence };
            mockSubscribeStream.update.mockResolvedValue(undefined);

            const error = await trackReader.update(config);

            expect(mockSubscribeStream.update).toHaveBeenCalledWith(config);
            expect(error).toBeUndefined();
        });
    });

    describe('cancel', () => {
        it('should cancel subscribeStream and call onCloseFunc', async () => {
            const code = 1;
            const message = 'Test cancellation';

            await trackReader.closeWithError(code, message);

            expect(mockSubscribeStream.closeWithError).toHaveBeenCalledWith(code, message);
            expect(mockOnCloseFunc).toHaveBeenCalled();
        });
    });
});
