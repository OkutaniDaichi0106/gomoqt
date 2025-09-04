import { Writer, Reader } from "../io";

// export class FrameMessage {
//     data: Uint8Array;

//     constructor(data?: Uint8Array) {
//         this.data = data ?? new Uint8Array(0);
//     }

//     length(): number {
//         return this.data.length;
//     }

//     async encode(writer: Writer): Promise<Error | undefined> {
//         writer.writeUint8Array(this.data);
//         return await writer.flush();
//     }

//     async decode(reader: Reader): Promise<Error | undefined> {
//         let arr: Uint8Array<ArrayBufferLike> | undefined;
//         let err: Error | undefined;
//         [arr, err] = await reader.readUint8Array();
//         if (err) {
//             return err;
//         }

//         this.data = arr!;

//         return undefined;
//     }

//     get byteLength(): number {
//         return this.data.byteLength;
//     }

//     copyTo(dest: AllowSharedBufferSource): void {
//         if (dest instanceof Uint8Array) {
//             dest.set(this.data);
//         } else if (dest instanceof ArrayBuffer) {
//             new Uint8Array(dest).set(this.data);
//         } else {
//             throw new Error("Unsupported destination type");
//         }
//     }

//     clone(buffer?: Uint8Array<ArrayBufferLike>): Frame {
//         if (buffer && buffer.byteLength >= this.data.byteLength) {
//             buffer.set(this.data);
//             return new Frame(buffer.subarray(0, this.data.byteLength));
//         }
//         return new Frame(this.data.slice());
//     }

//     copyFrom(src: Source): void {
//         if (src.byteLength > this.data.byteLength) {
//             this.data = new Uint8Array(src.byteLength);
//         }
//         src.copyTo(this.data);
//     }
// }