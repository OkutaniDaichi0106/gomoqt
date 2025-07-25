import { SubscribeMessage } from './subscribe';
import { Writer, Reader } from '../io';

describe('SubscribeMessage', () => {
  it('should encode and decode', async () => {
    const subscribeId = 123n;
    const broadcastPath = 'path';
    const trackName = 'track';
    const trackPriority = 1n;
    const minGroupSequence = 2n;
    const maxGroupSequence = 3n;

    // Create a TransformStream to connect readable and writable streams
    const { readable, writable } = new TransformStream<Uint8Array, Uint8Array>();
    const writer = new Writer(writable);
    const reader = new Reader(readable);

    const [encodedMessage, encodeErr] = await SubscribeMessage.encode(
      writer,
      subscribeId,
      broadcastPath,
      trackName,
      trackPriority,
      minGroupSequence,
      maxGroupSequence
    );
    expect(encodeErr).toBeUndefined();
    expect(encodedMessage).toBeDefined();
    expect(encodedMessage?.subscribeId).toEqual(subscribeId);
    expect(encodedMessage?.broadcastPath).toEqual(broadcastPath);
    expect(encodedMessage?.trackName).toEqual(trackName);
    expect(encodedMessage?.trackPriority).toEqual(trackPriority);
    expect(encodedMessage?.minGroupSequence).toEqual(minGroupSequence);
    expect(encodedMessage?.maxGroupSequence).toEqual(maxGroupSequence);

    const [decodedMessage, decodeErr] = await SubscribeMessage.decode(reader);
    expect(decodeErr).toBeUndefined();
    expect(decodedMessage).toBeDefined();
    expect(decodedMessage?.subscribeId).toEqual(subscribeId);
    expect(decodedMessage?.broadcastPath).toEqual(broadcastPath);
    expect(decodedMessage?.trackName).toEqual(trackName);
    expect(decodedMessage?.trackPriority).toEqual(trackPriority);
    expect(decodedMessage?.minGroupSequence).toEqual(minGroupSequence);
    expect(decodedMessage?.maxGroupSequence).toEqual(maxGroupSequence);
  });
});
