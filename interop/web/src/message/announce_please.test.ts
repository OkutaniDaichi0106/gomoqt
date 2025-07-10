import { AnnouncePleaseMessage } from './announce_please';
import { Writer, Reader } from '../io';

describe('AnnouncePleaseMessage', () => {
  it('should encode and decode', async () => {
    const prefix = 'test';

    // Create a TransformStream to connect readable and writable streams
    const { readable, writable } = new TransformStream<Uint8Array, Uint8Array>();
    
    const writer = new Writer(writable);
    const reader = new Reader(readable);

    const [encodedMessage, encodeErr] = await AnnouncePleaseMessage.encode(writer, prefix);
    expect(encodeErr).toBeUndefined();
    expect(encodedMessage).toBeDefined();
    expect(encodedMessage?.prefix).toEqual(prefix);

    const [decodedMessage, decodeErr] = await AnnouncePleaseMessage.decode(reader);
    expect(decodeErr).toBeUndefined();
    expect(decodedMessage).toBeDefined();
    expect(decodedMessage?.prefix).toEqual(prefix);
  });
});
