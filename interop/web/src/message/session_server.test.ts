import { SessionServerMessage } from './session_server';
import { Writer, Reader } from '../io';
import { Version } from '../internal/version';
import { Extensions } from '../internal/extensions';

describe('SessionServerMessage', () => {
  it('should encode and decode', async () => {
    const version = 1n;
    const extensions = new Extensions();
    extensions.addBytes(1n, new Uint8Array([1, 2, 3]));
    extensions.addBytes(2n, new Uint8Array([4, 5, 6]));

    // Create a TransformStream to connect readable and writable streams
    const { readable, writable } = new TransformStream<Uint8Array, Uint8Array>();
    const writer = new Writer(writable);
    const reader = new Reader(readable);

    const [encodedMessage, encodeErr] = await SessionServerMessage.encode(writer, version, extensions);
    expect(encodeErr).toBeUndefined();
    expect(encodedMessage).toBeDefined();
    expect(encodedMessage?.version).toEqual(version);
    expect(encodedMessage?.extensions).toEqual(extensions);

    const [decodedMessage, decodeErr] = await SessionServerMessage.decode(reader);
    expect(decodeErr).toBeUndefined();
    expect(decodedMessage).toBeDefined();
    expect(decodedMessage?.version).toEqual(version);
    expect(decodedMessage?.extensions).toEqual(extensions);
  });
});
