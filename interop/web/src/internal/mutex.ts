export class Mutex {
	#lockChain: Promise<void> = Promise.resolve();

	async lock(): Promise<UnlockFunction> {
		const previousLock = this.#lockChain;
		let releaseFunction: UnlockFunction;

		// Create next link in the chain with microtask scheduling
		this.#lockChain = new Promise<void>((resolve) => {
			releaseFunction = () => queueMicrotask(resolve);
		});

		await previousLock;
		return releaseFunction!;
	}

	get isLocked(): boolean {
		return this.#lockChain !== Promise.resolve();
	}
}

type UnlockFunction = () => void;

// // For debugging and performance analysis (optional)
// export class DebugMutex {
// 	#tail: Promise<void> = Promise.resolve();
// 	#lockCount = 0;
// 	#maxConcurrency = 0;
// 	#totalLocks = 0;

// 	lock(): Promise<() => void> {
// 		const waitFor = this.#tail;
// 		const lockId = ++this.#totalLocks;
// 		this.#lockCount++;
// 		this.#maxConcurrency = Math.max(this.#maxConcurrency, this.#lockCount);
		
// 		let unlock: () => void;

// 		this.#tail = new Promise<void>((resolve) => {
// 			unlock = () => {
// 				this.#lockCount--;
// 				queueMicrotask(resolve);
// 			};
// 		});

// 		return waitFor.then(() => unlock!);
// 	}

// 	get stats() {
// 		return {
// 			currentLocks: this.#lockCount,
// 			maxConcurrency: this.#maxConcurrency,
// 			totalLocks: this.#totalLocks,
// 			isLocked: this.#lockCount > 0
// 		};
// 	}
// }

// Usage:
// const mutex = new Mutex();
// const release = await mutex.lock();
// try {
//   // Critical section
// } finally {
//   release();
// }