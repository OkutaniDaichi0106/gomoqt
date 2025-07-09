interface BytesPoolOptions {
    maxPerBucket?: number; // Maximum arrays per bucket
    maxTotalBytes?: number; // Optional total byte limit
}

// Global default configuration (similar to Go's package-level variables)
export let DefaultBytesPoolOptions: Required<BytesPoolOptions> = {
    maxPerBucket: 5,
    maxTotalBytes: 0, // 0 means no limit
};

export class BytesPool {
    #pool: Map<number, WeakRef<ArrayBufferLike>[]>; // Size-based buckets with WeakRef management
    #cleanupRegistry: FinalizationRegistry<{ size: number; index: number }>;
    #maxPerBucket: number; // Maximum arrays per bucket
    #maxTotalBytes: number; // Optional total byte limit
    #currentBytes: number; // Current total bytes in pool

    constructor(options: BytesPoolOptions = DefaultBytesPoolOptions) {
        this.#pool = new Map();
        this.#maxPerBucket = options.maxPerBucket ?? DefaultBytesPoolOptions.maxPerBucket;
        this.#maxTotalBytes = options.maxTotalBytes ?? DefaultBytesPoolOptions.maxTotalBytes;
        this.#currentBytes = 0;

        // Automatically remove from pool when GC'd
        this.#cleanupRegistry = new FinalizationRegistry((heldValue) => {
            const { size, index } = heldValue;
            const refs = this.#pool.get(size);
            if (refs && refs[index]) {
                // Update byte count when automatically cleaned up
                this.#currentBytes -= size;
                refs.splice(index, 1);
                if (refs.length === 0) {
                    this.#pool.delete(size);
                }
            }
        });
    }

    acquire(capacity: number): ArrayBufferLike {
        // Find appropriate size bucket
        for (const [size, refs] of this.#pool.entries()) {
            if (size >= capacity) {
                // Look for a living WeakRef
                for (let i = refs.length - 1; i >= 0; i--) {
                    const bytes = refs[i].deref();
                    if (bytes) {
                        refs.splice(i, 1);
                        if (refs.length === 0) {
                            this.#pool.delete(size);
                        }
                        // Update current byte count when retrieving from pool
                        this.#currentBytes -= size;
                        return bytes;
                    } else {
                        // Already GC'd, remove the dead reference
                        refs.splice(i, 1);
                    }
                }
                if (refs.length === 0) {
                    this.#pool.delete(size);
                }
            }
        }

        // No suitable size in pool, create new
        return new ArrayBuffer(capacity);
    }

    release(bytes: ArrayBufferLike): void {
        const size = bytes.byteLength;

        // Check total byte limit if specified
        if (this.#maxTotalBytes && this.#currentBytes + size > this.#maxTotalBytes) {
            // Don't add to pool if it would exceed byte limit
            return;
        }

        // Check bucket size limit
        const refs = this.#pool.get(size) || [];
        if (refs.length >= this.#maxPerBucket) {
            // If bucket is full, evict oldest entry from this bucket
            const evicted = refs.shift();
            if (evicted) {
                this.#currentBytes -= size;
            }
        }

        // Add to pool with WeakRef
        const index = refs.length;
        const weakRef = new WeakRef(bytes);
        refs.push(weakRef);
        this.#pool.set(size, refs);
        this.#currentBytes += size;

        // Register for automatic cleanup when GC'd
        this.#cleanupRegistry.register(bytes, { size, index });
    }

    // Explicit cleanup (optional)
    cleanup(): number {
        let removedCount = 0;

        for (const [size, refs] of this.#pool.entries()) {
            for (let i = refs.length - 1; i >= 0; i--) {
                if (!refs[i].deref()) {
                    refs.splice(i, 1);
                    removedCount++;
                }
            }
            if (refs.length === 0) {
                this.#pool.delete(size);
            }
        }

        return removedCount;
    }
}

// Global default pool instance (similar to Go's package-level variables)
export const DefaultBytesPool = new BytesPool();