import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows } from "../../deps.ts";
import { AnnounceInitMessage } from './announce_init.ts';
import { Writer, Reader } from '../io';
import { createIsolatedStreams } from './test-utils.test';

describe('AnnounceInitMessage', () => {
  it('should encode and decode with empty suffixes array', async () => {
    const suffixes: string[] = [];
    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      // Encode the message
      const message = new AnnounceInitMessage({ suffixes });
      const encodeErr = await message.encode(writer);
      assertEquals(encodeErr, undefined);

      // Close writer to signal end of stream
      await writer.close();

      // Decode the message
      const decodedMessage = new AnnounceInitMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      assertEquals(decodeErr, undefined);
      assertEquals(decodedMessage.suffixes, suffixes);
    } finally {
      await cleanup();
    }
  });

  it('should encode and decode with single suffix', async () => {
    const suffixes = ['test-suffix'];
    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      // Encode the message
      const message = new AnnounceInitMessage({ suffixes });
      const encodeErr = await message.encode(writer);
      assertEquals(encodeErr, undefined);

      // Close writer to signal end of stream
      await writer.close();

      // Decode the message
      const decodedMessage = new AnnounceInitMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      assertEquals(decodeErr, undefined);
      assertEquals(decodedMessage.suffixes, suffixes);
    } finally {
      await cleanup();
    }
  });

  it('should encode and decode with multiple suffixes', async () => {
    const suffixes = ['suffix1', 'suffix2', 'suffix3'];
    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      // Encode the message
      const message = new AnnounceInitMessage({ suffixes });
      const encodeErr = await message.encode(writer);
      assertEquals(encodeErr, undefined);

      // Close writer to signal end of stream
      await writer.close();

      // Decode the message
      const decodedMessage = new AnnounceInitMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      assertEquals(decodeErr, undefined);
      assertEquals(decodedMessage.suffixes, suffixes);
    } finally {
      await cleanup();
    }
  });

  it('should handle special characters in suffixes', async () => {
    const suffixes = ['suffix-with-dashes', 'suffix_with_underscores', 'suffix/with/slashes', 'suffix with spaces'];
    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      // Encode the message
      const message = new AnnounceInitMessage({ suffixes });
      const encodeErr = await message.encode(writer);
      assertEquals(encodeErr, undefined);

      // Close writer to signal end of stream
      await writer.close();

      // Decode the message
      const decodedMessage = new AnnounceInitMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      assertEquals(decodeErr, undefined);
      assertEquals(decodedMessage.suffixes, suffixes);
    } finally {
      await cleanup();
    }
  });

  it('should create instance with constructor', () => {
    const suffixes = ['test1', 'test2'];
    const message = new AnnounceInitMessage({ suffixes });

    assertEquals(message.suffixes, suffixes);
  });

  it('should handle empty strings in suffixes array', async () => {
    const suffixes = ['', 'valid-suffix', ''];
    const { writer, reader, cleanup } = createIsolatedStreams();

    try {
      // Encode the message
      const message = new AnnounceInitMessage({ suffixes });
      const encodeErr = await message.encode(writer);
      assertEquals(encodeErr, undefined);

      // Close writer to signal end of stream
      await writer.close();

      // Decode the message
      const decodedMessage = new AnnounceInitMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      assertEquals(decodeErr, undefined);
      assertEquals(decodedMessage.suffixes, suffixes);
    } finally {
      await cleanup();
    }
  });
});
