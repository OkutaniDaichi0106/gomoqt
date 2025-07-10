import { FrameMessage } from './frame';
import { Writer, Reader } from '../io';

describe('FrameMessage', () => {
  it('should encode and decode', async () => {
    const data = new Uint8Array([1, 2, 3, 4, 5]);

    // Create a TransformStream to connect readable and writable streams
    const { readable, writable } = new TransformStream<Uint8Array, Uint8Array>();
    const writer = new Writer(writable);
    const reader = new Reader(readable);

    const [encodedMessage, encodeErr] = await FrameMessage.encode(writer, data);
    expect(encodeErr).toBeUndefined();
    expect(encodedMessage).toBeDefined();
    expect(encodedMessage?.data).toEqual(data);

    const [decodedMessage, decodeErr] = await FrameMessage.decode(reader);
    expect(decodeErr).toBeUndefined();
    expect(decodedMessage).toBeDefined();
    expect(decodedMessage?.data).toEqual(data);
  });
});
