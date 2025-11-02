import type { Source } from "./internal/webtransport/mod.ts";

export interface Frame extends Source {
	data: Uint8Array;
}

export class BytesFrame implements Frame {
	bytes: Uint8Array;

	constructor(bytes: Uint8Array) {
		this.bytes = bytes;
	}

	get data(): Uint8Array {
		return this.bytes;
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

	clone(buffer?: Uint8Array<ArrayBufferLike>): BytesFrame {
		if (buffer && buffer.byteLength >= this.bytes.byteLength) {
			buffer.set(this.bytes);
			return new BytesFrame(buffer.subarray(0, this.bytes.byteLength));
		}
		return new BytesFrame(this.bytes.slice());
	}

	copyFrom(src: Frame): void {
		if (src.byteLength > this.bytes.byteLength) {
			this.bytes = new Uint8Array(src.byteLength);
		}
		src.copyTo(this.bytes);
	}
}
