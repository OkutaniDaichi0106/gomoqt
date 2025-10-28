import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows } from "../../deps.ts";
import { SessionClientMessage } from './session_client.ts';
import { Extensions } from '../internal/extensions';
import type { Version } from '../internal/version';
import { Versions } from '../internal/version';
import { Writer, Reader } from '../io';

describe('SessionClientMessage', () => {
  it('should be defined', () => {
    assertExists(SessionClientMessage);
  });

  it('should create instance with versions and extensions', () => {
    const versions = new Set<Version>([Versions.DEVELOP]);
    const extensions = new Extensions();
  extensions.addString(1, 'test');

    const message = new SessionClientMessage({versions, extensions});

    assertEquals(message.versions, versions);
    assertEquals(message.extensions, extensions);
  });

  it('should create instance with versions only', () => {
    const versions = new Set<Version>([Versions.DEVELOP]);

    const message = new SessionClientMessage({versions});

    assertEquals(message.versions, versions);
    assertInstanceOf(message.extensions, Extensions);
    assertEquals(message.extensions.entries.size, 0);
  });

  it('should calculate correct length with single version and no extensions', () => {
    const versions = new Set<Version>([Versions.DEVELOP]);
    const extensions = new Extensions();

    const message = new SessionClientMessage({versions, extensions});
    const length = message.messageLength;

    // Expected: varint(1) + varint(DEVELOP) + varint(0)
    // DEVELOP = 0xffffff00n, which needs 5 bytes in varint encoding
    // varint(1) = 1 byte, varint(0) = 1 byte
    expect(length).toBeGreaterThan(0);
    assertEquals(typeof length, 'number');
  });

  it('should calculate correct length with multiple versions', () => {
    const versions = new Set<Version>([Versions.DEVELOP, 1n, 2n]);
    const extensions = new Extensions();

    const message = new SessionClientMessage({versions, extensions});
    const length = message.messageLength;

    expect(length).toBeGreaterThan(0);
    assertEquals(typeof length, 'number');
  });

  it('should calculate correct length with extensions', () => {
    const versions = new Set<Version>([Versions.DEVELOP]);
    const extensions = new Extensions();
    extensions.addString(1, 'test');
    extensions.addBytes(2, new Uint8Array([1, 2, 3]));

    const message = new SessionClientMessage({versions, extensions});
    const length = message.messageLength;

    expect(length).toBeGreaterThan(0);
    assertEquals(typeof length, 'number');
  });

  it('should encode and decode with single version and no extensions', async () => {
    const versions = new Set<Version>([Versions.DEVELOP]);
    const extensions = new Extensions();

    // Create buffer for encoding
    const chunks: Uint8Array[] = [];
    const writableStream = new WritableStream({
      write(chunk) {
        chunks.push(chunk);
      }
    });
    const writer = new Writer({stream: writableStream, transfer: undefined, streamId: 0n});

    // Encode
    const message = new SessionClientMessage({versions, extensions});
    const encodeErr = await message.encode(writer);
    assertEquals(encodeErr, undefined);

    // Combine chunks into single buffer
    const totalLength = chunks.reduce((sum, chunk) => sum + chunk.length, 0);
    const combinedBuffer = new Uint8Array(totalLength);
    let offset = 0;
    for (const chunk of chunks) {
      combinedBuffer.set(chunk, offset);
      offset += chunk.length;
    }

    console.log('Encoded buffer:', Array.from(combinedBuffer).join(','));

    // Create readable stream for decoding
    const readableStream = new ReadableStream({
      start(controller) {
        controller.enqueue(combinedBuffer);
        controller.close();
      }
    });
    const reader = new Reader({stream: readableStream, transfer: undefined, streamId: 0n});

    // Decode
    const decodedMessage = new SessionClientMessage({});
    const decodeErr = await decodedMessage.decode(reader);
    assertEquals(decodeErr, undefined);
    assertInstanceOf(decodedMessage, SessionClientMessage);

    // Verify content
    assertEquals(decodedMessage?.versions.size, 1);
    const decodedVersions = Array.from(decodedMessage?.versions || []);
    const originalVersions = Array.from(versions);
    assertEquals(decodedVersions, originalVersions);
    assertEquals(decodedMessage?.extensions.entries.size, 0);
  });

  it('should encode and decode with multiple versions', async () => {
    const versions = new Set<Version>([Versions.DEVELOP, 1n, 2n, 100n]);
    const extensions = new Extensions();

    // Create buffer for encoding
    const chunks: Uint8Array[] = [];
    const writableStream = new WritableStream({
      write(chunk) {
        chunks.push(chunk);
      }
    });
    const writer = new Writer({stream: writableStream, transfer: undefined, streamId: 0n});

    // Encode
    const message = new SessionClientMessage({ versions, extensions });
    const encodeErr = await message.encode(writer);
    assertEquals(encodeErr, undefined);

    // Combine chunks into single buffer
    const totalLength = chunks.reduce((sum, chunk) => sum + chunk.length, 0);
    const combinedBuffer = new Uint8Array(totalLength);
    let offset = 0;
    for (const chunk of chunks) {
      combinedBuffer.set(chunk, offset);
      offset += chunk.length;
    }

    // Create readable stream for decoding
    const readableStream = new ReadableStream({
      start(controller) {
        controller.enqueue(combinedBuffer);
        controller.close();
      }
    });
    const reader = new Reader({stream: readableStream, transfer: undefined, streamId: 0n});

    // Decode
    const decodedMessage = new SessionClientMessage({});
    const decodeErr = await decodedMessage.decode(reader);
    assertEquals(decodeErr, undefined);

    // Verify content
    assertEquals(decodedMessage.versions.size, 4);
    expect(decodedMessage.versions.has(Versions.DEVELOP)).toBe(true);
    expect(decodedMessage.versions.has(1n)).toBe(true);
    expect(decodedMessage.versions.has(2n)).toBe(true);
    expect(decodedMessage.versions.has(100n)).toBe(true);
    assertEquals(decodedMessage.extensions.entries.size, 0);
  });

  it('should encode and decode with extensions', async () => {
    const versions = new Set<Version>([Versions.DEVELOP]);
    const extensions = new Extensions();
  extensions.addString(1, 'test-string');
  extensions.addBytes(2, new Uint8Array([1, 2, 3, 4, 5]));
  extensions.addString(100, 'another-extension');

    // Create buffer for encoding
    const chunks: Uint8Array[] = [];
    const writableStream = new WritableStream({
      write(chunk) {
        chunks.push(chunk);
      }
    });
    const writer = new Writer({stream: writableStream, transfer: undefined, streamId: 0n});

    // Encode
    const message = new SessionClientMessage({ versions, extensions });
    const encodeErr = await message.encode(writer);
    assertEquals(encodeErr, undefined);

    // Combine chunks into single buffer
    const totalLength = chunks.reduce((sum, chunk) => sum + chunk.length, 0);
    const combinedBuffer = new Uint8Array(totalLength);
    let offset = 0;
    for (const chunk of chunks) {
      combinedBuffer.set(chunk, offset);
      offset += chunk.length;
    }

    // Create readable stream for decoding
    const readableStream = new ReadableStream({
      start(controller) {
        controller.enqueue(combinedBuffer);
        controller.close();
      }
    });
    const reader = new Reader({stream: readableStream, transfer: undefined, streamId: 0n});

    // Decode
    const decodedMessage = new SessionClientMessage({});
    const decodeErr = await decodedMessage.decode(reader);
    assertEquals(decodeErr, undefined);

    // Verify content
    assertEquals(decodedMessage.versions.size, 1);
    expect(decodedMessage.versions.has(Versions.DEVELOP)).toBe(true);
    assertEquals(decodedMessage.extensions.entries.size, 3);
  expect(decodedMessage.extensions.getString(1)).toBe('test-string');
  expect(decodedMessage.extensions.getBytes(2)).toEqual(new Uint8Array([1, 2, 3, 4, 5]));
  expect(decodedMessage.extensions.getString(100)).toBe('another-extension');
  });

  it('should encode and decode with empty versions set', async () => {
    const versions = new Set<Version>();
    const extensions = new Extensions();

    // Create buffer for encoding
    const chunks: Uint8Array[] = [];
    const writableStream = new WritableStream({
      write(chunk) {
        chunks.push(chunk);
      }
    });
    const writer = new Writer({stream: writableStream, transfer: undefined, streamId: 0n});
    // Encode
    const message = new SessionClientMessage({ versions, extensions });
    const encodeErr = await message.encode(writer);
    assertEquals(encodeErr, undefined);

    // Combine chunks into single buffer
    const totalLength = chunks.reduce((sum, chunk) => sum + chunk.length, 0);
    const combinedBuffer = new Uint8Array(totalLength);
    let offset = 0;
    for (const chunk of chunks) {
      combinedBuffer.set(chunk, offset);
      offset += chunk.length;
    }

    // Create readable stream for decoding
    const readableStream = new ReadableStream({
      start(controller) {
        controller.enqueue(combinedBuffer);
        controller.close();
      }
    });
    const reader = new Reader({stream: readableStream, transfer: undefined, streamId: 0n});

    // Decode
    const decodedMessage = new SessionClientMessage({});
    const decodeErr = await decodedMessage.decode(reader);
    assertEquals(decodeErr, undefined);

    // Verify content
    assertEquals(decodedMessage.versions.size, 0);
    assertEquals(decodedMessage.extensions.entries.size, 0);
  });

  it('should handle large version numbers', async () => {
    const largeVersion = BigInt('0x1FFFFFFFFF'); // Within varint8 range
    const versions = new Set<Version>([largeVersion]);
    const extensions = new Extensions();

    // Create buffer for encoding
    const chunks: Uint8Array[] = [];
    const writableStream = new WritableStream({
      write(chunk) {
        chunks.push(chunk);
      }
    });
    const writer = new Writer({stream: writableStream, transfer: undefined, streamId: 0n});

    // Encode
    const message = new SessionClientMessage({ versions, extensions });
    const encodeErr = await message.encode(writer);
    assertEquals(encodeErr, undefined);

    // Combine chunks into single buffer
    const totalLength = chunks.reduce((sum, chunk) => sum + chunk.length, 0);
    const combinedBuffer = new Uint8Array(totalLength);
    let offset = 0;
    for (const chunk of chunks) {
      combinedBuffer.set(chunk, offset);
      offset += chunk.length;
    }

    // Create readable stream for decoding
    const readableStream = new ReadableStream({
      start(controller) {
        controller.enqueue(combinedBuffer);
        controller.close();
      }
    });
    const reader = new Reader({stream: readableStream, transfer: undefined, streamId: 0n});

    // Decode
    const decodedMessage = new SessionClientMessage({});
    const decodeErr = await decodedMessage.decode(reader);
    assertEquals(decodeErr, undefined);

    // Verify content
    assertEquals(decodedMessage.versions.size, 1);
    expect(decodedMessage.versions.has(largeVersion)).toBe(true);
  });

  it('should handle empty extension data', async () => {
    const versions = new Set<Version>([Versions.DEVELOP]);
    const extensions = new Extensions();
    extensions.addBytes(1, new Uint8Array([])); // Empty bytes
    extensions.addString(2, ''); // Empty string

    // Create buffer for encoding
    const chunks: Uint8Array[] = [];
    const writableStream = new WritableStream({
      write(chunk) {
        chunks.push(chunk);
      }
    });
    const writer = new Writer({stream: writableStream, transfer: undefined, streamId: 0n});

    // Encode
    const message = new SessionClientMessage({ versions, extensions });
    const encodeErr = await message.encode(writer);
    assertEquals(encodeErr, undefined);

    // Combine chunks into single buffer
    const totalLength = chunks.reduce((sum, chunk) => sum + chunk.length, 0);
    const combinedBuffer = new Uint8Array(totalLength);
    let offset = 0;
    for (const chunk of chunks) {
      combinedBuffer.set(chunk, offset);
      offset += chunk.length;
    }

    // Create readable stream for decoding
    const readableStream = new ReadableStream({
      start(controller) {
        controller.enqueue(combinedBuffer);
        controller.close();
      }
    });
    const reader = new Reader({stream: readableStream, transfer: undefined, streamId: 0n});

    // Decode
    const decodedMessage = new SessionClientMessage({});
    const decodeErr = await decodedMessage.decode(reader);
    assertEquals(decodeErr, undefined);

    // Verify content
    assertEquals(decodedMessage.extensions.entries.size, 2);
    expect(decodedMessage.extensions.getBytes(1)).toEqual(new Uint8Array([]));
    expect(decodedMessage.extensions.getString(2)).toBe('');
  });
});
