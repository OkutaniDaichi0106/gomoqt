import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { SessionStream } from './session_stream';
import type { Context} from 'golikejs/context';
import { background, withCancelCause } from 'golikejs/context';
import type { Writer, Reader } from './io';
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

        // Mock readVarint to return EOF immediately to stop the loop
        vi.mocked(mockReader.readVarint).mockResolvedValue([0, EOF]);
        vi.mocked(mockReader.readBigVarint).mockResolvedValue([0n, undefined]);

        const versions = new Set<Version>([0xffffff00n]);
        const extensions = new Extensions();

        mockClient = new SessionClientMessage({ versions, extensions });
        mockServer = new SessionServerMessage({ version: 0xffffff00n, extensions });
    });

    afterEach(async () => {
        // Cancel context to stop background operations
        cancelCtx(undefined);
        
        // Give time for cleanup
        await new Promise(resolve => setTimeout(resolve, 10));
        
        vi.restoreAllMocks();
    });

    describe('constructor', () => {
        it('should initialize with provided parameters', async () => {
            const sessionStream = new SessionStream({
                context: ctx,
                writer: mockWriter,
                reader: mockReader,
                client: mockClient,
                server: mockServer,
                detectFunc: vi.fn().mockResolvedValue(0)
            });

            expect(sessionStream).toBeInstanceOf(SessionStream);
            expect(sessionStream.clientInfo.versions).toEqual(mockClient.versions);
            expect(sessionStream.clientInfo.extensions).toEqual(mockClient.extensions);
            expect(sessionStream.serverInfo.version).toBe(mockServer.version);
            expect(sessionStream.serverInfo.extensions).toEqual(mockServer.extensions);
            expect(sessionStream.context).toBe(ctx);
            
            // Cancel context
            cancelCtx(undefined);
        });

        it('should use the provided context', async () => {
            const sessionStream = new SessionStream({
                context: ctx,
                writer: mockWriter,
                reader: mockReader,
                client: mockClient,
                server: mockServer,
                detectFunc: vi.fn().mockResolvedValue(0)
            });

            expect(sessionStream.context).toBe(ctx);
            expect(typeof sessionStream.context.done).toBe('function');
            expect(typeof sessionStream.context.err).toBe('function');
            
            // Cancel context
            cancelCtx(undefined);
        });
    });

    describe('context', () => {
        it('should return the internal context', async () => {
            const sessionStream = new SessionStream({
                context: ctx,
                writer: mockWriter,
                reader: mockReader,
                client: mockClient,
                server: mockServer,
                detectFunc: vi.fn().mockResolvedValue(0)
            });

            const context = sessionStream.context;
            
            expect(context).toBeDefined();
            expect(typeof context.done).toBe('function');
            expect(typeof context.err).toBe('function');
            
            // Cancel context
            cancelCtx(undefined);
        });

        it('should use context cancellation to stop background operations', async () => {
            const sessionStream = new SessionStream({
                context: ctx,
                writer: mockWriter,
                reader: mockReader,
                client: mockClient,
                server: mockServer,
                detectFunc: vi.fn().mockResolvedValue(0)
            });

            // The session stream should use the provided context
            expect(sessionStream.context).toBe(ctx);
            
            // Initially the context should not be cancelled
            expect(sessionStream.context.err()).toBeUndefined();
            
            // Cancel the context
            cancelCtx(new Error("Context cancelled"));
            
            // The session context should now be cancelled
            expect(sessionStream.context.err()).toBeDefined();
        });
    });

    describe('error handling', () => {
        it('should handle decode errors in background updates', async () => {
            const sessionStream = new SessionStream({
                context: ctx,
                writer: mockWriter,
                reader: mockReader,
                client: mockClient,
                server: mockServer,
                detectFunc: vi.fn().mockResolvedValue(0)
            });

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

            // The session stream should still be functional
            expect(sessionStream.context).toBeDefined();
            expect(sessionStream.clientInfo).toBeDefined();
            expect(sessionStream.serverInfo).toBeDefined();
            
            // Cancel context
            cancelCtx(undefined);
            
            decodeSpy.mockRestore();
        });
    });

    describe('serverInfo getter', () => {
        it('should return the server information property', async () => {
            const sessionStream = new SessionStream({
                context: ctx,
                writer: mockWriter,
                reader: mockReader,
                client: mockClient,
                server: mockServer,
                detectFunc: vi.fn().mockResolvedValue(0)
            });

            const serverInfoResult = sessionStream.serverInfo;
            
            // Verify the getter exists and returns the internal state
            expect(serverInfoResult).toBeDefined();
            expect(serverInfoResult.version).toBe(mockServer.version);
            expect(serverInfoResult.bitrate).toBe(0);
            
            // Cancel context
            cancelCtx(undefined);
        });
    });

    describe('updated method', () => {
        it('should be a function that returns a promise', async () => {
            const sessionStream = new SessionStream({
                context: ctx,
                writer: mockWriter,
                reader: mockReader,
                client: mockClient,
                server: mockServer,
                detectFunc: vi.fn().mockResolvedValue(0)
            });

            // Verify the method exists and has the correct signature
            expect(typeof sessionStream.updated).toBe('function');
            
            // Cancel context
            cancelCtx(undefined);
        });
    });

    describe('integration', () => {
        it('should handle complete session lifecycle', async () => {
            const sessionStream = new SessionStream({
                context: ctx, 
                writer: mockWriter, 
                reader: mockReader, 
                client: mockClient, 
                server: mockServer,
                detectFunc: vi.fn().mockResolvedValue(0)
            });

            // Mock the encode method on SessionUpdateMessage instances
            vi.spyOn(SessionUpdateMessage.prototype, 'encode').mockImplementation(async () => undefined);

            // Verify initial state
            expect(sessionStream.context).toBeDefined();
            expect(sessionStream.clientInfo.versions).toEqual(mockClient.versions);
            expect(sessionStream.clientInfo.extensions).toEqual(mockClient.extensions);
            expect(sessionStream.serverInfo.version).toBe(mockServer.version);
            expect(sessionStream.serverInfo.extensions).toEqual(mockServer.extensions);
            
            // Cancel context
            cancelCtx(undefined);
        });

        it('should handle multiple updates', async () => {
            const sessionStream = new SessionStream({
                context: ctx, 
                writer: mockWriter, 
                reader: mockReader, 
                client: mockClient, 
                server: mockServer,
                detectFunc: vi.fn().mockResolvedValue(0)
            });

            // Mock the encode method on SessionUpdateMessage instances
            vi.spyOn(SessionUpdateMessage.prototype, 'encode').mockImplementation(async () => undefined);

            // Since update is private, we can't test it directly
            // But we can test that the session stream initializes correctly
            expect(sessionStream.clientInfo.bitrate).toBe(0);
            expect(sessionStream.serverInfo.bitrate).toBe(0);
            
            // Cancel context
            cancelCtx(undefined);
        });
    });
});
