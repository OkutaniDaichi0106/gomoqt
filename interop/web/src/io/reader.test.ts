import { Reader } from './reader';

describe('Reader', () => {
    it('should initialize with an empty buffer', () => {
        const reader = new Reader();
        expect(reader.size()).toBe(0);
    });

    it('should initialize with a Uint8Array', () => {
        const data = new Uint8Array([1, 2, 3]);
        const reader = new Reader(data);
        expect(reader.size()).toBe(3);
    });

    it('should read bytes', () => {
        const data = new Uint8Array([1, 2, 3, 4, 5]);
        const reader = new Reader(data);
        const readData = reader.read(3);
        expect(readData).toEqual(new Uint8Array([1, 2, 3]));
        expect(reader.size()).toBe(2);
    });

    it('should return null when reading more bytes than available', () => {
        const data = new Uint8Array([1, 2]);
        const reader = new Reader(data);
        const readData = reader.read(3);
        expect(readData).toBeNull();
    });

    it('should read a single byte', () => {
        const data = new Uint8Array([10, 20]);
        const reader = new Reader(data);
        expect(reader.readByte()).toBe(10);
        expect(reader.size()).toBe(1);
        expect(reader.readByte()).toBe(20);
        expect(reader.size()).toBe(0);
        expect(reader.readByte()).toBeNull();
    });

    it('should read a Uint16', () => {
        const data = new Uint8Array([0x01, 0x02, 0x03, 0x04]); // 258, 772
        const reader = new Reader(data);
        expect(reader.readUint16()).toBe(258);
        expect(reader.size()).toBe(2);
        expect(reader.readUint16()).toBe(772);
        expect(reader.size()).toBe(0);
        expect(reader.readUint16()).toBeNull();
    });

    it('should read a Uint32', () => {
        const data = new Uint8Array([0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08]); // 16909060, 84281096
        const reader = new Reader(data);
        expect(reader.readUint32()).toBe(16909060);
        expect(reader.size()).toBe(4);
        expect(reader.readUint32()).toBe(84281096);
        expect(reader.size()).toBe(0);
        expect(reader.readUint32()).toBeNull();
    });

    it('should read a Uint64', () => {
        const data = new Uint8Array([0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10]);
        const reader = new Reader(data);
        expect(reader.readUint64()).toBe(0x0102030405060708n);
        expect(reader.size()).toBe(8);
        expect(reader.readUint64()).toBe(0x090A0B0C0D0E0F10n);
        expect(reader.size()).toBe(0);
        expect(reader.readUint64()).toBeNull();
    });

    it('should read a VarInt', () => {
        const reader1 = new Reader(new Uint8Array([0x01]));
        expect(reader1.readVarInt()).toBe(1);
        expect(reader1.size()).toBe(0);

        const reader2 = new Reader(new Uint8Array([0x81, 0x01])); // 129
        expect(reader2.readVarInt()).toBe(129);
        expect(reader2.size()).toBe(0);

        const reader3 = new Reader(new Uint8Array([0xFF, 0xFF, 0xFF, 0xFF, 0x0F]));
        expect(reader3.readVarInt()).toBe(4294967295);
        expect(reader3.size()).toBe(0);

        const reader4 = new Reader(new Uint8Array([0x80])); // Incomplete varint
        expect(reader4.readVarInt()).toBeNull();
    });

    it('should read bytes (alias for read)', () => {
        const data = new Uint8Array([1, 2, 3, 4, 5]);
        const reader = new Reader(data);
        const readData = reader.readBytes(3);
        expect(readData).toEqual(new Uint8Array([1, 2, 3]));
        expect(reader.size()).toBe(2);
    });

    it('should peek bytes without advancing offset', () => {
        const data = new Uint8Array([1, 2, 3, 4, 5]);
        const reader = new Reader(data);
        const peekedData = reader.peek(3);
        expect(peekedData).toEqual(new Uint8Array([1, 2, 3]));
        expect(reader.size()).toBe(5); // Offset should not advance
        const readData = reader.read(5);
        expect(readData).toEqual(new Uint8Array([1, 2, 3, 4, 5]));
    });

    it('should return null when peeking more bytes than available', () => {
        const data = new Uint8Array([1, 2]);
        const reader = new Reader(data);
        const peekedData = reader.peek(3);
        expect(peekedData).toBeNull();
    });

    it('should skip bytes', () => {
        const data = new Uint8Array([1, 2, 3, 4, 5]);
        const reader = new Reader(data);
        expect(reader.skip(2)).toBe(true);
        expect(reader.size()).toBe(3);
        const remaining = reader.read(3);
        expect(remaining).toEqual(new Uint8Array([3, 4, 5]));
    });

    it('should return false when skipping more bytes than available', () => {
        const data = new Uint8Array([1, 2]);
        const reader = new Reader(data);
        expect(reader.skip(3)).toBe(false);
        expect(reader.size()).toBe(2); // Size should remain unchanged
    });

    it('should reset the reader', () => {
        const data = new Uint8Array([1, 2, 3]);
        const reader = new Reader(data);
        reader.read(1);
        expect(reader.size()).toBe(2);
        reader.reset();
        expect(reader.size()).toBe(0);
    });
});