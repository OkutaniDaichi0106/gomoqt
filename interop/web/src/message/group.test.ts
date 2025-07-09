import { GroupMessage } from './group';
import { Writer, Reader } from '../io';

describe('GroupMessage', () => {
  it('should encode and decode', async () => {
    const subscribeId = 123n;
    const sequence = 456n;

    // Create a TransformStream to connect readable and writable streams
    const { readable, writable } = new TransformStream<Uint8Array, Uint8Array>();
    const writer = new Writer(writable);
    const reader = new Reader(readable);

    const [encodedMessage, encodeErr] = await GroupMessage.encode(writer, subscribeId, sequence);
    expect(encodeErr).toBeUndefined();
    expect(encodedMessage).toBeDefined();
    expect(encodedMessage?.subscribeId).toEqual(subscribeId);
    expect(encodedMessage?.sequence).toEqual(sequence);

    const [decodedMessage, decodeErr] = await GroupMessage.decode(reader);
    expect(decodeErr).toBeUndefined();
    expect(decodedMessage).toBeDefined();
    expect(decodedMessage?.subscribeId).toEqual(subscribeId);
    expect(decodedMessage?.sequence).toEqual(sequence);
  });
});
