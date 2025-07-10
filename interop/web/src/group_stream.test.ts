import { GroupWriter, GroupReader } from "./group_stream";
import { CancelCauseFunc, Context, withCancelCause, background } from "./internal/context";
import { Reader, Writer } from "./io";
import { StreamError } from "./io/error";
import { GroupMessage } from "./message";

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
            flush: jest.fn().mockResolvedValue(undefined),
            close: jest.fn(),
            cancel: jest.fn()
        } as any;
        
        mockContext = {
            done: jest.fn().mockResolvedValue(undefined),
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
            expect(mockContext.done).toHaveBeenCalled();
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
            mockWriter.flush.mockResolvedValue(undefined);
            
            const error = await groupWriter.write(data);
            
            expect(mockWriter.writeUint8Array).toHaveBeenCalledWith(data);
            expect(mockWriter.flush).toHaveBeenCalled();
            expect(error).toBeUndefined();
        });
        
        it("should return error if flush fails", async () => {
            const data = new Uint8Array([1, 2, 3, 4]);
            const flushError = new Error("Flush failed");
            mockWriter.flush.mockResolvedValue(flushError);
            
            const error = await groupWriter.write(data);
            
            expect(mockWriter.writeUint8Array).toHaveBeenCalledWith(data);
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
            
            expect(StreamError).toHaveBeenCalledWith(code, message);
            expect(mockWriter.cancel).toHaveBeenCalled();
            expect(mockCancelFunc).toHaveBeenCalled();
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
            cancel: jest.fn()
        } as any;
        
        mockContext = {
            done: jest.fn().mockResolvedValue(undefined),
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
            mockReader.readUint8Array.mockResolvedValue([expectedData, undefined]);
            
            const [data, error] = await groupReader.read();
            
            expect(mockReader.readUint8Array).toHaveBeenCalled();
            expect(data).toBe(expectedData);
            expect(error).toBeUndefined();
        });
        
        it("should handle read errors", async () => {
            const readError = new Error("Read failed");
            mockReader.readUint8Array.mockResolvedValue([undefined, readError]);
            
            const [data, error] = await groupReader.read();
            
            expect(mockReader.readUint8Array).toHaveBeenCalled();
            expect(data).toBeUndefined();
            expect(error).toBe(readError);
        });
    });
    
    describe("cancel", () => {
        it("should cancel reader with code", () => {
            const code = 404;
            
            groupReader.cancel(code);
            
            expect(mockReader.cancel).toHaveBeenCalledWith(code, "cancelled");
            expect(mockCancelFunc).toHaveBeenCalledWith(expect.any(Error));
        });
    });
});
