import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { SessionStream } from './session_stream';
import type { Context} from 'golikejs/context';
import { background, withCancelCause } from 'golikejs/context';
import type { Writer, Reader } from './io';
import { StreamError } from './io/error';
import { SessionUpdateMessage } from './message/session_update';
import { SessionClientMessage } from './message/session_client';
import { SessionServerMessage } from './message/session_server';
import { Extensions } from './internal/extensions';
import type { Version } from './internal/version';
import { EOF } from './io';

describe('SessionStream', () => {
    let ctx: Context;
    let cancelCtx: (reason: Error | undefined) => void;
    let mockWriter: Writer;
    let mockReader: Reader;
    let mockClient: SessionClientMessage;
    let mockServer: SessionServerMessage;
    let sessionStream: SessionStream;

    beforeEach(() => {
        [ctx, cancelCtx] = withCancelCause(background());
        
        mockWriter = {
            writeBoolean: vi.fn(),
            writeBigVarint: vi.fn(),
            writeString: vi.fn(),
            writeUint8Array: vi.fn(),
            writeUint8: vi.fn(),
            writeVarint: vi.fn(),
            flush: vi.fn(),
            close: vi.fn(),
            cancel: vi.fn(),
            closed: vi.fn()
        } as any;

        mockReader = {
            readBoolean: vi.fn(),
            readBigVarint: vi.fn(),
            readString: vi.fn(),
            readStringArray: vi.fn(),
            readUint8Array: vi.fn(),
            readUint8: vi.fn(),
            readVarint: vi.fn(),
            copy: vi.fn(),
            fill: vi.fn(),
            cancel: vi.fn(),
            closed: vi.fn()
        } as any;

        // Mock readVarint to return [number, Error | undefined] - return EOF to stop the loop
        vi.mocked(mockReader.readVarint).mockResolvedValue([0, EOF]);
        vi.mocked(mockReader.readBigVarint).mockResolvedValue([0n, undefined]);

        const versions = new Set<Version>([0xffffff00n]);
        const extensions = new Extensions();

        mockClient = new SessionClientMessage({ versions, extensions });
        mockServer = new SessionServerMessage({ version: 0xffffff00n, extensions });
    });

    afterEach(async () => {
        // Cancel context to stop background operations
        if (cancelCtx) {
            cancelCtx(undefined);
        }
        
        // Give time for cleanup
        await new Promise(resolve => setTimeout(resolve, 10));
        
        vi.restoreAllMocks();
    });

    describe('constructor', () => {
        beforeEach(() => {
            sessionStream = new SessionStream(ctx, mockWriter, mockReader, mockClient, mockServer);
        });

        it('should initialize with provided parameters', () => {
            expect(sessionStream).toBeInstanceOf(SessionStream);
            expect(sessionStream.client).toBe(mockClient);
            expect(sessionStream.server).toBe(mockServer);
            expect(sessionStream.context).toBeDefined();
        });

        it('should use the provided context', () => {
            expect(sessionStream.context).toBe(ctx);
            expect(typeof sessionStream.context.done).toBe('function');
            expect(typeof sessionStream.context.err).toBe('function');
        });
    });

    describe('update', () => {
        beforeEach(() => {
            sessionStream = new SessionStream(ctx, mockWriter, mockReader, mockClient, mockServer);
        });

        it('should encode and send session update message', async () => {
            const bitrate = 1000n;
            const encodeSpy = vi.spyOn(SessionUpdateMessage.prototype, 'encode').mockResolvedValue(undefined);

            await sessionStream.update(bitrate);

            expect(encodeSpy).toHaveBeenCalledWith(mockWriter);
            expect(sessionStream.clientInfo).toBeDefined();
            expect(sessionStream.clientInfo.bitrate).toBe(bitrate);
        });

        it('should throw error when encoding fails', async () => {
            const bitrate = 1000n;
            const error = new Error('Encoding failed');
            
            vi.spyOn(SessionUpdateMessage.prototype, 'encode').mockResolvedValue(error);

            await expect(sessionStream.update(bitrate)).rejects.toThrow('Failed to encode session update message: Error: Encoding failed');
        });

        it('should update clientInfo with the sent message', async () => {
            const bitrate = 2000n;
            vi.spyOn(SessionUpdateMessage.prototype, 'encode').mockResolvedValue(undefined);

            await sessionStream.update(bitrate);

            const clientInfo = sessionStream.clientInfo;
            expect(clientInfo).toBeInstanceOf(SessionUpdateMessage);
            expect(clientInfo.bitrate).toBe(bitrate);
        });
    });

    describe('context', () => {
        beforeEach(() => {
            sessionStream = new SessionStream(ctx, mockWriter, mockReader, mockClient, mockServer);
        });

        it('should return the internal context', () => {
            const context = sessionStream.context;
            
            expect(context).toBeDefined();
            expect(typeof context.done).toBe('function');
            expect(typeof context.err).toBe('function');
        });

        it('should use context cancellation to stop background operations', async () => {
            // The session stream should use the provided context
            expect(sessionStream.context).toBe(ctx);
            
            // Initially the context should not be cancelled
            expect(sessionStream.context.err()).toBeUndefined();
            
            // Cancel the context
            cancelCtx(new Error("Context cancelled"));
            
            // Give minimal time for the background loop to detect cancellation
            await new Promise(resolve => setTimeout(resolve, 5));
            
            // The session context should now be cancelled
            expect(sessionStream.context.err()).toBeInstanceOf(Error);
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
            // Spy on decode to track calls and return errors
            let callCount = 0;
            const decodeSpy = vi.spyOn(SessionUpdateMessage.prototype, 'decode');
            decodeSpy.mockImplementation(async () => {
                callCount++;
                if (callCount === 1) {
                    return undefined; // First call succeeds
                }
                return new Error('Decode error'); // Subsequent calls fail
            });

            // Wait minimal time for the background handler to process
            await new Promise(resolve => setTimeout(resolve, 10));
            
            // The session stream should still be functional
            expect(sessionStream.context).toBeDefined();
            expect(sessionStream.client).toBeDefined();
            expect(sessionStream.server).toBeDefined();
            
            decodeSpy.mockRestore();
        });
    });

    describe('serverInfo getter', () => {
        beforeEach(() => {
            sessionStream = new SessionStream(ctx, mockWriter, mockReader, mockClient, mockServer);
        });

        it('should return the server information property', () => {
            // serverInfo is initially undefined until the background handler
            // receives a SessionUpdateMessage from the server
            const serverInfoResult = sessionStream.serverInfo;
            
            // We can only verify the getter exists and returns the internal state
            // In a real scenario, this would be populated by decode()
            expect(serverInfoResult).toBeUndefined();
        });
    });

    describe('updated method', () => {
        beforeEach(() => {
            sessionStream = new SessionStream(ctx, mockWriter, mockReader, mockClient, mockServer);
        });

        it('should be a function that returns a promise', () => {
            // Verify the method exists and has the correct signature
            expect(typeof sessionStream.updated).toBe('function');
            
            // Note: We cannot easily test the actual waiting behavior without
            // complex coordination with the background handler and proper mutex locking
            // The method uses Cond.wait() which requires the caller to lock the mutex first
        });
    });

    describe('integration', () => {
        beforeEach(() => {
            sessionStream = new SessionStream(ctx, mockWriter, mockReader, mockClient, mockServer);
            // Mock the encode method on SessionUpdateMessage instances
            vi.spyOn(SessionUpdateMessage.prototype, 'encode').mockImplementation(async () => undefined);
        });

        afterEach(() => {
            vi.restoreAllMocks();
        });

        it('should handle complete session lifecycle', async () => {
            const bitrate = 2000n;
            const encodeSpy = vi.mocked(SessionUpdateMessage.prototype.encode);

            // Verify initial state
            expect(sessionStream.context).toBeDefined();
            expect(sessionStream.client).toBe(mockClient);
            expect(sessionStream.server).toBe(mockServer);

            // Update session
            await sessionStream.update(bitrate);
            
            // Verify update was successful - clientInfo is a SessionUpdateMessage
            expect(sessionStream.clientInfo).toBeInstanceOf(SessionUpdateMessage);
            expect(sessionStream.clientInfo.bitrate).toBe(bitrate);
            
            // Verify encode was called
            expect(encodeSpy).toHaveBeenCalled();
        });

        it('should handle multiple sequential updates', async () => {
            const bitrateValues = [1000n, 2000n, 3000n];
            const encodeSpy = vi.mocked(SessionUpdateMessage.prototype.encode);

            for (const bitrate of bitrateValues) {
                await sessionStream.update(bitrate);
                expect(sessionStream.clientInfo).toBeInstanceOf(SessionUpdateMessage);
                expect(sessionStream.clientInfo.bitrate).toBe(bitrate);
            }
            
            // Verify encode was called for each update
            expect(encodeSpy).toHaveBeenCalledTimes(3);
        });
    });
});
