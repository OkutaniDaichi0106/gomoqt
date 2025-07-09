import { SubscribeOkMessage } from './subscribe_ok';
import { Writer, Reader } from '../io';

describe('SubscribeOkMessage', () => {
  it('should encode and decode', async () => {
    const groupOrder = 1n;

    // Create a TransformStream to connect readable and writable streams
    const { readable, writable } = new TransformStream<Uint8Array, Uint8Array>();
    const writer = new Writer(writable);
    const reader = new Reader(readable);

    const [encodedMessage, encodeErr] = await SubscribeOkMessage.encode(writer, groupOrder);
    expect(encodeErr).toBeUndefined();
    expect(encodedMessage).toBeDefined();
    expect(encodedMessage?.groupOrder).toEqual(groupOrder);

    const [decodedMessage, decodeErr] = await SubscribeOkMessage.decode(reader);
    expect(decodeErr).toBeUndefined();
    expect(decodedMessage).toBeDefined();
    expect(decodedMessage?.groupOrder).toEqual(groupOrder);
  });
});
