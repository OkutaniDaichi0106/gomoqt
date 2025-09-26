import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import type { TrackConfig } from './subscribe_stream';
import { SendSubscribeStream, ReceiveSubscribeStream } from './subscribe_stream';
import type { SubscribeID } from './protocol';
import type { SubscribeMessage, SubscribeOkMessage} from './message';
import { SubscribeUpdateMessage } from './message';
import type { Writer, Reader } from './io';
import type { Context} from './internal/context';
import { background, withCancelCause } from './internal/context';
import type { Info } from './info';
import { StreamError } from './io/error';

// Mock dependencies (announce_stream.test.ts style)
jest.mock('./io');
const mockSubscribeUpdateMessage = {
    trackPriority: 1,
    minGroupSequence: 0n,
    maxGroupSequence: 100n,
    encode: jest.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined)),
    decode: jest.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined))
};
const mockSubscribeMessage = {
    subscribeId: 123n,
    broadcastPath: '/test/path',
    trackName: 'test-track',
    trackPriority: 1,
    minGroupSequence: 0n,
    maxGroupSequence: 100n,
    messageLength: 0,
    encode: jest.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined)),
    decode: jest.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined))
} as SubscribeMessage;
const mockSubscribeOkMessage = {
    encode: jest.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined))
};
jest.mock('./message', () => ({
    SubscribeMessage: jest.fn().mockImplementation(() => mockSubscribeMessage),
    SubscribeOkMessage: jest.fn().mockImplementation(() => mockSubscribeOkMessage),
    SubscribeUpdateMessage: jest.fn().mockImplementation(() => mockSubscribeUpdateMessage)
}));

describe('SendSubscribeStream', () => {
	describe('trackConfig getter', () => {
		it('should return subscribe message config when no update exists', () => {
			const config = sendStream.config;
			expect(config.trackPriority).toBe(mockSubscribe.trackPriority);
			expect(config.minGroupSequence).toBe(mockSubscribe.minGroupSequence);
			expect(config.maxGroupSequence).toBe(mockSubscribe.maxGroupSequence);
		});
		it('should return update config when update exists', async () => {
			mockSubscribeUpdateMessage.encode.mockImplementation(() => Promise.resolve(undefined as Error | undefined));
			await sendStream.update({ trackPriority: 2, minGroupSequence: 10n, maxGroupSequence: 200n });
			const config = sendStream.config;
			expect(config.trackPriority).toBe(2);
			expect(config.minGroupSequence).toBe(10n);
			expect(config.maxGroupSequence).toBe(200n);
		});
	});
		afterEach(() => {
			if (typeof sendStream?.cancel === 'function') {
				sendStream.cancel(999, 'test cleanup');
			}
		});

	describe('update', () => {
		it('should encode and send subscribe update message', async () => {
			mockSubscribeUpdateMessage.encode.mockImplementation(() => Promise.resolve(undefined as Error | undefined));
			const result = await sendStream.update({ trackPriority: 2, minGroupSequence: 10n, maxGroupSequence: 200n });
			expect(mockSubscribeUpdateMessage.encode).toHaveBeenCalled();
			expect(mockWriter.flush).toHaveBeenCalled();
			expect(result).toBeUndefined();
		});
		it('should return error when encoding fails', async () => {
			mockSubscribeUpdateMessage.encode.mockImplementation(() => Promise.resolve(new Error('Encoding failed')));
			const result = await sendStream.update({ trackPriority: 2, minGroupSequence: 10n, maxGroupSequence: 200n });
			expect(result).toBeInstanceOf(Error);
			expect(result?.message).toBe('Failed to write subscribe update: Error: Encoding failed');
		});
	});

	describe('cancel', () => {
		it('should cancel writer and context with StreamError', () => {
			expect(() => sendStream.cancel(500, 'Test error')).not.toThrow();
			expect(mockWriter.cancel).toHaveBeenCalledWith(expect.any(StreamError));
		});
	});

	describe('context getter', () => {
		it('should return the internal context', () => {
			expect(sendStream.context).toBeDefined();
			expect(typeof sendStream.context.done).toBe('function');
			expect(typeof sendStream.context.err).toBe('function');
		});
	});
	let mockWriter: jest.Mocked<Writer>;
	let mockReader: jest.Mocked<Reader>;
	let mockSubscribe: SubscribeMessage;
	let mockSubscribeOk: SubscribeOkMessage;
	let ctx: Context;
	let sendStream: SendSubscribeStream;

	beforeEach(() => {
		ctx = background();
		
		mockWriter = {
			writeVarint: jest.fn(),
			writeBoolean: jest.fn(),
			writeBigVarint: jest.fn(),
			writeString: jest.fn(),
			writeUint8Array: jest.fn(),
			writeUint8: jest.fn(),
			flush: jest.fn<() => Promise<Error | undefined>>().mockResolvedValue(undefined),
			close: jest.fn().mockReturnValue(undefined),
			cancel: jest.fn().mockReturnValue(undefined),
			closed: jest.fn().mockReturnValue(Promise.resolve())
		} as any;
		mockReader = {
			readVarint: jest.fn(),
			readBoolean: jest.fn(),
			readBigVarint: jest.fn(),
			readString: jest.fn(),
			readStringArray: jest.fn(),
			readUint8Array: jest.fn(),
			readUint8: jest.fn(),
			copy: jest.fn(),
			fill: jest.fn(),
			cancel: jest.fn().mockReturnValue(undefined),
			closed: jest.fn().mockReturnValue(Promise.resolve())
		} as any;
		mockSubscribe = mockSubscribeMessage;
		mockSubscribeOk = {} as SubscribeOkMessage;
		sendStream = new SendSubscribeStream(ctx, mockWriter, mockReader, mockSubscribe, mockSubscribeOk);
	});

	describe('constructor', () => {
		it('should initialize with provided parameters', () => {
			expect(sendStream).toBeInstanceOf(SendSubscribeStream);
			expect(sendStream.context).toBeDefined();
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
			writeBigVarint: jest.fn(),
			writeString: jest.fn(),
			writeUint8Array: jest.fn(),
			writeUint8: jest.fn(),
			flush: jest.fn<() => Promise<Error | undefined>>().mockResolvedValue(undefined),
			close: jest.fn().mockReturnValue(undefined),
			cancel: jest.fn().mockReturnValue(undefined),
			closed: jest.fn().mockReturnValue(Promise.resolve())
		} as any;
		mockReader = {
			readVarint: jest.fn(),
			readBoolean: jest.fn(),
			readBigVarint: jest.fn(),
			readString: jest.fn(),
			readStringArray: jest.fn(),
			readUint8Array: jest.fn(),
			readUint8: jest.fn(),
			copy: jest.fn(),
			fill: jest.fn(),
			cancel: jest.fn().mockReturnValue(undefined),
			closed: jest.fn().mockReturnValue(Promise.resolve())
		} as any;
		mockSubscribe = {
			subscribeId: 789n,
			broadcastPath: '/receive/path',
			trackName: 'receive-track',
			trackPriority: 3,
			minGroupSequence: 5n,
			maxGroupSequence: 150n,
			messageLength: 0,
			encode: jest.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined)),
			decode: jest.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined))
		} as SubscribeMessage;
	receiveStream = new ReceiveSubscribeStream(ctx, mockWriter, mockReader, mockSubscribe);
	// Set decode return value to error so the async loop terminates immediately
	mockSubscribeUpdateMessage.decode.mockImplementation(() => Promise.resolve(new Error('mock error')));
	});
		afterEach(() => {
			if (typeof receiveStream?.closeWithError === 'function') {
				receiveStream.closeWithError(999, 'test cleanup');
			}
		});

	describe('constructor', () => {
		it('should initialize with provided parameters', () => {
			expect(receiveStream).toBeInstanceOf(ReceiveSubscribeStream);
			expect(receiveStream.context).toBeDefined();
		});
	});
});


describe('ReceiveSubscribeStream methods', () => {
	let mockWriter: jest.Mocked<Writer>;
	let mockReader: jest.Mocked<Reader>;
	let mockSubscribe: SubscribeMessage;
	let ctx: Context;
	let receiveStream: ReceiveSubscribeStream;

	beforeEach(() => {
		ctx = background();
		
		mockWriter = {
			writeBoolean: jest.fn(),
			writeBigVarint: jest.fn(),
			writeString: jest.fn(),
			writeUint8Array: jest.fn(),
			writeUint8: jest.fn(),
			flush: jest.fn<() => Promise<Error | undefined>>().mockResolvedValue(undefined),
			close: jest.fn().mockReturnValue(undefined),
			cancel: jest.fn().mockReturnValue(undefined),
			closed: jest.fn().mockReturnValue(Promise.resolve())
		} as any;
		mockReader = {
			readVarint: jest.fn(),
			readBoolean: jest.fn(),
			readBigVarint: jest.fn(),
			readString: jest.fn(),
			readStringArray: jest.fn(),
			readUint8Array: jest.fn(),
			readUint8: jest.fn(),
			copy: jest.fn(),
			fill: jest.fn(),
			cancel: jest.fn().mockReturnValue(undefined),
			closed: jest.fn().mockReturnValue(Promise.resolve())
		} as any;
		mockSubscribe = {
			subscribeId: 789n,
			broadcastPath: '/receive/path',
			trackName: 'receive-track',
			trackPriority: 3,
			minGroupSequence: 5n,
			maxGroupSequence: 150n,
			messageLength: 0,
			encode: jest.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined)),
			decode: jest.fn().mockImplementation(() => Promise.resolve(undefined as Error | undefined))
		} as SubscribeMessage;
		receiveStream = new ReceiveSubscribeStream(ctx, mockWriter, mockReader, mockSubscribe);
	});

	afterEach(() => {
		// No cleanup needed since console.error is handled globally
	});

	describe('trackConfig getter', () => {
		it('should return subscribe message config when no update exists', () => {
			const config = receiveStream.trackConfig;
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
		it('should return error when encoding fails', async () => {
			mockSubscribeOkMessage.encode.mockImplementation(() => Promise.resolve(new Error('Encoding failed')));
			const info: Info = { groupPeriod: 100 };
			const result = await receiveStream.writeInfo(info);
			expect(result).toBeInstanceOf(Error);
			expect(result?.message).toBe('moq: failed to encode SUBSCRIBE_OK message: Error: Encoding failed');
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
			expect(() => receiveStream.closeWithError(404, 'Not found')).not.toThrow();
			expect(mockWriter.cancel).toHaveBeenCalledWith(expect.any(StreamError));
		});
	});

	// Type test: TrackConfig
	describe('TrackConfig type', () => {
		it('should define the correct structure', () => {
			const config: TrackConfig = {
				trackPriority: 1,
				minGroupSequence: 0n,
				maxGroupSequence: 100n
			};
			expect(typeof config.trackPriority).toBe('number');
			expect(typeof config.minGroupSequence).toBe('bigint');
			expect(typeof config.maxGroupSequence).toBe('bigint');
		});
	});

	// Type test: SubscribeID
	describe('SubscribeID type', () => {
		it('should be a bigint', () => {
			const id: SubscribeID = 123n;
			expect(typeof id).toBe('bigint');
		});
	});
});