import { assertEquals, assertThrows, fail } from "../../deps.ts";
import { BytesBuffer, writeVarint, writeBigVarint, writeUint8Array, writeString, readVarint, readBigVarint, readUint8Array, readString } from './bytes.ts';

Deno.test('webtransport/bytes - varint and basic io', async (t) => {
    await t.step('varint roundtrip small values', () => {
        const buf = new Uint8Array(8);
        const len = writeVarint(buf, 42, 0);
        const [v, n] = readVarint(buf, 0);
        assertEquals(n, len);
        assertEquals(v, 42);
    });

    await t.step('varint 2-byte roundtrip', () => {
        const buf = new Uint8Array(8);
        const len = writeVarint(buf, 0x123, 0);
        const [v, n] = readVarint(buf, 0);
        assertEquals(n, len);
        assertEquals(v, 0x123);
    });

    await t.step('write/read bytes roundtrip', () => {
        const data = new Uint8Array([1,2,3,4,5]);
        const buf = new Uint8Array(16);
        const wrote = writeUint8Array(buf, data, 0);
        const [out, n] = readUint8Array(buf, 0);
        assertEquals(n, wrote);
        // compare typed arrays directly
        assertEquals(out, data);
    });

    await t.step('write/read string roundtrip', () => {
        const s = 'hello こんにちは';
        const buf = new Uint8Array(64);
        const w = writeString(buf, s, 0);
        const [out, n] = readString(buf, 0);
        assertEquals(n, w);
        assertEquals(out, s);
    });
});

Deno.test('webtransport/BytesBuffer behavior', async (t) => {
    await t.step('should write and read data', () => {
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

    await t.step('should grow when capacity is exceeded', () => {
        const buffer = new BytesBuffer(new ArrayBuffer(2));
        const data = new Uint8Array([1, 2]);
        buffer.write(data);
        if (buffer.capacity < 2) fail(`capacity ${buffer.capacity} < 2`);
        const moreData = new Uint8Array([3, 4, 5]);
        buffer.write(moreData);
        if (buffer.capacity < 5) fail(`capacity ${buffer.capacity} < 5`);
    });

    await t.step('readUint8 should read a single byte', () => {
        const buffer = new BytesBuffer(new ArrayBuffer(10));
        const data = new Uint8Array([1, 2, 3]);
        buffer.write(data);
        assertEquals(buffer.size, 3);
        const byte = buffer.readUint8();
        assertEquals(byte, 1);
        assertEquals(buffer.size, 2);
    });

    await t.step('writeUint8 should write single bytes', () => {
        const buffer = new BytesBuffer(new ArrayBuffer(10));
        buffer.writeUint8(42);
        buffer.writeUint8(43);
        buffer.writeUint8(44);
        buffer.writeUint8(45);
        buffer.writeUint8(46);
        assertEquals(buffer.size, 5);
    });

    await t.step('reserve should return writable buffer with sufficient capacity', () => {
        const buffer = new BytesBuffer(new ArrayBuffer(10));
        const data1 = new Uint8Array([1, 2, 3]);
        buffer.write(data1);
        assertEquals(buffer.size, 3);

        const reservedBuffer = buffer.reserve(6);
        if (reservedBuffer.length < 6) fail(`reserved length ${reservedBuffer.length} < 6`);
        if (buffer.capacity < 9) fail(`capacity ${buffer.capacity} < 9`);
    });

    await t.step('construction with initial data initializes capacity', () => {
        const initialData = new Uint8Array([1, 2, 3]);
        const buffer = new BytesBuffer(initialData.buffer);
        // Initially buffer is empty for writing, data needs to be written first
        assertEquals(buffer.size, 0);
        if (buffer.capacity < 3) fail(`capacity ${buffer.capacity} < 3`);
    });

    await t.step('bytes method returns content without consuming', () => {
        const buffer = new BytesBuffer(new ArrayBuffer(10));
        const data = new Uint8Array([1, 2, 3, 4, 5]);
        buffer.write(data);

        const currentBytes = buffer.bytes();
        assertEquals(currentBytes, data);
        // bytes() doesn't consume the data, it's still there
        assertEquals(buffer.size, 5);
    });

    await t.step('write method handles large data', () => {
        const buffer = new BytesBuffer(new ArrayBuffer(2));
        const largeData = new Uint8Array([1, 2, 3, 4]);
        buffer.write(largeData);
        assertEquals(buffer.size, 4);
        if (buffer.capacity < 4) fail(`capacity ${buffer.capacity} < 4`);
        assertEquals(buffer.bytes(), largeData);
    });

    await t.step('edge cases: empty writes and reads', () => {
        const buffer = new BytesBuffer(new ArrayBuffer(10));
        const emptyData = new Uint8Array(0);
        buffer.write(emptyData);
        assertEquals(buffer.size, 0);

        const readBuf = new Uint8Array(5);
        const bytesRead = buffer.read(readBuf);
        assertEquals(bytesRead, 0);
        assertEquals(buffer.size, 0);
    });

    await t.step('edge cases: reset', () => {
        const buffer = new BytesBuffer(new ArrayBuffer(10));
        const data = new Uint8Array([1, 2, 3]);
        buffer.write(data);
        assertEquals(buffer.size, 3);
        buffer.reset();
        assertEquals(buffer.size, 0);
    });

    await t.step('writeVarint encodings and errors', () => {
        const buf1 = new Uint8Array(10);
        const len1 = writeVarint(buf1, 63);
        assertEquals(len1, 1);
        assertEquals(buf1[0], 63);

        const buf2 = new Uint8Array(10);
        const len2 = writeVarint(buf2, 100);
        assertEquals(len2, 2);
        assertEquals(buf2[0], (100 >> 8) | 0x40);
        assertEquals(buf2[1], 100 & 0xff);

        const buf3 = new Uint8Array(10);
        const len3 = writeVarint(buf3, 100000);
        assertEquals(len3, 4);

        const buf4 = new Uint8Array(10);
        const len4 = writeVarint(buf4, 10000000000);
        assertEquals(len4, 8);

        const bufNeg = new Uint8Array(10);
        assertThrows(() => writeVarint(bufNeg, -1));

        const smallBuf = new Uint8Array(1);
        assertThrows(() => writeVarint(smallBuf, 100));
    });

    await t.step('writeBigVarint encodings and errors', () => {
        const buf = new Uint8Array(10);
        const len = writeBigVarint(buf, 100n);
        assertEquals(len, 2); // 100 > 63, so 2 bytes
        assertEquals(buf[0], (100 >> 8) | 0x40);
        assertEquals(buf[1], 100 & 0xff);

        const bufNeg = new Uint8Array(10);
        assertThrows(() => writeBigVarint(bufNeg, -1n));
    });

    await t.step('writeUint8Array writes varint length + data', () => {
        const buf = new Uint8Array(20);
        const data = new Uint8Array([1, 2, 3]);
        const len = writeUint8Array(buf, data);
        assertEquals(len, 4); // 1 (varint) + 3
        assertEquals(buf.subarray(1, 4), data);
    });

    await t.step('writeString writes length + bytes', () => {
        const buf = new Uint8Array(20);
        const len = writeString(buf, "abc");
        assertEquals(len, 4); // 1 + 3
    });

    await t.step('readVarint', () => {
        const buf = new Uint8Array([63]);
        const [value, len] = readVarint(buf);
        assertEquals(value, 63);
        assertEquals(len, 1);

        const buf2 = new Uint8Array([(100 >> 8) | 0x40, 100 & 0xff]);
        const [value2, len2] = readVarint(buf2);
        assertEquals(value2, 100);
        assertEquals(len2, 2);
    });

    await t.step('readBigVarint', () => {
        const buf = new Uint8Array([(100 >> 8) | 0x40, 100 & 0xff]); // 2 bytes for 100
        const [value, len] = readBigVarint(buf);
        assertEquals(value, 100n);
        assertEquals(len, 2);
    });

    await t.step('readUint8Array', () => {
        const buf = new Uint8Array([3, 1, 2, 3]); // len=3, data=1,2,3
        const [data, len] = readUint8Array(buf);
        assertEquals(data, new Uint8Array([1, 2, 3]));
        assertEquals(len, 4);
    });

    await t.step('readString', () => {
        const buf = new Uint8Array([3, 97, 98, 99]); // len=3, "abc"
        const [str, len] = readString(buf);
        assertEquals(str, "abc");
        assertEquals(len, 4);
    });
});
