import { SessionUpdateMessage } from './session_update';
import { Writer, Reader } from '../io';

describe('SessionUpdateMessage', () => {
  it('should encode and decode', async () => {
    const bitrate = 1000n;

    // Create a TransformStream to connect readable and writable streams
    const { readable, writable } = new TransformStream<Uint8Array, Uint8Array>();
    const writer = new Writer(writable);
    const reader = new Reader(readable);

    const [encodedMessage, encodeErr] = await SessionUpdateMessage.encode(writer, bitrate);
    expect(encodeErr).toBeUndefined();
    expect(encodedMessage).toBeDefined();
    expect(encodedMessage?.bitrate).toEqual(bitrate);

    const [decodedMessage, decodeErr] = await SessionUpdateMessage.decode(reader);
    expect(decodeErr).toBeUndefined();
    expect(decodedMessage).toBeDefined();
    expect(decodedMessage?.bitrate).toEqual(bitrate);
  });
});
