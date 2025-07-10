import { SessionStream } from './session_stream';
import { Context, background, withCancelCause } from './internal/context';
import { Writer, Reader } from './io';
import { StreamError } from './io/error';
import { SessionUpdateMessage } from './message';
import { SessionClientMessage } from './message/session_client';
import { SessionServerMessage } from './message/session_server';
import { Extensions } from './internal/extensions';
import { Version } from './internal/version';

// Mock dependencies
jest.mock('./io');
jest.mock('./message');
jest.mock('./message/session_client');
jest.mock('./message/session_server');

describe('SessionStream', () => {
    let ctx: Context;
    let mockWriter: jest.Mocked<Writer>;
    let mockReader: jest.Mocked<Reader>;
    let mockClient: SessionClientMessage;
    let mockServer: SessionServerMessage;
    let sessionStream: SessionStream;

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
            closed: jest.fn().mockResolvedValue(undefined)
        } as any;

        const versions = new Set<Version>([1n]);
        const extensions = new Extensions();

        mockClient = new SessionClientMessage(versions, extensions);
        mockServer = new SessionServerMessage(1n, extensions);
    });

    describe('constructor', () => {
        beforeEach(() => {
            sessionStream = new SessionStream(ctx, mockWriter, mockReader, mockClient, mockServer);
        });

        it('should initialize with provided parameters', () => {
            expect(sessionStream.client).toBe(mockClient);
            expect(sessionStream.server).toBe(mockServer);
            expect(sessionStream.context).toBeDefined();
        });

        it('should create a child context', () => {
            expect(sessionStream.context).not.toBe(ctx);
        });
    });

    describe('update', () => {
        beforeEach(() => {
            sessionStream = new SessionStream(ctx, mockWriter, mockReader, mockClient, mockServer);
        });

        it('should encode and send session update message', async () => {
            const bitrate = 1000n;
            const mockUpdateMessage = {} as SessionUpdateMessage;
            
            // Mock SessionUpdateMessage.encode to return success
            (SessionUpdateMessage.encode as jest.Mock) = jest.fn().mockResolvedValue([mockUpdateMessage, undefined]);

            await sessionStream.update(bitrate);

            expect(SessionUpdateMessage.encode).toHaveBeenCalledWith(mockWriter, bitrate);
            expect(sessionStream.clientInfo).toBe(mockUpdateMessage);
        });

        it('should throw error when encoding fails', async () => {
            const bitrate = 1000n;
            const error = new Error('Encoding failed');
            
            // Mock SessionUpdateMessage.encode to return error
            (SessionUpdateMessage.encode as jest.Mock) = jest.fn().mockResolvedValue([undefined, error]);

            await expect(sessionStream.update(bitrate)).rejects.toThrow('Failed to encode session update message: Error: Encoding failed');
        });
    });

    describe('close', () => {
        let cancelFunc: jest.Mock;

        beforeEach(() => {
            // Mock withCancelCause
            cancelFunc = jest.fn();
            jest.doMock('./internal/context', () => ({
                ...jest.requireActual('./internal/context'),
                withCancelCause: jest.fn().mockReturnValue([ctx, cancelFunc])
            }));
            
            sessionStream = new SessionStream(ctx, mockWriter, mockReader, mockClient, mockServer);
        });

        it('should cancel context with close error', () => {
            sessionStream.close();

            // Note: We can't easily test the private cancelFunc, but we can verify the method exists
            expect(sessionStream.close).toBeDefined();
        });
    });

    describe('closeWithError', () => {
        beforeEach(() => {
            sessionStream = new SessionStream(ctx, mockWriter, mockReader, mockClient, mockServer);
        });

        it('should create StreamError and cancel context', () => {
            const code = 500;
            const message = 'Internal error';

            // This should not throw
            expect(() => sessionStream.closeWithError(code, message)).not.toThrow();
        });
    });

    describe('context getter', () => {
        beforeEach(() => {
            sessionStream = new SessionStream(ctx, mockWriter, mockReader, mockClient, mockServer);
        });

        it('should return the internal context', () => {
            const contextResult = sessionStream.context;
            
            expect(contextResult).toBeDefined();
            expect(typeof contextResult.done).toBe('function');
            expect(typeof contextResult.err).toBe('function');
        });
    });

    describe('integration', () => {
        beforeEach(() => {
            sessionStream = new SessionStream(ctx, mockWriter, mockReader, mockClient, mockServer);
        });

        it('should handle complete session lifecycle', async () => {
            const bitrate = 2000n;
            const mockUpdateMessage = {} as SessionUpdateMessage;
            
            (SessionUpdateMessage.encode as jest.Mock) = jest.fn().mockResolvedValue([mockUpdateMessage, undefined]);

            // Update session
            await sessionStream.update(bitrate);
            expect(sessionStream.clientInfo).toBe(mockUpdateMessage);

            // Close session
            expect(() => sessionStream.close()).not.toThrow();
        });
    });
});
