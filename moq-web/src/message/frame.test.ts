// import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
// import { FrameMessage } from './frame';
// import { Writer, Reader } from '../io';

// describe('FrameMessage', () => {
//   it('should encode and decode', async () => {
//     const data = new Uint8Array([1, 2, 3, 4, 5]);

//     // Create a buffer to store the encoded data
//     const chunks: Uint8Array[] = [];
//     const writableStream = new WritableStream<Uint8Array>({
//       write(chunk) {
//         chunks.push(chunk);
//       }
//     });

//     // Encode the message
//     const writer = new Writer(writableStream);
//     const [encodedMessage, encodeErr] = await FrameMessage.encode(writer, data);
//     expect(encodeErr).toBeUndefined();
//     expect(encodedMessage).toBeDefined();
//     expect(encodedMessage?.data).toEqual(data);

//     // Create a readable stream from the encoded data
//     const combinedData = new Uint8Array(chunks.reduce((total, chunk) => total + chunk.length, 0));
//     let offset = 0;
//     for (const chunk of chunks) {
//       combinedData.set(chunk, offset);
//       offset += chunk.length;
//     }

//     const readableStream = new ReadableStream<Uint8Array>({
//       start(controller) {
//         controller.enqueue(combinedData);
//         controller.close();
//       }
//     });

//     // Decode the message
//     const reader = new Reader(readableStream);
//     const [decodedMessage, decodeErr] = await FrameMessage.decode(reader);
//     expect(decodeErr).toBeUndefined();
//     expect(decodedMessage).toBeDefined();
//     expect(decodedMessage?.data).toEqual(data);
//   }, 10000); // Increase timeout to 10 seconds
// });
