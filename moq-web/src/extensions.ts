export class Extensions {
	entries: Map<number, Uint8Array>;

	constructor() {
		this.entries = new Map<number, Uint8Array>();
	}

	has(id: number): boolean {
		return this.entries.has(id);
	}

	delete(id: number): boolean {
		return this.entries.delete(id);
	}

	addBytes(id: number, bytes: Uint8Array): void {
		this.entries.set(id, bytes);
	}

	getBytes(id: number): Uint8Array | undefined {
		return this.entries.get(id);
	}

	addString(id: number, str: string): void {
		const encoder = new TextEncoder();
		this.entries.set(id, encoder.encode(str));
	}

	getString(id: number): string | undefined {
		const bytes = this.entries.get(id);
		if (bytes) {
			const decoder = new TextDecoder();
			return decoder.decode(bytes);
		}
		return undefined;
	}

	addNumber(id: number, num: bigint): void {
		const buffer = new ArrayBuffer(8);
		const view = new DataView(buffer);
		view.setBigUint64(0, BigInt(num), true); // true for little-endian
		this.entries.set(id, new Uint8Array(buffer));
	}

	getNumber(id: number): bigint | undefined {
		const bytes = this.entries.get(id);
		if (bytes && bytes.length === 8) {
			const view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength);
			return view.getBigUint64(0, true); // true for little-endian
		}
		return undefined;
	}

	addBoolean(id: number, value: boolean): void {
		const byte = new Uint8Array([value ? 1 : 0]);
		this.entries.set(id, byte);
	}

	getBoolean(id: number): boolean | undefined {
		const bytes = this.entries.get(id);
		if (bytes && bytes.length === 1) {
			return bytes[0] === 1;
		}
		return undefined;
	}
}
