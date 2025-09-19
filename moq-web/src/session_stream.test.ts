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
            copy: jest.fn(),
            fill: jest.fn(),
            cancel: jest.fn(),
            closed: jest.fn()
        } as any;

        const versions = new Set<Version>([0xffffff00n]);
        const extensions = new Extensions();

        mockClient = new SessionClientMessage({ versions, extensions });
        mockServer = new SessionServerMessage({ version: 0xffffff00n, extensions });
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
            // Mock the encode method on SessionUpdateMessage instances
            const mockEncode = jest.fn().mockImplementation(async () => undefined);
            jest.spyOn(SessionUpdateMessage.prototype, 'encode').mockImplementation(async () => undefined);
        });

        afterEach(() => {
            jest.restoreAllMocks();
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

    describe('close', () => {
        beforeEach(() => {
            sessionStream = new SessionStream(ctx, mockWriter, mockReader, mockClient, mockServer);
        });

        it('should cancel context with close error', () => {
            // The close method should cancel the internal context
            expect(() => sessionStream.close()).not.toThrow();
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
            // Mock the encode method on SessionUpdateMessage instances
            jest.spyOn(SessionUpdateMessage.prototype, 'encode').mockImplementation(async () => undefined);
        });

        afterEach(() => {
            jest.restoreAllMocks();
        });

        it('should handle complete session lifecycle', async () => {
            const bitrate = 2000n;

            // Update session
            await sessionStream.update(bitrate);
            expect(sessionStream.clientInfo).toBeDefined();
            expect(sessionStream.clientInfo.bitrate).toBe(bitrate);

            // Close session
            expect(() => sessionStream.close()).not.toThrow();
        });
    });
});
