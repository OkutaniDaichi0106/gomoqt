import type { Source } from "./internal/webtransport/mod.ts";

export interface Frame extends Source {
	data: Uint8Array;
}

export class BytesFrame implements Frame {
	data: Uint8Array;

	constructor(bytes: Uint8Array) {
		this.data = bytes;
	}

	get bytes(): Uint8Array {
		return this.data;
	}

	get byteLength(): number {
		return this.data.byteLength;
	}

	copyTo(dest: AllowSharedBufferSource): void {
		if (dest instanceof Uint8Array) {
			dest.set(this.data);
		} else if (dest instanceof ArrayBuffer) {
			new Uint8Array(dest).set(this.data);
		} else {
			throw new Error("Unsupported destination type");
		}
	}

	clone(buffer?: Uint8Array<ArrayBufferLike>): BytesFrame {
		if (buffer && buffer.byteLength >= this.data.byteLength) {
			buffer.set(this.data);
			return new BytesFrame(buffer.subarray(0, this.data.byteLength));
		}
		return new BytesFrame(this.data.slice());
	}

	copyFrom(src: Frame): void {
		if (src.byteLength > this.data.byteLength) {
			this.data = new Uint8Array(src.byteLength);
		}
		src.copyTo(this.data);
	}
}

export const Frame: {
	new (bytes: Uint8Array): Frame;
} = BytesFrame;
