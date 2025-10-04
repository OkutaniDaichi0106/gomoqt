import { describe, test, expect, beforeEach, afterEach, vi } from 'vitest';
import type { MockedFunction } from 'vitest';

// Mock the external dependencies before importing the module under test
vi.mock("@okutanidaichi/moqt/io", () => ({
    writeVarint: vi.fn((dest: Uint8Array, value: number) => {
        // Simple mock implementation - write the number as a single byte
        // In real implementation this would write a varint
        dest[0] = value & 0xFF;
        return 1;
    }),
    varintLen: vi.fn((value: number) => {
        // Simple mock - all varints are 1 byte for testing
        return 1;
    }),
    readVarint: vi.fn().mockReturnValue([BigInt(0), 0]),
}));

import { EncodedContainer } from "./container";
import type { EncodedChunk } from "./container";
import { writeVarint, varintLen } from "@okutanidaichi/moqt/io";

// Type-safe mock functions
const mockWriteVarint = writeVarint as MockedFunction<typeof writeVarint>;
const mockVarintLen = varintLen as MockedFunction<typeof varintLen>;

// Test constants
const SMALL_BUFFER_SIZE = 4;
const MEDIUM_BUFFER_SIZE = 10;
const LARGE_CHUNK_SIZE = 10000;
const STRESS_TEST_ITERATIONS = 1000;

// Mock implementation of EncodedChunk for testing
class MockEncodedChunk implements EncodedChunk {
    type: "key" | "delta";
    byteLength: number;
    timestamp?: number;
    #data: Uint8Array;

    constructor(data: Uint8Array, type: "key" | "delta" = "key", timestamp?: number) {
        this.#data = data;
        this.type = type;
        this.byteLength = data.byteLength;
        this.timestamp = timestamp;
    }

    copyTo(dest: AllowSharedBufferSource): void {
        let view: Uint8Array;
        
        if (dest instanceof Uint8Array) {
            view = dest;
        } else if (dest instanceof ArrayBuffer) {
            view = new Uint8Array(dest);
        } else if (ArrayBuffer.isView(dest)) {
            const v = dest as ArrayBufferView;
            view = new Uint8Array(v.buffer, v.byteOffset, v.byteLength);
        } else {
            throw new Error("Unsupported destination type");
        }

        if (view.length < this.#data.length) {
            throw new Error("Destination buffer is too small");
        }

        view.set(this.#data);
    }
}

describe("EncodedContainer", () => {
    beforeEach(() => {
        // Reset mocks before each test to ensure isolation
        vi.clearAllMocks();
        
        // Set up default mock behavior
        mockWriteVarint.mockImplementation((dest: Uint8Array, value: number) => {
            dest[0] = value & 0xFF;
            return 1;
        });
        mockVarintLen.mockImplementation((value: number | bigint) => 1);
    });

    describe("constructor", () => {
        test("creates container with chunk", () => {
            const mockChunk = new MockEncodedChunk(new Uint8Array([1, 2, 3]));
            const container = new EncodedContainer(mockChunk);

            expect(container).toBeDefined();
            expect(container.chunk).toBe(mockChunk);
        });

        test("creates container with key chunk", () => {
            const mockChunk = new MockEncodedChunk(new Uint8Array([1, 2, 3]), "key");
            const container = new EncodedContainer(mockChunk);

            expect(container.chunk.type).toBe("key");
        });

        test("creates container with delta chunk", () => {
            const mockChunk = new MockEncodedChunk(new Uint8Array([1, 2, 3]), "delta");
            const container = new EncodedContainer(mockChunk);

            expect(container.chunk.type).toBe("delta");
        });

        test("creates container with timestamp", () => {
            const timestamp = Date.now();
            const mockChunk = new MockEncodedChunk(new Uint8Array([1, 2, 3]), "key", timestamp);
            const container = new EncodedContainer(mockChunk);

            expect(container.chunk.timestamp).toBe(timestamp);
        });
    });

    describe("byteLength", () => {
        test("returns chunk byte length", () => {
            const data = new Uint8Array([1, 2, 3, 4, 5]);
            const mockChunk = new MockEncodedChunk(data);
            const container = new EncodedContainer(mockChunk);

            expect(container.byteLength).toBe(5);
        });

        test("returns zero for empty chunk", () => {
            const mockChunk = new MockEncodedChunk(new Uint8Array(0));
            const container = new EncodedContainer(mockChunk);

            expect(container.byteLength).toBe(0);
        });

        test("returns correct length for large chunk", () => {
            const data = new Uint8Array(1000);
            const mockChunk = new MockEncodedChunk(data);
            const container = new EncodedContainer(mockChunk);

            expect(container.byteLength).toBe(1000);
        });
    });

    describe("copyTo", () => {
        test("copies to Uint8Array destination with correct format", () => {
            const sourceData = new Uint8Array([1, 2, 3]);
            const mockChunk = new MockEncodedChunk(sourceData);
            const container = new EncodedContainer(mockChunk);

            const dest = new Uint8Array(MEDIUM_BUFFER_SIZE);
            container.copyTo(dest);

            // Verify varint encoding is called with correct length
            expect(mockWriteVarint).toHaveBeenCalledWith(dest, 3);
            expect(mockVarintLen).toHaveBeenCalledWith(3);
            
            // Verify actual data is written after varint (starting at index 1)
            expect(dest.subarray(1, 4)).toEqual(sourceData);
        });

        test("copies to ArrayBuffer destination", () => {
            const sourceData = new Uint8Array([1, 2, 3]);
            const mockChunk = new MockEncodedChunk(sourceData);
            const container = new EncodedContainer(mockChunk);

            const dest = new ArrayBuffer(10);
            container.copyTo(dest);

            expect(writeVarint).toHaveBeenCalled();
            expect(varintLen).toHaveBeenCalled();
        });

        test("copies to DataView destination", () => {
            const sourceData = new Uint8Array([1, 2, 3]);
            const mockChunk = new MockEncodedChunk(sourceData);
            const container = new EncodedContainer(mockChunk);

            const buffer = new ArrayBuffer(10);
            const dest = new DataView(buffer);
            container.copyTo(dest);

            expect(writeVarint).toHaveBeenCalled();
            expect(varintLen).toHaveBeenCalled();
        });

        test("copies to Int8Array destination", () => {
            const sourceData = new Uint8Array([1, 2, 3]);
            const mockChunk = new MockEncodedChunk(sourceData);
            const container = new EncodedContainer(mockChunk);

            const dest = new Int8Array(10);
            container.copyTo(dest);

            expect(writeVarint).toHaveBeenCalled();
            expect(varintLen).toHaveBeenCalled();
        });

        test("throws error for too small destination buffer", () => {
            const sourceData = new Uint8Array([1, 2, 3, 4, 5]);
            const mockChunk = new MockEncodedChunk(sourceData);
            const container = new EncodedContainer(mockChunk);

            // Mock varintLen to return 1, so total required is 6 bytes (1 + 5)
            mockVarintLen.mockReturnValue(1);

            const dest = new Uint8Array(SMALL_BUFFER_SIZE); // Too small (4 bytes < 6 bytes needed)
            
            expect(() => container.copyTo(dest)).toThrow("Destination buffer is too small");
        });

        test("throws error for unsupported destination type", () => {
            const sourceData = new Uint8Array([1, 2, 3]);
            const mockChunk = new MockEncodedChunk(sourceData);
            const container = new EncodedContainer(mockChunk);

            const invalidDest = "not a buffer" as any;
            
            expect(() => container.copyTo(invalidDest)).toThrow("Unsupported destination type");
        });

        test("handles exact size destination buffer", () => {
            const sourceData = new Uint8Array([1, 2, 3]);
            const mockChunk = new MockEncodedChunk(sourceData);
            const container = new EncodedContainer(mockChunk);

            // Mock varintLen to return 1, so total required is 4 bytes (1 + 3)
            mockVarintLen.mockReturnValue(1);

            const dest = new Uint8Array(SMALL_BUFFER_SIZE); // Exact size (4 bytes)
            
            expect(() => container.copyTo(dest)).not.toThrow();
        });

        test("calls chunk.copyTo with correct subarray", () => {
            const sourceData = new Uint8Array([1, 2, 3]);
            const mockChunk = new MockEncodedChunk(sourceData);
            const copyToSpy = vi.spyOn(mockChunk, 'copyTo');
            const container = new EncodedContainer(mockChunk);

            // Mock varintLen to return 2 bytes for the length prefix
            mockVarintLen.mockReturnValue(2);

            const dest = new Uint8Array(MEDIUM_BUFFER_SIZE);
            container.copyTo(dest);

            // Verify chunk data is written after the 2-byte varint prefix
            expect(copyToSpy).toHaveBeenCalledWith(dest.subarray(2));
        });
    });

    describe("SharedArrayBuffer support", () => {
        test("handles SharedArrayBuffer when available", () => {
            // SharedArrayBuffer is used in multi-threaded scenarios
            if (typeof SharedArrayBuffer !== "undefined") {
                const sourceData = new Uint8Array([1, 2, 3]);
                const mockChunk = new MockEncodedChunk(sourceData);
                const container = new EncodedContainer(mockChunk);

                const dest = new SharedArrayBuffer(MEDIUM_BUFFER_SIZE);
                
                expect(() => container.copyTo(dest)).not.toThrow();
                expect(mockWriteVarint).toHaveBeenCalled();
            } else {
                expect(true).toBe(true);
            }
        });

        test("handles SharedArrayBuffer with exact size", () => {
            if (typeof SharedArrayBuffer !== "undefined") {
                const sourceData = new Uint8Array([1, 2, 3]);
                const mockChunk = new MockEncodedChunk(sourceData);
                const container = new EncodedContainer(mockChunk);

                mockVarintLen.mockReturnValue(1);
                // Exact size: 1 byte varint + 3 bytes data = 4 bytes
                const dest = new SharedArrayBuffer(SMALL_BUFFER_SIZE);
                
                expect(() => container.copyTo(dest)).not.toThrow();
            } else {
                expect(true).toBe(true);
            }
        });
    });

    describe("Error Handling", () => {
        test("handles chunk with zero length", () => {
            const mockChunk = new MockEncodedChunk(new Uint8Array(0));
            const container = new EncodedContainer(mockChunk);

            expect(container.byteLength).toBe(0);

            const dest = new Uint8Array(MEDIUM_BUFFER_SIZE);
            expect(() => container.copyTo(dest)).not.toThrow();
        });

        test("handles chunk copyTo throwing error", () => {
            const mockChunk = new MockEncodedChunk(new Uint8Array([1, 2, 3]));
            vi.spyOn(mockChunk, 'copyTo').mockImplementation(() => {
                throw new Error("Chunk copy error");
            });
            const container = new EncodedContainer(mockChunk);

            const dest = new Uint8Array(MEDIUM_BUFFER_SIZE);
            
            expect(() => container.copyTo(dest)).toThrow("Chunk copy error");
        });

        test("handles writeVarint throwing error", () => {
            const mockChunk = new MockEncodedChunk(new Uint8Array([1, 2, 3]));
            const container = new EncodedContainer(mockChunk);

            // Simulate writeVarint failure (e.g., invalid buffer state)
            mockWriteVarint.mockImplementationOnce(() => {
                throw new Error("Varint write error");
            });

            const dest = new Uint8Array(MEDIUM_BUFFER_SIZE);
            
            expect(() => container.copyTo(dest)).toThrow("Varint write error");
        });

        test("handles destination without byteLength property", () => {
            const mockChunk = new MockEncodedChunk(new Uint8Array([1, 2, 3]));
            const container = new EncodedContainer(mockChunk);

            // Create object that looks like a buffer but has no byteLength
            const fakeBuffer = new Uint8Array(MEDIUM_BUFFER_SIZE);
            Object.defineProperty(fakeBuffer, 'byteLength', {
                value: undefined,
                configurable: true
            });

            // Should still work - code handles missing byteLength
            expect(() => container.copyTo(fakeBuffer)).not.toThrow();
        });
    });

    describe("Boundary Value Tests", () => {
        test("handles maximum safe integer byte length", () => {
            // Test theoretical maximum to ensure no overflow issues
            const mockChunk = {
                type: "key" as const,
                byteLength: Number.MAX_SAFE_INTEGER,
                copyTo: vi.fn()
            };
            const container = new EncodedContainer(mockChunk);

            expect(container.byteLength).toBe(Number.MAX_SAFE_INTEGER);
            
            // varintLen is only called during copyTo operation
            const largeBuffer = new Uint8Array(MEDIUM_BUFFER_SIZE);
            // This will fail size check, but varintLen should be called first
            try {
                container.copyTo(largeBuffer);
            } catch {
                // Expected to throw due to buffer size
            }
            expect(mockVarintLen).toHaveBeenCalledWith(Number.MAX_SAFE_INTEGER);
        });

        test("handles single byte chunk", () => {
            const mockChunk = new MockEncodedChunk(new Uint8Array([42]));
            const container = new EncodedContainer(mockChunk);

            expect(container.byteLength).toBe(1);
            
            const dest = new Uint8Array(MEDIUM_BUFFER_SIZE);
            expect(() => container.copyTo(dest)).not.toThrow();
            
            // Verify the single byte is written correctly
            expect(dest[1]).toBe(42);
        });

        test("handles large chunk efficiently", () => {
            // Test with 10KB chunk (typical for video frames)
            const largeData = new Uint8Array(LARGE_CHUNK_SIZE).fill(42);
            const mockChunk = new MockEncodedChunk(largeData);
            const container = new EncodedContainer(mockChunk);

            expect(container.byteLength).toBe(LARGE_CHUNK_SIZE);
            
            const dest = new Uint8Array(LARGE_CHUNK_SIZE + MEDIUM_BUFFER_SIZE);
            expect(() => container.copyTo(dest)).not.toThrow();
        });

        test("handles chunk without timestamp", () => {
            const mockChunk = new MockEncodedChunk(new Uint8Array([1, 2, 3]));
            const container = new EncodedContainer(mockChunk);

            expect(container.chunk.timestamp).toBeUndefined();
        });

        test("handles chunk with zero timestamp", () => {
            // Zero timestamp is valid (e.g., start of stream)
            const mockChunk = new MockEncodedChunk(new Uint8Array([1, 2, 3]), "key", 0);
            const container = new EncodedContainer(mockChunk);

            expect(container.chunk.timestamp).toBe(0);
            expect(container.chunk.type).toBe("key");
        });

        test("handles varint length of zero", () => {
            const mockChunk = new MockEncodedChunk(new Uint8Array([1, 2, 3]));
            const container = new EncodedContainer(mockChunk);

            // Edge case: varintLen returns 0 (shouldn't happen but test defensive coding)
            mockVarintLen.mockReturnValue(0);

            const dest = new Uint8Array(SMALL_BUFFER_SIZE);
            // With 0-byte varint, only need 3 bytes for data
            expect(() => container.copyTo(dest)).not.toThrow();
        });
    });

    describe("Integration with varint functions", () => {
        test("calls varintLen with correct chunk length", () => {
            const sourceData = new Uint8Array([1, 2, 3, 4, 5]);
            const mockChunk = new MockEncodedChunk(sourceData);
            const container = new EncodedContainer(mockChunk);

            const dest = new Uint8Array(10);
            container.copyTo(dest);

            expect(varintLen).toHaveBeenCalledWith(5);
        });

        test("calls writeVarint with correct parameters", () => {
            const sourceData = new Uint8Array([1, 2, 3]);
            const mockChunk = new MockEncodedChunk(sourceData);
            const container = new EncodedContainer(mockChunk);

            const dest = new Uint8Array(10);
            container.copyTo(dest);

            expect(writeVarint).toHaveBeenCalledWith(dest, 3);
        });

        test("respects varint length in buffer calculation", () => {
            const sourceData = new Uint8Array([1, 2, 3]);
            const mockChunk = new MockEncodedChunk(sourceData);
            const container = new EncodedContainer(mockChunk);

            // Mock varintLen to return 3 bytes (simulating larger value encoding)
            mockVarintLen.mockReturnValue(3);

            // Should need 6 bytes total (3 for varint + 3 for data)
            const dest = new Uint8Array(5); // Too small (5 < 6)
            
            expect(() => container.copyTo(dest)).toThrow("Destination buffer is too small");
        });
    });
});

describe("EncodedChunk Interface", () => {
    test("MockEncodedChunk implements interface correctly", () => {
        const data = new Uint8Array([1, 2, 3]);
        const chunk = new MockEncodedChunk(data, "key", 12345);

        expect(chunk.type).toBe("key");
        expect(chunk.byteLength).toBe(3);
        expect(chunk.timestamp).toBe(12345);
        expect(typeof chunk.copyTo).toBe("function");
    });

    test("chunk copyTo works correctly", () => {
        const sourceData = new Uint8Array([1, 2, 3, 4, 5]);
        const chunk = new MockEncodedChunk(sourceData);

        const dest = new Uint8Array(10);
        chunk.copyTo(dest);

        expect(dest.subarray(0, 5)).toEqual(sourceData);
    });

    test("chunk copyTo throws for too small destination", () => {
        const sourceData = new Uint8Array([1, 2, 3, 4, 5]);
        const chunk = new MockEncodedChunk(sourceData);

        const dest = new Uint8Array(3); // Too small
        
        expect(() => chunk.copyTo(dest)).toThrow("Destination buffer is too small");
    });
});

describe("Stress and Stability Tests", () => {
    test("handles multiple containers with different sizes", () => {
        // Create containers with varying sizes to test consistency
        const sizes = [1, 10, 100, 1000];
        const containers = sizes.map(size => {
            const data = new Uint8Array(size);
            const chunk = new MockEncodedChunk(data);
            return new EncodedContainer(chunk);
        });

        containers.forEach((container, i) => {
            expect(container.byteLength).toBe(sizes[i]);
        });
    });

    test("handles large chunk data correctly", () => {
        // Test with realistic large chunk size (10KB)
        const data = new Uint8Array(LARGE_CHUNK_SIZE).fill(42);
        const chunk = new MockEncodedChunk(data);
        const container = new EncodedContainer(chunk);

        expect(container.byteLength).toBe(LARGE_CHUNK_SIZE);
        
        const dest = new Uint8Array(LARGE_CHUNK_SIZE + MEDIUM_BUFFER_SIZE);
        expect(() => container.copyTo(dest)).not.toThrow();
    });

    test("handles repeated copyTo calls to different buffers", () => {
        const data = new Uint8Array([1, 2, 3, 4, 5]);
        const chunk = new MockEncodedChunk(data);
        const container = new EncodedContainer(chunk);

        vi.clearAllMocks();

        // Simulate repeated encoding operations (e.g., retransmissions)
        for (let i = 0; i < STRESS_TEST_ITERATIONS; i++) {
            const dest = new Uint8Array(20);
            container.copyTo(dest);
        }

        expect(mockWriteVarint).toHaveBeenCalledTimes(STRESS_TEST_ITERATIONS);
    });
});

describe("Type Safety and Interface Compliance", () => {
    test("EncodedContainer implements Source interface", () => {
        const chunk = new MockEncodedChunk(new Uint8Array([1, 2, 3]));
        const container = new EncodedContainer(chunk);

        // Source interface methods
        expect(typeof container.byteLength).toBe("number");
        expect(typeof container.copyTo).toBe("function");
    });

    test("container maintains chunk reference integrity", () => {
        const chunk = new MockEncodedChunk(new Uint8Array([1, 2, 3]));
        const container = new EncodedContainer(chunk);

        expect(container.chunk).toBe(chunk);
        expect(Object.is(container.chunk, chunk)).toBe(true);
    });

    test("chunk type is preserved", () => {
        const keyChunk = new MockEncodedChunk(new Uint8Array([1]), "key");
        const deltaChunk = new MockEncodedChunk(new Uint8Array([2]), "delta");
        
        const keyContainer = new EncodedContainer(keyChunk);
        const deltaContainer = new EncodedContainer(deltaChunk);

        expect(keyContainer.chunk.type).toBe("key");
        expect(deltaContainer.chunk.type).toBe("delta");
    });

    test("timestamp is properly handled", () => {
        const now = Date.now();
        const chunk = new MockEncodedChunk(new Uint8Array([1]), "key", now);
        const container = new EncodedContainer(chunk);

        expect(container.chunk.timestamp).toBe(now);
        expect(typeof container.chunk.timestamp).toBe("number");
    });
});

describe("Buffer View Compatibility", () => {
    test("works with different TypedArray views", () => {
        const data = new Uint8Array([1, 2, 3]);
        const chunk = new MockEncodedChunk(data);
        const container = new EncodedContainer(chunk);

        // Use buffer sizes that are compatible with all view types
        const buffer = new ArrayBuffer(24); // Multiple of 8 for Float64Array
        const views = [
            new Uint8Array(buffer),
            new Int8Array(buffer),
            new Uint16Array(buffer),
            new Int16Array(buffer),
            new Uint32Array(buffer),
            new Int32Array(buffer),
            new Float32Array(buffer),
            new Float64Array(buffer)
        ];

        views.forEach(view => {
            expect(() => container.copyTo(view)).not.toThrow();
        });
    });

    test("handles DataView with offset and length", () => {
        const data = new Uint8Array([1, 2, 3]);
        const chunk = new MockEncodedChunk(data);
        const container = new EncodedContainer(chunk);

        const buffer = new ArrayBuffer(20);
        const view = new DataView(buffer, 5, 10); // offset=5, length=10

        expect(() => container.copyTo(view)).not.toThrow();
    });

    test("handles Uint8Array with offset", () => {
        const data = new Uint8Array([1, 2, 3]);
        const chunk = new MockEncodedChunk(data);
        const container = new EncodedContainer(chunk);

        const buffer = new ArrayBuffer(20);
        const view = new Uint8Array(buffer, 3, 10); // offset=3, length=10

        expect(() => container.copyTo(view)).not.toThrow();
    });
});

describe("Error Message Quality", () => {
    test("provides clear error for buffer size", () => {
        const data = new Uint8Array([1, 2, 3, 4, 5]);
        const chunk = new MockEncodedChunk(data);
        const container = new EncodedContainer(chunk);

        const smallDest = new Uint8Array(3);
        
        expect(() => container.copyTo(smallDest)).toThrow("Destination buffer is too small");
    });

    test("provides clear error for unsupported type", () => {
        const data = new Uint8Array([1, 2, 3]);
        const chunk = new MockEncodedChunk(data);
        const container = new EncodedContainer(chunk);

        const invalidDest = { notABuffer: true } as any;
        
        expect(() => container.copyTo(invalidDest)).toThrow("Unsupported destination type");
    });
});

describe("Edge Cases and Robustness", () => {
    test("handles chunk with undefined optional properties", () => {
        const chunk = {
            type: "key" as const,
            byteLength: 3,
            copyTo: vi.fn()
        };
        const container = new EncodedContainer(chunk);

        expect(container.chunk.timestamp).toBeUndefined();
        expect(container.byteLength).toBe(3);
    });

    test("handles negative timestamp", () => {
        const chunk = new MockEncodedChunk(new Uint8Array([1]), "key", -1);
        const container = new EncodedContainer(chunk);

        expect(container.chunk.timestamp).toBe(-1);
    });

    test("handles fractional timestamp", () => {
        const chunk = new MockEncodedChunk(new Uint8Array([1]), "key", 123.45);
        const container = new EncodedContainer(chunk);

        expect(container.chunk.timestamp).toBe(123.45);
    });

    test("container creation is idempotent", () => {
        const chunk = new MockEncodedChunk(new Uint8Array([1, 2, 3]));
        const container1 = new EncodedContainer(chunk);
        const container2 = new EncodedContainer(chunk);

        expect(container1.chunk).toBe(container2.chunk);
        expect(container1.byteLength).toBe(container2.byteLength);
    });

    test("handles concurrent access to same chunk", () => {
        const chunk = new MockEncodedChunk(new Uint8Array([1, 2, 3]));
        const container1 = new EncodedContainer(chunk);
        const container2 = new EncodedContainer(chunk);

        const dest1 = new Uint8Array(10);
        const dest2 = new Uint8Array(10);

        expect(() => {
            container1.copyTo(dest1);
            container2.copyTo(dest2);
        }).not.toThrow();
    });
});
