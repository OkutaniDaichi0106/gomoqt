import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { SessionServerMessage } from './session_server';
import { Writer, Reader } from '../io';
import { Version } from '../internal/version';
import { Extensions } from '../internal/extensions';
import { createIsolatedStreams } from './test-utils.test';

describe('SessionServerMessage', () => {
  it('should encode and decode', async () => {
    const version = 1n;
    const extensions = new Extensions();
    extensions.addBytes(1n, new Uint8Array([1, 2, 3]));
    extensions.addBytes(2n, new Uint8Array([4, 5, 6]));

    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      const [encodedMessage, encodeErr] = await SessionServerMessage.encode(writer, version, extensions);
      expect(encodeErr).toBeUndefined();
      expect(encodedMessage).toBeDefined();
      expect(encodedMessage?.version).toEqual(version);
      expect(encodedMessage?.extensions).toEqual(extensions);

      // Close writer to signal end of stream
      await writer.close();

      const [decodedMessage, decodeErr] = await SessionServerMessage.decode(reader);
      expect(decodeErr).toBeUndefined();
      expect(decodedMessage).toBeDefined();
      expect(decodedMessage?.version).toEqual(version);
      expect(decodedMessage?.extensions).toEqual(extensions);
    } finally {
      await cleanup();
    }
  });
});
