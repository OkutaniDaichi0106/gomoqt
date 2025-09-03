interface BytesPoolOptions {
    maxPerBucket?: number; // Maximum arrays per bucket
    maxTotalBytes?: number; // Optional total byte limit
}

// Global default configuration (similar to Go's package-level variables)
export let DefaultBytesPoolOptions: Required<BytesPoolOptions> = {
    maxPerBucket: 5,
    maxTotalBytes: 0, // 0 means no limit
};

export class BufferPool {
    #min: number;
    #middle: number;
    #max: number;
    #buckets: Array<ArrayBufferLike[]>; // 0:min, 1:middle, 2:max
    #maxPerBucket: number;
    #maxTotalBytes: number;
    #currentBytes: number;

    constructor(min: number, middle: number, max: number, options: BytesPoolOptions = DefaultBytesPoolOptions) {
        if (!(min > 0 && middle > 0 && max > 0)) {
            throw new Error("min, middle, max must be greater than 0");
        }
        if (!(min < middle && middle < max)) {
            throw new Error("min, middle, max must be in ascending order");
        }
        this.#min = min;
        this.#middle = middle;
        this.#max = max;
        this.#buckets = [[], [], []];
        this.#maxPerBucket = options.maxPerBucket ?? DefaultBytesPoolOptions.maxPerBucket;
        this.#maxTotalBytes = options.maxTotalBytes ?? DefaultBytesPoolOptions.maxTotalBytes;
        this.#currentBytes = 0;
    }

    acquire(capacity: number): ArrayBufferLike {
        let idx: number;
        let size: number;
        if (capacity <= this.#min) {
            idx = 0; size = this.#min;
        } else if (capacity <= this.#middle) {
            idx = 1; size = this.#middle;
        } else if (capacity <= this.#max) {
            idx = 2; size = this.#max;
        } else {
            return new ArrayBuffer(capacity);
        }
        const bucket = this.#buckets[idx];
        if (bucket.length > 0) {
            this.#currentBytes -= size;
            return bucket.pop()!;
        }
        return new ArrayBuffer(size);
    }

    release(bytes: ArrayBufferLike): void {
        const size = bytes.byteLength;
        let idx = 0;
        if (size === this.#min) {
            // No action needed, already idx = 0
        }else if (size === this.#middle){ 
            idx = 1;
        }else if (size === this.#max) {
            idx = 2;
        }else return; // Non-matching sizes are GC'd
        if (this.#maxTotalBytes && this.#currentBytes + size > this.#maxTotalBytes) return;
        const bucket = this.#buckets[idx];
        if (bucket.length >= this.#maxPerBucket) {
            bucket.shift();
            this.#currentBytes -= size;
        }
        bucket.push(bytes);
        this.#currentBytes += size;
    }

    // Explicit cleanup (optional)
    cleanup(): void {
        for (let i = 0; i < this.#buckets.length; i++) {
            this.#buckets[i].length = 0;
        }

        return;
    }
}

// Global default pool instance
export const DefaultBufferPool = new BufferPool(64, 256, 1024);