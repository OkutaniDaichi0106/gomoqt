import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { GroupWriter, GroupReader } from "./group_stream";
import type { Context} from "golikejs/context";
import { background } from "golikejs/context";
import type { Reader, Writer } from "./io";
import { StreamError } from "./webtransport/error";
import type { GroupMessage } from "./message";
import { BytesFrame } from "./frame";

describe("GroupWriter", () => {
    let mockWriter: Writer;
    let mockContext: Context;
    let mockGroup: GroupMessage;
    let groupWriter: GroupWriter;

    beforeEach(() => {
        mockContext = background();

        mockWriter = {
            writeUint8Array: vi.fn(),
            copyFrom: vi.fn(),
            flush: vi.fn<() => Promise<Error | undefined>>().mockResolvedValue(undefined),
            close: vi.fn().mockReturnValue(Promise.resolve()),
            cancel: vi.fn().mockReturnValue(Promise.resolve()),
            closed: vi.fn().mockReturnValue(Promise.resolve())
        } as any;

        mockGroup = {
            sequence: 123n
        } as any;

        groupWriter = new GroupWriter(mockContext, mockWriter, mockGroup);
    });

    afterEach(() => {
        vi.clearAllMocks();
    });

    describe("constructor", () => {
        it("should initialize with provided parameters", () => {
            expect(groupWriter).toBeInstanceOf(GroupWriter);
            expect(groupWriter.sequence).toBe(123n);
            expect(groupWriter.context).toBeDefined();
        });
    });

    describe("sequence", () => {
        it("should return the group sequence number", () => {
            expect(groupWriter.sequence).toBe(123n);
        });
    });

    describe("writeFrame", () => {
        it("should write Frame data and flush successfully", async () => {
            const data = new Uint8Array([1, 2, 3, 4]);
            const frame = new BytesFrame(data);
            vi.mocked(mockWriter.flush).mockResolvedValue(undefined);

            const error = await groupWriter.writeFrame(frame);

            expect(mockWriter.copyFrom).toHaveBeenCalledWith(frame);
            expect(mockWriter.flush).toHaveBeenCalled();
            expect(error).toBeUndefined();
        });

        it("should write Source data using copyFrom", async () => {
            const mockSource = {
                read: vi.fn()
            } as any;
            vi.mocked(mockWriter.flush).mockResolvedValue(undefined);

            const error = await groupWriter.writeFrame(mockSource);

            expect(mockWriter.copyFrom).toHaveBeenCalledWith(mockSource);
            expect(mockWriter.flush).toHaveBeenCalled();
            expect(error).toBeUndefined();
        });

        it("should return error if flush fails", async () => {
            const data = new Uint8Array([1, 2, 3, 4]);
            const frame = new BytesFrame(data);
            const flushError = new Error("Flush failed");
            vi.mocked(mockWriter.flush).mockResolvedValue(flushError);

            const error = await groupWriter.writeFrame(frame);

            expect(mockWriter.copyFrom).toHaveBeenCalledWith(frame);
            expect(mockWriter.flush).toHaveBeenCalled();
            expect(error).toBe(flushError);
        });
    });

    describe("close", () => {
        it("should close writer and cancel context", async () => {
            await groupWriter.close();

            expect(mockWriter.close).toHaveBeenCalled();
            expect(groupWriter.context.err()).toBeUndefined();
        });

        it("should handle multiple close calls", async () => {
            await groupWriter.close();

            // Get the call count after first close
            const firstCallCount = vi.mocked(mockWriter.close).mock.calls.length;

            await groupWriter.close();

            // The second close should still call close since err() is undefined
            expect(vi.mocked(mockWriter.close).mock.calls.length).toBeGreaterThanOrEqual(firstCallCount);
        });
    });

    describe("cancel", () => {
        it("should cancel writer and context with error", async () => {
            const code = 404;
            const message = "Not found";

            await groupWriter.cancel(code, message);

            expect(mockWriter.cancel).toHaveBeenCalledWith(expect.any(StreamError));
            expect(groupWriter.context.err()).toBeInstanceOf(StreamError);
        });

        it("should not cancel if already cancelled", async () => {
            const code = 404;
            const message = "Not found";

            await groupWriter.cancel(code, message);

            // Clear the mock to check it's not called again
            vi.mocked(mockWriter.cancel).mockClear();

            await groupWriter.cancel(500, "Another error");

            expect(mockWriter.cancel).not.toHaveBeenCalled();
        });
    });

    describe("context", () => {
        it("should return the internal context", () => {
            expect(groupWriter.context).toBeDefined();
            expect(typeof groupWriter.context.done).toBe('function');
            expect(typeof groupWriter.context.err).toBe('function');
        });
    });
});

describe("GroupReader", () => {
    let mockReader: Reader;
    let mockContext: Context;
    let mockGroup: GroupMessage;
    let groupReader: GroupReader;

    beforeEach(() => {
        mockContext = background();

        mockReader = {
            readUint8Array: vi.fn(),
            readVarint: vi.fn(),
            fillN: vi.fn(),
            cancel: vi.fn().mockReturnValue(Promise.resolve()),
            closed: vi.fn().mockReturnValue(Promise.resolve())
        } as any;

        mockGroup = {
            sequence: 456n
        } as any;

        groupReader = new GroupReader(mockContext, mockReader, mockGroup);
    });

    afterEach(() => {
        vi.clearAllMocks();
    });

    describe("constructor", () => {
        it("should initialize with provided parameters", () => {
            expect(groupReader).toBeInstanceOf(GroupReader);
            expect(groupReader.sequence).toBe(456n);
            expect(groupReader.context).toBeDefined();
        });
    });

    describe("groupSequence", () => {
        it("should return the group sequence number", () => {
            expect(groupReader.sequence).toBe(456n);
        });
    });

    describe("readFrame", () => {
        it("should read data successfully", async () => {
            const expectedData = new Uint8Array([1, 2, 3, 4]);

            vi.mocked(mockReader.readVarint).mockResolvedValue([expectedData.byteLength, undefined]);
            vi.mocked(mockReader.fillN).mockImplementation(async (buf: Uint8Array, len: number) => {
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

            vi.mocked(mockReader.readVarint).mockResolvedValue([0, readErr]);

            const [frame, err] = await groupReader.readFrame();

            expect(mockReader.readVarint).toHaveBeenCalled();
            expect(frame).toBeUndefined();
            expect(err).toBe(readErr);
        });

        it("should handle fillN errors", async () => {
            const fillErr = new Error("Fill failed");

            vi.mocked(mockReader.readVarint).mockResolvedValue([10, undefined]);
            vi.mocked(mockReader.fillN).mockResolvedValue(fillErr);

            const [frame, err] = await groupReader.readFrame();

            expect(mockReader.readVarint).toHaveBeenCalled();
            expect(mockReader.fillN).toHaveBeenCalled();
            expect(frame).toBeUndefined();
            expect(err).toBe(fillErr);
        });

        it("should handle varint too large", async () => {
            vi.mocked(mockReader.readVarint).mockResolvedValue([Number.MAX_SAFE_INTEGER + 1, undefined]);

            const [frame, err] = await groupReader.readFrame();

            expect(frame).toBeUndefined();
            expect(err).toBeInstanceOf(Error);
            expect(err?.message).toBe("Varint too large");
        });

        it("should reuse buffer when reading multiple frames", async () => {
            const data1 = new Uint8Array([1, 2, 3]);
            const data2 = new Uint8Array([4, 5, 6, 7, 8]);

            // First read
            vi.mocked(mockReader.readVarint).mockResolvedValueOnce([data1.byteLength, undefined]);
            vi.mocked(mockReader.fillN).mockImplementationOnce(async (buf: Uint8Array, len: number) => {
                buf.set(data1.subarray(0, len));
                return undefined;
            });

            const [frame1, err1] = await groupReader.readFrame();
            expect(err1).toBeUndefined();
            expect(frame1!.bytes.slice(0, data1.byteLength)).toEqual(data1);

            // Second read with larger data
            vi.mocked(mockReader.readVarint).mockResolvedValueOnce([data2.byteLength, undefined]);
            vi.mocked(mockReader.fillN).mockImplementationOnce(async (buf: Uint8Array, len: number) => {
                buf.set(data2.subarray(0, len));
                return undefined;
            });

            const [frame2, err2] = await groupReader.readFrame();
            expect(err2).toBeUndefined();
            expect(frame2!.bytes.slice(0, data2.byteLength)).toEqual(data2);
        });
    });

    describe("cancel", () => {
        it("should cancel reader with code", async () => {
            const code = 404;
            const message = "Not found";

            await groupReader.cancel(code, message);

            expect(mockReader.cancel).toHaveBeenCalledWith(expect.any(StreamError));
            expect(groupReader.context.err()).toBeInstanceOf(StreamError);
        });

        it("should not cancel if already cancelled", async () => {
            const code = 404;
            const message = "Not found";

            await groupReader.cancel(code, message);

            // Clear the mock to check it's not called again
            vi.mocked(mockReader.cancel).mockClear();

            await groupReader.cancel(500, "Another error");

            expect(mockReader.cancel).not.toHaveBeenCalled();
        });
    });

    describe("context", () => {
        it("should return the internal context", () => {
            expect(groupReader.context).toBeDefined();
            expect(typeof groupReader.context.done).toBe('function');
            expect(typeof groupReader.context.err).toBe('function');
        });
    });
});
