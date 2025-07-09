import { AnnounceMessage } from './announce';
import { Writer, Reader } from '../io';

describe('AnnounceMessage', () => {
  it('should encode and decode', async () => {
    const suffix = 'test';
    const active = true;

    // Create a TransformStream to connect readable and writable streams
    const { readable, writable } = new TransformStream<Uint8Array, Uint8Array>();
    const writer = new Writer(writable);
    const reader = new Reader(readable);

    const [encodedMessage, encodeErr] = await AnnounceMessage.encode(writer, suffix, active);
    expect(encodeErr).toBeUndefined();
    expect(encodedMessage).toBeDefined();
    expect(encodedMessage?.suffix).toEqual(suffix);
    expect(encodedMessage?.active).toEqual(active);

    const [decodedMessage, decodeErr] = await AnnounceMessage.decode(reader);
    expect(decodeErr).toBeUndefined();
    expect(decodedMessage).toBeDefined();
    expect(decodedMessage?.suffix).toEqual(suffix);
    expect(decodedMessage?.active).toEqual(active);
  });
});
