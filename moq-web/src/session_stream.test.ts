import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { SessionStream } from './session_stream';
import type { Context} from './internal/context';
import { background, withCancelCause } from './internal/context';
import type { Writer, Reader } from './io';
import { StreamError } from './io/error';
import { SessionUpdateMessage } from './message/session_update';
import { SessionClientMessage } from './message/session_client';
import { SessionServerMessage } from './message/session_server';
import { Extensions } from './internal/extensions';
import type { Version } from './internal/version';

describe('SessionStream', () => {
    let ctx: Context;
    let mockWriter: Writer;
    let mockReader: Reader;
    let mockClient: SessionClientMessage;
    let mockServer: SessionServerMessage;
    let sessionStream: SessionStream;

    beforeEach(() => {
        ctx = background();
        
        mockWriter = {
            writeBoolean: jest.fn(),
            writeBigVarint: jest.fn(),
            writeString: jest.fn(),
            writeUint8Array: jest.fn(),
            writeUint8: jest.fn(),
            flush: jest.fn(),
            close: jest.fn(),
            cancel: jest.fn(),
            closed: jest.fn()
        } as any;

        mockReader = {
            readBoolean: jest.fn(),
            readBigVarint: jest.fn(),
            readString: jest.fn(),
            readStringArray: jest.fn(),
            readUint8Array: jest.fn(),
            readUint8: jest.fn(),
            readVarint: jest.fn(),
            copy: jest.fn(),
            fill: jest.fn(),
            cancel: jest.fn(),
            closed: jest.fn()
        } as any;

        // Mock readVarint to return [number, Error | undefined]
        (mockReader.readVarint as jest.MockedFunction<any>).mockResolvedValue([0, undefined]);
        (mockReader.readBigVarint as jest.MockedFunction<any>).mockResolvedValue([0n, undefined]);

        const versions = new Set<Version>([0xffffff00n]);
        const extensions = new Extensions();

        mockClient = new SessionClientMessage({ versions, extensions });
        mockServer = new SessionServerMessage({ version: 0xffffff00n, extensions });

        // Mock SessionUpdateMessage.decode to prevent actual decoding
        const originalDecode = SessionUpdateMessage.prototype.decode;
        SessionUpdateMessage.prototype.decode = jest.fn(async () => new Error('Mock decode error to break loop'));
    });

    afterEach(async () => {
        // Cancel context to clean up any background tasks
        if (ctx) {
            const [, cancel] = withCancelCause(ctx);
            cancel(new Error('Test cleanup'));
        }
        
        // Clean up session stream if it exists
        if (sessionStream) {
            try {
                // Wait a bit for any ongoing operations to complete
                await new Promise(resolve => setTimeout(resolve, 10));
            } catch (error) {
                // Ignore cleanup errors
            }
        }
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

        it('should use the provided context', () => {
            expect(sessionStream.context).toBe(ctx);
        });
    });

    describe('update', () => {
        beforeEach(() => {
            sessionStream = new SessionStream(ctx, mockWriter, mockReader, mockClient, mockServer);
            // Mock the encode method on SessionUpdateMessage instances
            const mockEncode = jest.fn().mockImplementation(async () => undefined);
            jest.spyOn(SessionUpdateMessage.prototype, 'encode').mockImplementation(async () => undefined);
        });

        afterEach(async () => {
            jest.restoreAllMocks();
            
            // Clean up session stream
            if (sessionStream) {
                try {
                    const [, cancel] = withCancelCause(sessionStream.context);
                    cancel(new Error('Test cleanup'));
                    await new Promise(resolve => setTimeout(resolve, 10));
                } catch (error) {
                    // Ignore cleanup errors
                }
            }
        });

        it('should encode and send session update message', async () => {
            const bitrate = 1000n;

            await sessionStream.update(bitrate);

            expect(SessionUpdateMessage.prototype.encode).toHaveBeenCalled();
            expect(sessionStream.clientInfo).toBeDefined();
            expect(sessionStream.clientInfo.bitrate).toBe(bitrate);
        });

        it('should throw error when encoding fails', async () => {
            const bitrate = 1000n;
            const error = new Error('Encoding failed');
            
            // Mock encode to return error
            jest.spyOn(SessionUpdateMessage.prototype, 'encode').mockImplementation(async () => error);

            await expect(sessionStream.update(bitrate)).rejects.toThrow('Failed to encode session update message: Error: Encoding failed');
        });
    });

    describe('context cancellation', () => {
        let ctxWithCancel: Context;
        let cancel: (reason: Error) => void;

        beforeEach(() => {
            [ctxWithCancel, cancel] = withCancelCause(background());
            sessionStream = new SessionStream(ctxWithCancel, mockWriter, mockReader, mockClient, mockServer);
        });

        it('should use the provided context that can be cancelled', () => {
            // The session stream should use the provided context
            expect(sessionStream.context).toBeDefined();
            expect(sessionStream.context).toBe(ctxWithCancel);
            
            // Initially the context should not be cancelled
            expect(sessionStream.context.err()).toBeFalsy();
            
            // Cancel the context
            cancel(new Error("Context cancelled"));
            
            // The session context should now be cancelled
            expect(sessionStream.context.err()).toBeTruthy();
        });
    });

    describe('error handling', () => {
        beforeEach(() => {
            sessionStream = new SessionStream(ctx, mockWriter, mockReader, mockClient, mockServer);
        });

        it('should handle StreamError appropriately', () => {
            const code = 500;
            const message = 'Internal error';

            // Create a StreamError instance to test
            const streamError = new StreamError(code, message);
            
            expect(streamError.code).toBe(code);
            expect(streamError.message).toBe(message);
            expect(streamError).toBeInstanceOf(Error);
            expect(streamError).toBeInstanceOf(StreamError);
        });

        it('should handle decode errors in background updates', async () => {
            // Reset the mock to return an error after first call
            let callCount = 0;
            (SessionUpdateMessage.prototype.decode as jest.MockedFunction<any>).mockImplementation(async () => {
                callCount++;
                if (callCount === 1) {
                    return undefined; // First call succeeds
                }
                return new Error('Decode error'); // Subsequent calls fail
            });

            // Wait a bit for the background handler to process
            await new Promise(resolve => setTimeout(resolve, 50));
            
            // The session stream should still be functional
            expect(sessionStream.context).toBeDefined();
            expect(sessionStream.client).toBeDefined();
            expect(sessionStream.server).toBeDefined();
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
            // Mock the encode method on SessionUpdateMessage instances
            jest.spyOn(SessionUpdateMessage.prototype, 'encode').mockImplementation(async () => undefined);
        });

        afterEach(async () => {
            jest.restoreAllMocks();
            
            // Clean up session stream
            if (sessionStream) {
                try {
                    const [, cancel] = withCancelCause(sessionStream.context);
                    cancel(new Error('Test cleanup'));
                    await new Promise(resolve => setTimeout(resolve, 10));
                } catch (error) {
                    // Ignore cleanup errors
                }
            }
        });

        it('should handle complete session lifecycle', async () => {
            const bitrate = 2000n;

            // Update session
            await sessionStream.update(bitrate);
            expect(sessionStream.clientInfo).toBeDefined();
            expect(sessionStream.clientInfo.bitrate).toBe(bitrate);

            // Close session
            // expect(() => sessionStream.close()).not.toThrow();
        });
    });
});
