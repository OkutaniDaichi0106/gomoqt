import { SendSubscribeStream, ReceiveSubscribeStream, SubscribeController, PublishController, SubscribeConfig, SubscribeID } from './subscribe_stream';
import { SubscribeMessage, SubscribeOkMessage, SubscribeUpdateMessage } from './message';
import { Writer, Reader } from './io';
import { Context, background, withCancelCause } from './internal/context';
import { Info } from './info';
import { StreamError } from './io/error';

// Mock dependencies
jest.mock('./message');
jest.mock('./io');

describe('SendSubscribeStream', () => {
    let mockWriter: jest.Mocked<Writer>;
    let mockReader: jest.Mocked<Reader>;
    let mockSubscribe: SubscribeMessage;
    let mockSubscribeOk: SubscribeOkMessage;
    let ctx: Context;
    let sendStream: SendSubscribeStream;

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

        mockSubscribe = {
            subscribeId: 123n,
            broadcastPath: '/test/path',
            trackName: 'test-track',
            trackPriority: 1n,
            minGroupSequence: 0n,
            maxGroupSequence: 100n
        } as SubscribeMessage;

        mockSubscribeOk = {
            groupOrder: 456n
        } as SubscribeOkMessage;

        sendStream = new SendSubscribeStream(ctx, mockWriter, mockReader, mockSubscribe, mockSubscribeOk);
    });

    describe('constructor', () => {
        it('should initialize with provided parameters', () => {
            expect(sendStream.subscribeId).toBe(123n);
            expect(sendStream.context).toBeDefined();
        });
    });

    describe('subscribeId getter', () => {
        it('should return the subscribe ID from the subscribe message', () => {
            expect(sendStream.subscribeId).toBe(mockSubscribe.subscribeId);
        });
    });

    describe('context getter', () => {
        it('should return the internal context', () => {
            expect(sendStream.context).toBeDefined();
            expect(typeof sendStream.context.done).toBe('function');
            expect(typeof sendStream.context.err).toBe('function');
        });
    });

    describe('subscribeConfig getter', () => {
        it('should return subscribe message config when no update exists', () => {
            const config = sendStream.subscribeConfig;
            expect(config.trackPriority).toBe(mockSubscribe.trackPriority);
            expect(config.minGroupSequence).toBe(mockSubscribe.minGroupSequence);
            expect(config.maxGroupSequence).toBe(mockSubscribe.maxGroupSequence);
        });

        it('should return update config when update exists', async () => {
            const mockUpdate = {
                trackPriority: 2n,
                minGroupSequence: 10n,
                maxGroupSequence: 200n
            } as SubscribeUpdateMessage;

            (SubscribeUpdateMessage.encode as jest.Mock).mockResolvedValue([mockUpdate, undefined]);

            await sendStream.update(2n, 10n, 200n);

            const config = sendStream.subscribeConfig;
            expect(config.trackPriority).toBe(mockUpdate.trackPriority);
            expect(config.minGroupSequence).toBe(mockUpdate.minGroupSequence);
            expect(config.maxGroupSequence).toBe(mockUpdate.maxGroupSequence);
        });
    });

    describe('update', () => {
        it('should encode and send subscribe update message', async () => {
            const mockUpdate = {
                trackPriority: 2n,
                minGroupSequence: 10n,
                maxGroupSequence: 200n
            } as SubscribeUpdateMessage;

            (SubscribeUpdateMessage.encode as jest.Mock).mockResolvedValue([mockUpdate, undefined]);

            await sendStream.update(2n, 10n, 200n);

            expect(SubscribeUpdateMessage.encode).toHaveBeenCalledWith(mockWriter, 2n, 10n, 200n);
            expect(mockWriter.flush).toHaveBeenCalled();
        });

        it('should throw error when encoding fails', async () => {
            const error = new Error('Encoding failed');
            (SubscribeUpdateMessage.encode as jest.Mock).mockResolvedValue([undefined, error]);

            await expect(sendStream.update(2n, 10n, 200n)).rejects.toThrow('Failed to write subscribe update: Error: Encoding failed');
        });

        it('should throw error when flush fails', async () => {
            const mockUpdate = {} as SubscribeUpdateMessage;
            (SubscribeUpdateMessage.encode as jest.Mock).mockResolvedValue([mockUpdate, undefined]);
            mockWriter.flush.mockResolvedValue(new Error('Flush failed'));

            await expect(sendStream.update(2n, 10n, 200n)).rejects.toThrow('Failed to flush subscribe update: Error: Flush failed');
        });
    });

    describe('cancel', () => {
        it('should cancel writer and context with StreamError', () => {
            const code = 500;
            const message = 'Test error';

            expect(() => sendStream.cancel(code, message)).not.toThrow();
            expect(mockWriter.cancel).toHaveBeenCalledWith(expect.any(StreamError));
        });
    });
});

describe('ReceiveSubscribeStream', () => {
    let mockWriter: jest.Mocked<Writer>;
    let mockReader: jest.Mocked<Reader>;
    let mockSubscribe: SubscribeMessage;
    let ctx: Context;
    let receiveStream: ReceiveSubscribeStream;

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

        mockSubscribe = {
            subscribeId: 789n,
            broadcastPath: '/receive/path',
            trackName: 'receive-track',
            trackPriority: 3n,
            minGroupSequence: 5n,
            maxGroupSequence: 150n
        } as SubscribeMessage;

        receiveStream = new ReceiveSubscribeStream(ctx, mockWriter, mockReader, mockSubscribe);
    });

    describe('constructor', () => {
        it('should initialize with provided parameters', () => {
            expect(receiveStream.subscribeId).toBe(789n);
            expect(receiveStream.context).toBeDefined();
        });
    });

    describe('subscribeId getter', () => {
        it('should return the subscribe ID from the subscribe message', () => {
            expect(receiveStream.subscribeId).toBe(mockSubscribe.subscribeId);
        });
    });

    describe('subscribeConfig getter', () => {
        it('should return subscribe message config when no update exists', () => {
            const config = receiveStream.subscribeConfig;
            expect(config.trackPriority).toBe(mockSubscribe.trackPriority);
            expect(config.minGroupSequence).toBe(mockSubscribe.minGroupSequence);
            expect(config.maxGroupSequence).toBe(mockSubscribe.maxGroupSequence);
        });
    });

    describe('context getter', () => {
        it('should return the internal context', () => {
            expect(receiveStream.context).toBeDefined();
            expect(typeof receiveStream.context.done).toBe('function');
            expect(typeof receiveStream.context.err).toBe('function');
        });
    });

    describe('accept', () => {
        it('should encode and send subscribe ok message', async () => {
            const info: Info = {
                groupOrder: 100n,
                trackPriority: 50n
            };

            const mockOk = {
                groupOrder: info.groupOrder
            } as SubscribeOkMessage;

            (SubscribeOkMessage.encode as jest.Mock).mockResolvedValue([mockOk, undefined]);

            await receiveStream.accept(info);

            expect(SubscribeOkMessage.encode).toHaveBeenCalledWith(mockWriter, info.groupOrder);
            expect(mockWriter.flush).toHaveBeenCalled();
        });

        it('should throw error when encoding fails', async () => {
            const info: Info = { groupOrder: 100n, trackPriority: 50n };
            const error = new Error('Encoding failed');
            (SubscribeOkMessage.encode as jest.Mock).mockResolvedValue([undefined, error]);

            await expect(receiveStream.accept(info)).rejects.toThrow('Failed to write subscribe ok: Error: Encoding failed');
        });

        it('should throw error when flush fails', async () => {
            const info: Info = { groupOrder: 100n, trackPriority: 50n };
            const mockOk = {} as SubscribeOkMessage;
            (SubscribeOkMessage.encode as jest.Mock).mockResolvedValue([mockOk, undefined]);
            mockWriter.flush.mockResolvedValue(new Error('Flush failed'));

            await expect(receiveStream.accept(info)).rejects.toThrow('Failed to flush subscribe ok: Error: Flush failed');
        });
    });

    describe('close', () => {
        it('should close writer and cancel context', () => {
            expect(() => receiveStream.close()).not.toThrow();
            expect(mockWriter.close).toHaveBeenCalled();
        });
    });

    describe('closeWithError', () => {
        it('should cancel writer and context with StreamError', () => {
            const code = 404;
            const message = 'Not found';

            expect(() => receiveStream.closeWithError(code, message)).not.toThrow();
            expect(mockWriter.cancel).toHaveBeenCalledWith(expect.any(StreamError));
        });
    });
});

describe('SubscribeConfig type', () => {
    it('should define the correct structure', () => {
        const config: SubscribeConfig = {
            trackPriority: 1n,
            minGroupSequence: 0n,
            maxGroupSequence: 100n
        };

        expect(typeof config.trackPriority).toBe('bigint');
        expect(typeof config.minGroupSequence).toBe('bigint');
        expect(typeof config.maxGroupSequence).toBe('bigint');
    });
});

describe('SubscribeID type', () => {
    it('should be a bigint', () => {
        const id: SubscribeID = 123n;
        expect(typeof id).toBe('bigint');
    });
});
