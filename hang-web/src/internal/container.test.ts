import { describe, test, expect, beforeEach, afterEach, jest } from 'vitest';

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
        // Reset mocks before each test
        vi.clearAllMocks();
        (writeVarint as vi.mock).mockImplementation((dest: Uint8Array, value: number) => {
            dest[0] = value & 0xFF;
            return 1;
        });
        (varintLen as vi.mock).mockImplementation((value: number) => 1);
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
        test("copies to Uint8Array destination", () => {
            const sourceData = new Uint8Array([1, 2, 3]);
            const mockChunk = new MockEncodedChunk(sourceData);
            const container = new EncodedContainer(mockChunk);

            const dest = new Uint8Array(10);
            container.copyTo(dest);

            expect(writeVarint).toHaveBeenCalledWith(dest, 3);
            expect(varintLen).toHaveBeenCalledWith(3);
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

            // Mock varintLen to return 1, so total required is 6 bytes
            (varintLen as vi.mock).mockReturnValue(1);

            const dest = new Uint8Array(4); // Too small
            
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

            // Mock varintLen to return 1, so total required is 4 bytes
            (varintLen as vi.mock).mockReturnValue(1);

            const dest = new Uint8Array(4); // Exact size
            
            expect(() => container.copyTo(dest)).not.toThrow();
        });

        test("calls chunk.copyTo with correct subarray", () => {
            const sourceData = new Uint8Array([1, 2, 3]);
            const mockChunk = new MockEncodedChunk(sourceData);
            const copyToSpy = vi.spyOn(mockChunk, 'copyTo');
            const container = new EncodedContainer(mockChunk);

            // Mock varintLen to return 2
            (varintLen as vi.mock).mockReturnValue(2);

            const dest = new Uint8Array(10);
            container.copyTo(dest);

            expect(copyToSpy).toHaveBeenCalledWith(dest.subarray(2));
        });
    });

    describe("SharedArrayBuffer support", () => {
        test("handles SharedArrayBuffer when available", () => {
            // Only test if SharedArrayBuffer is supported
            if (typeof SharedArrayBuffer !== "undefined") {
                const sourceData = new Uint8Array([1, 2, 3]);
                const mockChunk = new MockEncodedChunk(sourceData);
                const container = new EncodedContainer(mockChunk);

                const dest = new SharedArrayBuffer(10);
                
                expect(() => container.copyTo(dest)).not.toThrow();
            } else {
                // Skip test if SharedArrayBuffer is not available
                expect(true).toBe(true);
            }
        });
    });

    describe("Error Handling", () => {
        test("handles chunk with zero length", () => {
            const mockChunk = new MockEncodedChunk(new Uint8Array(0));
            const container = new EncodedContainer(mockChunk);

            const dest = new Uint8Array(10);
            
            expect(() => container.copyTo(dest)).not.toThrow();
        });

        test("handles chunk copyTo throwing error", () => {
            const mockChunk = new MockEncodedChunk(new Uint8Array([1, 2, 3]));
            vi.spyOn(mockChunk, 'copyTo').mockImplementation(() => {
                throw new Error("Chunk copy error");
            });
            const container = new EncodedContainer(mockChunk);

            const dest = new Uint8Array(10);
            
            expect(() => container.copyTo(dest)).toThrow("Chunk copy error");
        });

        test("handles writeVarint throwing error", () => {
            const mockChunk = new MockEncodedChunk(new Uint8Array([1, 2, 3]));
            const container = new EncodedContainer(mockChunk);

            (writeVarint as vi.mock).mockImplementationOnce(() => {
                throw new Error("Varint write error");
            });

            const dest = new Uint8Array(10);
            
            expect(() => container.copyTo(dest)).toThrow("Varint write error");
        });
    });

    describe("Boundary Value Tests", () => {
        test("handles maximum safe integer byte length", () => {
            const mockChunk = {
                type: "key" as const,
                byteLength: Number.MAX_SAFE_INTEGER,
                copyTo: vi.fn()
            };
            const container = new EncodedContainer(mockChunk);

            expect(container.byteLength).toBe(Number.MAX_SAFE_INTEGER);
        });

        test("handles single byte chunk", () => {
            const mockChunk = new MockEncodedChunk(new Uint8Array([42]));
            const container = new EncodedContainer(mockChunk);

            expect(container.byteLength).toBe(1);
            
            const dest = new Uint8Array(10);
            expect(() => container.copyTo(dest)).not.toThrow();
        });

        test("handles large chunk", () => {
            // Create a larger chunk to test memory handling
            const largeData = new Uint8Array(10000);
            largeData.fill(42);
            const mockChunk = new MockEncodedChunk(largeData);
            const container = new EncodedContainer(mockChunk);

            expect(container.byteLength).toBe(10000);
        });

        test("handles chunk without timestamp", () => {
            const mockChunk = new MockEncodedChunk(new Uint8Array([1, 2, 3]));
            const container = new EncodedContainer(mockChunk);

            expect(container.chunk.timestamp).toBeUndefined();
        });

        test("handles chunk with zero timestamp", () => {
            const mockChunk = new MockEncodedChunk(new Uint8Array([1, 2, 3]), "key", 0);
            const container = new EncodedContainer(mockChunk);

            expect(container.chunk.timestamp).toBe(0);
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

            // Mock varintLen to return 3
            (varintLen as vi.mock).mockReturnValue(3);

            // Should need 6 bytes total (3 for data + 3 for varint)
            const dest = new Uint8Array(5); // Too small
            
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

describe("Performance and Memory Tests", () => {
    test("handles multiple containers efficiently", () => {
        const containers = Array(100).fill(0).map((_, i) => {
            const data = new Uint8Array([i % 256]);
            const chunk = new MockEncodedChunk(data);
            return new EncodedContainer(chunk);
        });

        containers.forEach((container, i) => {
            expect(container.byteLength).toBe(1);
        });
    });

    test("memory usage is predictable for large chunks", () => {
        const size = 50000;
        const data = new Uint8Array(size);
        const chunk = new MockEncodedChunk(data);
        const container = new EncodedContainer(chunk);

        expect(container.byteLength).toBe(size);
        
        // Should be able to create destination buffer
        const dest = new Uint8Array(size + 10);
        expect(() => container.copyTo(dest)).not.toThrow();
    });

    test("handles rapid successive copyTo operations", () => {
        const data = new Uint8Array([1, 2, 3, 4, 5]);
        const chunk = new MockEncodedChunk(data);
        const container = new EncodedContainer(chunk);

        // Clear previous calls before counting
        vi.clearAllMocks();

        // Perform many copy operations
        for (let i = 0; i < 1000; i++) {
            const dest = new Uint8Array(20);
            container.copyTo(dest);
        }

        expect(writeVarint).toHaveBeenCalledTimes(1000);
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
