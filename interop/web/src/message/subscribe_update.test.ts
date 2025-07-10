import { SubscribeUpdateMessage } from './subscribe_update';
import { Writer, Reader } from '../io';

describe('SubscribeUpdateMessage', () => {
  it('should encode and decode', async () => {
    const trackPriority = 1n;
    const minGroupSequence = 2n;
    const maxGroupSequence = 3n;

    // Create a TransformStream to connect readable and writable streams
    const { readable, writable } = new TransformStream<Uint8Array, Uint8Array>();
    const writer = new Writer(writable);
    const reader = new Reader(readable);

    const [encodedMessage, encodeErr] = await SubscribeUpdateMessage.encode(
      writer,
      trackPriority,
      minGroupSequence,
      maxGroupSequence
    );
    expect(encodeErr).toBeUndefined();
    expect(encodedMessage).toBeDefined();
    expect(encodedMessage?.trackPriority).toEqual(trackPriority);
    expect(encodedMessage?.minGroupSequence).toEqual(minGroupSequence);
    expect(encodedMessage?.maxGroupSequence).toEqual(maxGroupSequence);

    const [decodedMessage, decodeErr] = await SubscribeUpdateMessage.decode(reader);
    expect(decodeErr).toBeUndefined();
    expect(decodedMessage).toBeDefined();
    expect(decodedMessage?.trackPriority).toEqual(trackPriority);
    expect(decodedMessage?.minGroupSequence).toEqual(minGroupSequence);
    expect(decodedMessage?.maxGroupSequence).toEqual(maxGroupSequence);
  });
});
