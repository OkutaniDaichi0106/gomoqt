import { SessionClientMessage } from './session_client';
import { Writer, Reader } from '../io';
import { Version } from '../internal/version';
import { Extensions } from '../internal/extensions';

// Test-specific memory-based Reader/Writer implementation
class TestMemoryWriter {
  private buffer: Uint8Array[] = [];

  writeVarint(num: bigint): void {
    if (num < 0) {
      throw new Error("Varint cannot be negative");
    }

    if (num < 64n) {
      this.buffer.push(new Uint8Array([Number(num)]));
    } else if (num < 16384n) {
      this.buffer.push(new Uint8Array([
        0x40 | Number(num >> 8n),
        Number(num & 0xFFn)
      ]));
    } else if (num < 1073741824n) {
      this.buffer.push(new Uint8Array([
        0x80 | Number(num >> 24n),
        Number((num >> 16n) & 0xFFn),
        Number((num >> 8n) & 0xFFn),
        Number(num & 0xFFn)
      ]));
    } else {
      this.buffer.push(new Uint8Array([
        0xC0 | Number(num >> 56n),
        Number((num >> 48n) & 0xFFn),
        Number((num >> 40n) & 0xFFn),
        Number((num >> 32n) & 0xFFn),
        Number((num >> 24n) & 0xFFn),
        Number((num >> 16n) & 0xFFn),
        Number((num >> 8n) & 0xFFn),
        Number(num & 0xFFn)
      ]));
    }
  }

  writeUint8Array(data: Uint8Array): void {
    this.writeVarint(BigInt(data.length));
    this.buffer.push(data.slice());
  }

  toBuffer(): Uint8Array {
    const totalLength = this.buffer.reduce((sum, chunk) => sum + chunk.length, 0);
    const result = new Uint8Array(totalLength);
    let offset = 0;
    for (const chunk of this.buffer) {
      result.set(chunk, offset);
      offset += chunk.length;
    }
    return result;
  }
}

class TestMemoryReader {
  private buffer: Uint8Array;
  private offset: number = 0;

  constructor(buffer: Uint8Array) {
    this.buffer = buffer;
  }

  async readVarint(): Promise<[bigint?, Error?]> {
    if (this.offset >= this.buffer.length) {
      return [undefined, new Error("End of buffer")];
    }

    const firstByte = this.buffer[this.offset++];
    const len = 1 << (firstByte >> 6);

    if (this.offset + len - 1 > this.buffer.length) {
      return [undefined, new Error("Incomplete varint")];
    }

    let value: bigint = BigInt(firstByte & 0x3f);

    for (let i = 1; i < len; i++) {
      value = (value << 8n) | BigInt(this.buffer[this.offset++]);
    }

    return [value, undefined];
  }

  async readUint8Array(): Promise<[Uint8Array?, Error?]> {
    const [len, err] = await this.readVarint();
    if (err) {
      return [undefined, err];
    }
    if (len === undefined) {
      return [undefined, new Error("Failed to read length")];
    }

    const length = Number(len);
    if (this.offset + length > this.buffer.length) {
      return [undefined, new Error("Incomplete byte array")];
    }

    const result = this.buffer.slice(this.offset, this.offset + length);
    this.offset += length;
    return [result, undefined];
  }
}

// Simple test helper using memory-based reader/writer
async function testEncodeDecodeMessage(versions: Set<Version>, extensions: Extensions): Promise<[SessionClientMessage?, Error?]> {
  const writer = new TestMemoryWriter();
  
  // Encode versions
  writer.writeVarint(BigInt(versions.size));
  for (const version of versions) {
    writer.writeVarint(version);
  }
  
  // Encode extensions
  writer.writeVarint(BigInt(extensions.entries.size));
  for (const [extId, extData] of extensions.entries) {
    writer.writeVarint(extId);
    writer.writeUint8Array(extData);
  }
  
  // Decode
  const buffer = writer.toBuffer();
  const reader = new TestMemoryReader(buffer);
  
  return await SessionClientMessage.decode(reader as any);
}

describe('SessionClientMessage', () => {
  it('should encode and decode', async () => {
    const versions = new Set<Version>([1n, 2n, 3n]);
    const extensions = new Extensions();
    extensions.addBytes(1n, new Uint8Array([1, 2, 3]));
    extensions.addBytes(2n, new Uint8Array([4, 5, 6]));

    const [decodedMessage, decodeErr] = await testEncodeDecodeMessage(versions, extensions);
    expect(decodeErr).toBeUndefined();
    expect(decodedMessage).toBeDefined();
    expect(decodedMessage?.versions).toEqual(versions);
    expect(decodedMessage?.extensions).toEqual(extensions);
  });

  it('should encode with empty extensions', async () => {
    const versions = new Set<Version>([1n]);
    const extensions = new Extensions();

    const [decodedMessage, decodeErr] = await testEncodeDecodeMessage(versions, extensions);
    expect(decodeErr).toBeUndefined();
    expect(decodedMessage).toBeDefined();
    expect(decodedMessage?.versions).toEqual(versions);
    expect(decodedMessage?.extensions.entries.size).toBe(0);
  });

  it('should handle single version', async () => {
    const versions = new Set<Version>([42n]);
    const extensions = new Extensions();
    extensions.addBytes(5n, new Uint8Array([0xde, 0xad, 0xbe, 0xef]));

    const [decodedMessage, decodeErr] = await testEncodeDecodeMessage(versions, extensions);
    expect(decodeErr).toBeUndefined();
    expect(decodedMessage).toBeDefined();
    expect(decodedMessage?.versions.has(42n)).toBe(true);
    expect(decodedMessage?.versions.size).toBe(1);
    expect(decodedMessage?.extensions.has(5n)).toBe(true);
  });

  it('should handle multiple versions and extensions', async () => {
    const versions = new Set<Version>([1n, 2n, 3n, 4n, 5n]);
    const extensions = new Extensions();
    extensions.addBytes(1n, new Uint8Array([0x01]));
    extensions.addBytes(2n, new Uint8Array([0x02, 0x03]));
    extensions.addBytes(3n, new Uint8Array([0x04, 0x05, 0x06]));

    const [decodedMessage, decodeErr] = await testEncodeDecodeMessage(versions, extensions);
    expect(decodeErr).toBeUndefined();
    expect(decodedMessage).toBeDefined();
    expect(decodedMessage?.versions.size).toBe(5);
    expect(decodedMessage?.extensions.entries.size).toBe(3);
    
    // Verify all versions are preserved
    for (const version of versions) {
      expect(decodedMessage?.versions.has(version)).toBe(true);
    }

    // Verify all extensions are preserved
    expect(decodedMessage?.extensions.has(1n)).toBe(true);
    expect(decodedMessage?.extensions.has(2n)).toBe(true);
    expect(decodedMessage?.extensions.has(3n)).toBe(true);
  });

  it('should create correct message object', () => {
    const versions = new Set<Version>([1n, 2n]);
    const extensions = new Extensions();
    extensions.addBytes(1n, new Uint8Array([1, 2, 3]));

    const message = new SessionClientMessage(versions, extensions);
    expect(message.versions).toEqual(versions);
    expect(message.extensions).toEqual(extensions);
  });

  it('should use default extensions when not provided', () => {
    const versions = new Set<Version>([1n]);
    const message = new SessionClientMessage(versions);
    
    expect(message.versions).toEqual(versions);
    expect(message.extensions).toBeInstanceOf(Extensions);
    expect(message.extensions.entries.size).toBe(0);
  });
});
