import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { SessionServerMessage } from './session_server';
import { Writer, Reader } from '../webtransport';
import { Version } from '../internal/version';
import { Extensions } from '../internal/extensions';
import { createIsolatedStreams } from './test-utils.test';

describe('SessionServerMessage', () => {
  it('should encode and decode', async () => {
    const version = 1n;
    const extensions = new Extensions();
  extensions.addBytes(1, new Uint8Array([1, 2, 3]));
  extensions.addBytes(2, new Uint8Array([4, 5, 6]));

    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      const message = new SessionServerMessage({ version, extensions });
      const encodeErr = await message.encode(writer);
      expect(encodeErr).toBeUndefined();

      // Close writer to signal end of stream
      await writer.close();

      const decodedMessage = new SessionServerMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      expect(decodeErr).toBeUndefined();
      expect(decodedMessage.version).toEqual(version);
      expect(decodedMessage.extensions).toEqual(extensions);
    } finally {
      await cleanup();
    }
  });
});
