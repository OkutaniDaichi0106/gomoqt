import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { GroupWriter, GroupReader } from "./group_stream";
import { CancelCauseFunc, Context, withCancelCause, background } from "./internal/context";
import { Reader, Writer } from "./io";
import { StreamError } from "./io/error";
import { GroupMessage } from "./message";
import { Frame } from "./frame";

// Mock dependencies
jest.mock("./internal/context", () => ({
    withCancelCause: jest.fn(),
    background: jest.fn()
}));

jest.mock("./io", () => ({
    Reader: jest.fn(),
    Writer: jest.fn()
}));

jest.mock("./io/error", () => ({
    StreamError: jest.fn()
}));

jest.mock("./message", () => ({
    GroupMessage: jest.fn()
}));

describe("GroupWriter", () => {
    let mockWriter: jest.Mocked<Writer>;
    let mockContext: jest.Mocked<Context>;
    let mockCancelFunc: jest.MockedFunction<CancelCauseFunc>;
    let mockGroup: GroupMessage;
    let groupWriter: GroupWriter;

    beforeEach(() => {
        mockWriter = {
            writeUint8Array: jest.fn(),
            flush: jest.fn<() => Promise<Error | undefined>>().mockResolvedValue(undefined),
            close: jest.fn().mockReturnValue(Promise.resolve()),
            cancel: jest.fn().mockReturnValue(Promise.resolve()),
            closed: jest.fn().mockReturnValue(Promise.resolve())
        } as any;

        mockContext = {
            done: jest.fn<() => Promise<void>>().mockResolvedValue(undefined),
            err: jest.fn().mockReturnValue(null)
        } as any;

        mockCancelFunc = jest.fn();

        mockGroup = {
            sequence: 123n
        } as any;

        (withCancelCause as jest.Mock).mockReturnValue([mockContext, mockCancelFunc]);

        groupWriter = new GroupWriter(mockContext, mockWriter, mockGroup);
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    describe("constructor", () => {
        it("should initialize with provided parameters", () => {
            expect(withCancelCause).toHaveBeenCalledWith(mockContext);
            expect(groupWriter.groupSequence).toBe(123n);
        });

        it("should set up context cancellation handling", () => {
            // The done() method is called asynchronously in the constructor
            // We can't easily test the async behavior in this synchronous test
            expect(withCancelCause).toHaveBeenCalledWith(mockContext);
        });
    });

    describe("groupSequence", () => {
        it("should return the group sequence number", () => {
            expect(groupWriter.groupSequence).toBe(123n);
        });
    });

    describe("write", () => {
        it("should write data and flush successfully", async () => {
            const data = new Uint8Array([1, 2, 3, 4]);
            const frame = new Frame(data);
            mockWriter.flush.mockResolvedValue(undefined);

            const error = await groupWriter.writeFrame(frame);

            expect(mockWriter.writeUint8Array).toHaveBeenCalledWith(frame.bytes);
            expect(mockWriter.flush).toHaveBeenCalled();
            expect(error).toBeUndefined();
        });

        it("should return error if flush fails", async () => {
            const data = new Uint8Array([1, 2, 3, 4]);
            const frame = new Frame(data);
            const flushError = new Error("Flush failed");
            mockWriter.flush.mockResolvedValue(flushError);

            const error = await groupWriter.writeFrame(frame);

            expect(mockWriter.writeUint8Array).toHaveBeenCalledWith(frame.bytes);
            expect(mockWriter.flush).toHaveBeenCalled();
            expect(error).toBe(flushError);
        });
    });

    describe("close", () => {
        it("should close writer and cancel context", () => {
            groupWriter.close();

            expect(mockWriter.close).toHaveBeenCalled();
            expect(mockCancelFunc).toHaveBeenCalledWith(expect.any(Error));
        });
    });

    describe("cancel", () => {
        it("should cancel writer and context with error", () => {
            const code = 404;
            const message = "Not found";

            groupWriter.cancel(code, message);

            expect(mockWriter.cancel).toHaveBeenCalledWith(expect.any(StreamError));
            expect(mockCancelFunc).toHaveBeenCalledWith(expect.any(StreamError));
        });
    });
});

describe("GroupReader", () => {
    let mockReader: jest.Mocked<Reader>;
    let mockContext: jest.Mocked<Context>;
    let mockCancelFunc: jest.MockedFunction<CancelCauseFunc>;
    let mockGroup: GroupMessage;
    let groupReader: GroupReader;

    beforeEach(() => {
        mockReader = {
            readUint8Array: jest.fn(),
            readVarint: jest.fn(),
            fillN: jest.fn(),
            cancel: jest.fn().mockReturnValue(Promise.resolve()),
            closed: jest.fn().mockReturnValue(Promise.resolve())
        } as any;

        mockContext = {
            done: jest.fn<() => Promise<void>>().mockResolvedValue(undefined),
            err: jest.fn().mockReturnValue(null)
        } as any;

        mockCancelFunc = jest.fn();

        mockGroup = {
            sequence: 456n
        } as any;

        (withCancelCause as jest.Mock).mockReturnValue([mockContext, mockCancelFunc]);

        groupReader = new GroupReader(mockContext, mockReader, mockGroup);
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    describe("constructor", () => {
        it("should initialize with provided parameters", () => {
            expect(withCancelCause).toHaveBeenCalledWith(mockContext);
            expect(groupReader.groupSequence).toBe(456n);
        });
    });

    describe("groupSequence", () => {
        it("should return the group sequence number", () => {
            expect(groupReader.groupSequence).toBe(456n);
        });
    });

    describe("read", () => {
        it("should read data successfully", async () => {
            const expectedData = new Uint8Array([1, 2, 3, 4]);

            (mockReader.readVarint as jest.MockedFunction<() => Promise<[number, Error | undefined]>>).mockResolvedValue([expectedData.byteLength, undefined]);
            (mockReader.fillN as jest.MockedFunction<(buf: Uint8Array, len: number) => Promise<Error | undefined>>).mockImplementation(async (buf: Uint8Array, len: number) => {
                buf.set(expectedData.subarray(0, len));
                return undefined;
            });

            const [frame, err] = await groupReader.readFrame();

            expect(mockReader.readVarint).toHaveBeenCalled();
            expect(mockReader.fillN).toHaveBeenCalledWith(frame!.bytes, expectedData.byteLength);
            expect(frame!.bytes.slice(0, expectedData.byteLength)).toEqual(expectedData);
            expect(err).toBeUndefined();
        });

        it("should handle read errors", async () => {
            const readErr = new Error("Read failed");

            (mockReader.readVarint as jest.MockedFunction<() => Promise<[number, Error | undefined]>>).mockResolvedValue([0, readErr]);

            const [frame, err] = await groupReader.readFrame();

            expect(mockReader.readVarint).toHaveBeenCalled();
            expect(frame).toBeUndefined();
            expect(err).toBe(readErr);
        });
    });

    describe("cancel", () => {
        it("should cancel reader with code", () => {
            const code = 404;

            groupReader.cancel(code);

            expect(mockReader.cancel).toHaveBeenCalledWith(expect.any(StreamError));
            expect(mockCancelFunc).toHaveBeenCalledWith(expect.any(StreamError));
        });
    });
});
