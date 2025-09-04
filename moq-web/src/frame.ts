import { Source } from "./io"

export class Frame implements Source {
    bytes: Uint8Array;

    constructor(bytes: Uint8Array) {
        this.bytes = bytes;
    }

    get byteLength(): number {
        return this.bytes.byteLength;
    }

    copyTo(dest: AllowSharedBufferSource): void {
        if (dest instanceof Uint8Array) {
            dest.set(this.bytes);
        } else if (dest instanceof ArrayBuffer) {
            new Uint8Array(dest).set(this.bytes);
        } else {
            throw new Error("Unsupported destination type");
        }
    }

    clone(buffer?: Uint8Array<ArrayBufferLike>): Frame {
        if (buffer && buffer.byteLength >= this.bytes.byteLength) {
            buffer.set(this.bytes);
            return new Frame(buffer.subarray(0, this.bytes.byteLength));
        }
        return new Frame(this.bytes.slice());
    }

    copyFrom(src: Source): void {
        if (src.byteLength > this.bytes.byteLength) {
            this.bytes = new Uint8Array(src.byteLength);
        }
        src.copyTo(this.bytes);
    }
}