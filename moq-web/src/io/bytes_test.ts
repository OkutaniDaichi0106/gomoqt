import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows } from "../../deps.ts";
import { BytesBuffer, writeVarint, writeBigVarint, writeUint8Array, writeString, readVarint, readBigVarint, readUint8Array, readString } from './bytes.ts';

describe('io bytes', () => {
  it('varint roundtrip small values', () => {
    const buf = new Uint8Array(8);
    const len = writeVarint(buf, 42, 0);
    const [v, n] = readVarint(buf, 0);
    assertEquals(n, len);
    assertEquals(v, 42);
  });

  it('varint 2-byte roundtrip', () => {
    const buf = new Uint8Array(8);
    const len = writeVarint(buf, 0x123, 0);
    const [v, n] = readVarint(buf, 0);
    assertEquals(n, len);
    assertEquals(v, 0x123);
  });

  it('write/read bytes roundtrip', () => {
    const data = new Uint8Array([1,2,3,4,5]);
    const buf = new Uint8Array(16);
    const wrote = writeUint8Array(buf, data, 0);
    const [out, n] = readUint8Array(buf, 0);
    assertEquals(n, wrote);
    expect(Array.from(out)).toEqual(Array.from(data));
  });

  it('write/read string roundtrip', () => {
    const s = 'hello こんにちは';
    const buf = new Uint8Array(64);
    const w = writeString(buf, s, 0);
    const [out, n] = readString(buf, 0);
    assertEquals(n, w);
    assertEquals(out, s);
  });
});

describe('BytesBuffer', () => {
    it('should write and read data', () => {
        const buffer = new BytesBuffer(new ArrayBuffer(1024));
        const data = new Uint8Array([1, 2, 3]);
        buffer.write(data);
        assertEquals(buffer.size, 3);
        const readBuf = new Uint8Array(3);
        const bytesRead = buffer.read(readBuf);
        assertEquals(bytesRead, 3);
        assertEquals(readBuf, data);
        assertEquals(buffer.size, 0);
    });

    it('should grow when capacity is exceeded', () => {
        const buffer = new BytesBuffer(new ArrayBuffer(2));
        const data = new Uint8Array([1, 2]);
        buffer.write(data);
        expect(buffer.capacity).toBeGreaterThanOrEqual(2);
        const moreData = new Uint8Array([3, 4, 5]);
        buffer.write(moreData);
        expect(buffer.capacity).toBeGreaterThanOrEqual(5);
    });

    describe('readUint8', () => {
        it('should read a single byte', () => {
            const buffer = new BytesBuffer(new ArrayBuffer(10));
            const data = new Uint8Array([1, 2, 3]);
            buffer.write(data);
            assertEquals(buffer.size, 3);
            const byte = buffer.readUint8();
            assertEquals(byte, 1);
            assertEquals(buffer.size, 2);
        });
    });

    describe('writeUint8', () => {
        it('should write a single byte', () => {
            const buffer = new BytesBuffer(new ArrayBuffer(10));
            buffer.writeUint8(42);
            buffer.writeUint8(43);
            buffer.writeUint8(44);
            buffer.writeUint8(45);
            buffer.writeUint8(46);
            assertEquals(buffer.size, 5);
        });
    });

    describe('reserve', () => {
        it('should return a writable buffer with sufficient capacity', () => {
            const buffer = new BytesBuffer(new ArrayBuffer(10));
            const data1 = new Uint8Array([1, 2, 3]);
            buffer.write(data1);
            assertEquals(buffer.size, 3);
            
            const reservedBuffer = buffer.reserve(6);
            expect(reservedBuffer.length).toBeGreaterThanOrEqual(6);
            expect(buffer.capacity).toBeGreaterThanOrEqual(9);
        });
    });

    describe('construction with initial data', () => {
        it('should initialize with existing data', () => {
            const initialData = new Uint8Array([1, 2, 3]);
            const buffer = new BytesBuffer(initialData.buffer);
            // Initially buffer is empty for writing, data needs to be written first
            assertEquals(buffer.size, 0);
            expect(buffer.capacity).toBeGreaterThanOrEqual(3);
        });
    });

    describe('bytes method', () => {
        it('should return current buffer content without consuming it', () => {
            const buffer = new BytesBuffer(new ArrayBuffer(10));
            const data = new Uint8Array([1, 2, 3, 4, 5]);
            buffer.write(data);
            
            const currentBytes = buffer.bytes();
            assertEquals(currentBytes, data);
            // bytes() doesn't consume the data, it's still there
            assertEquals(buffer.size, 5);
        });
    });

    describe('write method with large data', () => {
        it('should handle writing data larger than initial capacity', () => {
            const buffer = new BytesBuffer(new ArrayBuffer(2));
            const largeData = new Uint8Array([1, 2, 3, 4]);
            buffer.write(largeData);
            assertEquals(buffer.size, 4);
            expect(buffer.capacity).toBeGreaterThanOrEqual(4);
            expect(buffer.bytes()).toEqual(largeData);
        });
    });

    describe('edge cases', () => {
        it('should handle empty writes and reads', () => {
            const buffer = new BytesBuffer(new ArrayBuffer(10));
            const emptyData = new Uint8Array(0);
            buffer.write(emptyData);
            assertEquals(buffer.size, 0);

            const readBuf = new Uint8Array(5);
            const bytesRead = buffer.read(readBuf);
            assertEquals(bytesRead, 0);
            assertEquals(buffer.size, 0);
        });

        it('should handle reset correctly', () => {
            const buffer = new BytesBuffer(new ArrayBuffer(10));
            const data = new Uint8Array([1, 2, 3]);
            buffer.write(data);
            assertEquals(buffer.size, 3);
            buffer.reset();
            assertEquals(buffer.size, 0);
        });
    });

    describe('writeVarint', () => {
        it('should write 1-byte varint', () => {
            const buf = new Uint8Array(10);
            const len = writeVarint(buf, 63);
            assertEquals(len, 1);
            assertEquals(buf[0], 63);
        });

        it('should write 2-byte varint', () => {
            const buf = new Uint8Array(10);
            const len = writeVarint(buf, 100);
            assertEquals(len, 2);
            assertEquals(buf[0], (100 >> 8) | 0x40);
            assertEquals(buf[1], 100 & 0xff);
        });

        it('should write 4-byte varint', () => {
            const buf = new Uint8Array(10);
            const len = writeVarint(buf, 100000);
            assertEquals(len, 4);
        });

        it('should write 8-byte varint', () => {
            const buf = new Uint8Array(10);
            const len = writeVarint(buf, 10000000000);
            assertEquals(len, 8);
        });

        it('should throw for negative numbers', () => {
            const buf = new Uint8Array(10);
            expect(() => writeVarint(buf, -1)).toThrow();
        });

        it('should throw for buffer too small', () => {
            const buf = new Uint8Array(1);
            expect(() => writeVarint(buf, 100)).toThrow();
        });
    });

    describe('writeBigVarint', () => {
    it('should write bigint varint', () => {
        const buf = new Uint8Array(10);
        const len = writeBigVarint(buf, 100n);
        assertEquals(len, 2); // 100 > 63, so 2 bytes
        assertEquals(buf[0], (100 >> 8) | 0x40);
        assertEquals(buf[1], 100 & 0xff);
    });        it('should throw for negative bigint', () => {
            const buf = new Uint8Array(10);
            expect(() => writeBigVarint(buf, -1n)).toThrow();
        });
    });

    describe('writeUint8Array', () => {
        it('should write uint8array with varint length', () => {
            const buf = new Uint8Array(20);
            const data = new Uint8Array([1, 2, 3]);
            const len = writeUint8Array(buf, data);
            assertEquals(len, 4); // 1 (varint) + 3
            expect(buf.subarray(1, 4)).toEqual(data);
        });
    });

    describe('writeString', () => {
        it('should write string', () => {
            const buf = new Uint8Array(20);
            const len = writeString(buf, "abc");
            assertEquals(len, 4); // 1 + 3
        });
    });

    describe('readVarint', () => {
        it('should read 1-byte varint', () => {
            const buf = new Uint8Array([63]);
            const [value, len] = readVarint(buf);
            assertEquals(value, 63);
            assertEquals(len, 1);
        });

        it('should read 2-byte varint', () => {
            const buf = new Uint8Array([(100 >> 8) | 0x40, 100 & 0xff]);
            const [value, len] = readVarint(buf);
            assertEquals(value, 100);
            assertEquals(len, 2);
        });
    });

    describe('readBigVarint', () => {
        it('should read bigint varint', () => {
            const buf = new Uint8Array([(100 >> 8) | 0x40, 100 & 0xff]); // 2 bytes for 100
            const [value, len] = readBigVarint(buf);
            assertEquals(value, 100n);
            assertEquals(len, 2);
        });
    });

    describe('readUint8Array', () => {
        it('should read uint8array', () => {
            const buf = new Uint8Array([3, 1, 2, 3]); // len=3, data=1,2,3
            const [data, len] = readUint8Array(buf);
            assertEquals(data, new Uint8Array([1, 2, 3]));
            assertEquals(len, 4);
        });
    });

    describe('readString', () => {
        it('should read string', () => {
            const buf = new Uint8Array([3, 97, 98, 99]); // len=3, "abc"
            const [str, len] = readString(buf);
            assertEquals(str, "abc");
            assertEquals(len, 4);
        });
    });
});
