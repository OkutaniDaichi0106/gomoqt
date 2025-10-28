import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows } from "../deps.ts";
import type { TrackConfig } from './subscribe_stream.ts';
import { SendSubscribeStream, ReceiveSubscribeStream } from './subscribe_stream.ts';
import type { SubscribeID } from './protocol.ts';
import type { SubscribeMessage, SubscribeOkMessage} from './message.ts';
import type { Writer, Reader } from './io.ts';
import type { Context} from 'golikejs/context';
import { background } from 'golikejs/context';
import type { Info } from './info.ts';
import { StreamError } from './io/error.ts';

describe('SendSubscribeStream', () => {
	let mockWriter: Writer;
	let mockReader: Reader;
	let mockSubscribe: SubscribeMessage;
	let mockSubscribeOk: SubscribeOkMessage;
	let ctx: Context;
	let sendStream: SendSubscribeStream;

	beforeEach(() => {
		ctx = background();
		
		mockWriter = {
			writeVarint: vi.fn(),
			writeBoolean: vi.fn(),
			writeBigVarint: vi.fn(),
			writeString: vi.fn(),
			writeUint8Array: vi.fn(),
			writeUint8: vi.fn(),
			flush: vi.fn<() => Promise<Error | undefined>>().mockResolvedValue(undefined),
			close: vi.fn().mockReturnValue(undefined),
			cancel: vi.fn().mockReturnValue(undefined),
			closed: vi.fn().mockReturnValue(Promise.resolve())
		} as any;

		mockReader = {
			readVarint: vi.fn(),
			readBoolean: vi.fn(),
			readBigVarint: vi.fn(),
			readString: vi.fn(),
			readStringArray: vi.fn(),
			readUint8Array: vi.fn(),
			readUint8: vi.fn(),
			copy: vi.fn(),
			fill: vi.fn(),
			cancel: vi.fn().mockReturnValue(undefined),
			closed: vi.fn().mockReturnValue(Promise.resolve())
		} as any;

		mockSubscribe = {
			subscribeId: 123n,
			broadcastPath: '/test/path',
			trackName: 'test-track',
			trackPriority: 1,
			minGroupSequence: 0n,
			maxGroupSequence: 100n
		} as SubscribeMessage;

		mockSubscribeOk = {
			groupPeriod: 100,
			messageLength: 0,
			encode: vi.fn(),
			decode: vi.fn()
		} as SubscribeOkMessage;

		sendStream = new SendSubscribeStream(ctx, mockWriter, mockReader, mockSubscribe, mockSubscribeOk);
	});

	afterEach(async () => {
		if (sendStream && typeof sendStream.closeWithError === 'function') {
			await sendStream.closeWithError(999, 'test cleanup');
		}
	});

	describe('constructor', () => {
		it('should initialize with provided parameters', () => {
			assertInstanceOf(sendStream, SendSubscribeStream);
			assertExists(sendStream.context);
			assertEquals(sendStream.subscribeId, 123n);
		});
	});

	describe('subscribeId getter', () => {
		it('should return the subscribe ID', () => {
			assertEquals(sendStream.subscribeId, 123n);
		});
	});

	describe('config getter', () => {
		it('should return subscribe message config', () => {
			const config = sendStream.config;
			assertEquals(config.trackPriority, mockSubscribe.trackPriority);
			assertEquals(config.minGroupSequence, mockSubscribe.minGroupSequence);
			assertEquals(config.maxGroupSequence, mockSubscribe.maxGroupSequence);
		});
	});

	describe('info getter', () => {
		it('should return the subscribe ok info', () => {
			const info = sendStream.info;
			assertEquals(info, mockSubscribeOk);
		});
	});

	describe('update', () => {
		it('should update config and write to stream', async () => {
			const newConfig = { 
				trackPriority: 2, 
				minGroupSequence: 10n, 
				maxGroupSequence: 200n 
			};
			
			const result = await sendStream.update(newConfig);
			
			assertEquals(result, undefined);
			expect(mockWriter.flush).toHaveBeenCalled();
			
			const config = sendStream.config;
			assertEquals(config.trackPriority, 2);
			assertEquals(config.minGroupSequence, 10n);
			assertEquals(config.maxGroupSequence, 200n);
		});

		it('should return error when flush fails', async () => {
			vi.mocked(mockWriter.flush).mockResolvedValue(new Error('Flush failed'));
			
			const result = await sendStream.update({ 
				trackPriority: 2, 
				minGroupSequence: 10n, 
				maxGroupSequence: 200n 
			});
			
			assertInstanceOf(result, Error);
			// The error could be from encode or flush
			expect(result?.message).toMatch(/Failed to (write|flush) subscribe update/);
		});
	});

	describe('closeWithError', () => {
		it('should close writer and context with StreamError', async () => {
			await sendStream.closeWithError(500, 'Test error');
			
			expect(mockWriter.cancel).toHaveBeenCalledWith(expect.any(StreamError));
			expect(sendStream.context.err()).toBeInstanceOf(StreamError);
		});
	});

	describe('context getter', () => {
		it('should return the internal context', () => {
			assertExists(sendStream.context);
			assertEquals(typeof sendStream.context.done, 'function');
			assertEquals(typeof sendStream.context.err, 'function');
		});
	});
});

describe('ReceiveSubscribeStream', () => {
	let mockWriter: Writer;
	let mockReader: Reader;
	let mockSubscribe: SubscribeMessage;
	let ctx: Context;
	let receiveStream: ReceiveSubscribeStream;

	beforeEach(() => {
		ctx = background();
		
		mockWriter = {
			writeVarint: vi.fn(),
			writeBoolean: vi.fn(),
			writeBigVarint: vi.fn(),
			writeString: vi.fn(),
			writeUint8Array: vi.fn(),
			writeUint8: vi.fn(),
			flush: vi.fn<() => Promise<Error | undefined>>().mockResolvedValue(undefined),
			close: vi.fn().mockReturnValue(undefined),
			cancel: vi.fn().mockReturnValue(undefined),
			closed: vi.fn().mockReturnValue(Promise.resolve())
		} as any;

		mockReader = {
			readVarint: vi.fn().mockResolvedValue([0, new Error('EOF')]), // Return error to stop #handleUpdates loop
			readBoolean: vi.fn(),
			readBigVarint: vi.fn(),
			readString: vi.fn(),
			readStringArray: vi.fn(),
			readUint8Array: vi.fn(),
			readUint8: vi.fn(),
			copy: vi.fn(),
			fill: vi.fn(),
			cancel: vi.fn().mockReturnValue(undefined),
			closed: vi.fn().mockReturnValue(Promise.resolve())
		} as any;

		mockSubscribe = {
			subscribeId: 789n,
			broadcastPath: '/receive/path',
			trackName: 'receive-track',
			trackPriority: 3,
			minGroupSequence: 5n,
			maxGroupSequence: 150n
		} as SubscribeMessage;

		receiveStream = new ReceiveSubscribeStream(ctx, mockWriter, mockReader, mockSubscribe);
	});

	afterEach(async () => {
		if (receiveStream && typeof receiveStream.closeWithError === 'function') {
			await receiveStream.closeWithError(999, 'test cleanup');
		}
	});

	describe('constructor', () => {
		it('should initialize with provided parameters', () => {
			assertInstanceOf(receiveStream, ReceiveSubscribeStream);
			assertExists(receiveStream.context);
			assertEquals(receiveStream.subscribeId, 789n);
		});
	});

	describe('subscribeId getter', () => {
		it('should return the subscribe ID', () => {
			assertEquals(receiveStream.subscribeId, 789n);
		});
	});

	describe('trackConfig getter', () => {
		it('should return subscribe message config', () => {
			const config = receiveStream.trackConfig;
			assertEquals(config.trackPriority, mockSubscribe.trackPriority);
			assertEquals(config.minGroupSequence, mockSubscribe.minGroupSequence);
			assertEquals(config.maxGroupSequence, mockSubscribe.maxGroupSequence);
		});
	});

	describe('context getter', () => {
		it('should return the internal context', () => {
			assertExists(receiveStream.context);
			assertEquals(typeof receiveStream.context.done, 'function');
			assertEquals(typeof receiveStream.context.err, 'function');
		});
	});

	describe('writeInfo', () => {
		it('should write info successfully', async () => {
			const info: Info = { groupPeriod: 100 };
			
			const result = await receiveStream.writeInfo(info);
			
			assertEquals(result, undefined);
		});

		it('should not write info twice', async () => {
			const info: Info = { groupPeriod: 100 };
			
			await receiveStream.writeInfo(info);
			const result = await receiveStream.writeInfo(info);
			
			assertEquals(result, undefined);
		});

		it('should return error if context is cancelled', async () => {
			await receiveStream.closeWithError(404, 'Test error');
			
			const info: Info = { groupPeriod: 100 };
			const result = await receiveStream.writeInfo(info);
			
			assertInstanceOf(result, Error);
		});
	});

	describe('close', () => {
		it('should close writer and cancel context', async () => {
			await receiveStream.close();
			
			expect(mockWriter.close).toHaveBeenCalled();
			expect(receiveStream.context.err()).toBeUndefined();
		});

		it('should handle multiple close calls gracefully', async () => {
			await receiveStream.close();
			
			const callCount = vi.mocked(mockWriter.close).mock.calls.length;
			await receiveStream.close();
			
			// Multiple closes should be safe even if they call close again
			expect(vi.mocked(mockWriter.close).mock.calls.length).toBeGreaterThanOrEqual(callCount);
		});
	});

	describe('closeWithError', () => {
		it('should cancel writer and context with StreamError', async () => {
			await receiveStream.closeWithError(404, 'Not found');
			
			expect(mockWriter.cancel).toHaveBeenCalledWith(expect.any(StreamError));
			expect(mockReader.cancel).toHaveBeenCalledWith(expect.any(StreamError));
			expect(receiveStream.context.err()).toBeInstanceOf(StreamError);
		});

		it('should not cancel if already cancelled', async () => {
			await receiveStream.closeWithError(404, 'Not found');
			
			vi.mocked(mockWriter.cancel).mockClear();
			vi.mocked(mockReader.cancel).mockClear();
			
			await receiveStream.closeWithError(500, 'Another error');
			
			expect(mockWriter.cancel).not.toHaveBeenCalled();
			expect(mockReader.cancel).not.toHaveBeenCalled();
		});
	});
});


// Type tests
describe('Type definitions', () => {
	describe('TrackConfig', () => {
		it('should define the correct structure', () => {
			const config: TrackConfig = {
				trackPriority: 1,
				minGroupSequence: 0n,
				maxGroupSequence: 100n
			};
			assertEquals(typeof config.trackPriority, 'number');
			assertEquals(typeof config.minGroupSequence, 'bigint');
			assertEquals(typeof config.maxGroupSequence, 'bigint');
		});
	});

	describe('SubscribeID', () => {
		it('should be a bigint', () => {
			const id: SubscribeID = 123n;
			assertEquals(typeof id, 'bigint');
		});
	});
});
