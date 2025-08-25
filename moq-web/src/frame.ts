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

    copyFrom(src: Source): void {
        src.copyTo(this.bytes);
    }
}