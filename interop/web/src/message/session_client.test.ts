import { SessionClientMessage } from './session_client';
import { Writer, Reader } from '../io';
import { Version } from '../internal/version';
import { Extensions } from '../internal/extensions';

// Simple test helper using memory-based reader/writer
async function testEncodeDecodeMessage(versions: Set<Version>, extensions: Extensions): Promise<[SessionClientMessage?, Error?]> {
  const { readable, writable } = new TransformStream<Uint8Array, Uint8Array>();
  const writer = new Writer(writable);
  const reader = new Reader(readable); // Cast to any to satisfy TypeScript

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

  return await SessionClientMessage.decode(reader);
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
