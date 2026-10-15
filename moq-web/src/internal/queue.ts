import { Mutex } from "@okdaichi/golikejs/sync";

export class Queue<T> {
	#items: T[] = [];
	#pending?: [() => void, Promise<void>];
	#mutex: Mutex = new Mutex();
	#closed: boolean = false;

	async enqueue(item: T): Promise<void> {
		await this.#mutex.lock();
		try {
			this.#items.push(item);

			if (this.#pending) {
				const [resolve] = this.#pending;
				this.#pending = undefined;
				resolve();
			}
		} finally {
			this.#mutex.unlock();
		}
	}

	async tryEnqueue(item: T): Promise<boolean> {
		await this.#mutex.lock();
		try {
			if (this.#closed) return false;
			this.#items.push(item);
			if (this.#pending) {
				const [resolve] = this.#pending;
				this.#pending = undefined;
				resolve();
			}
			return true;
		} finally {
			this.#mutex.unlock();
		}
	}

	/**
	 * Takes the first item, waiting for one. Returns undefined once the queue
	 * is closed and empty, or when `stop` settles first; a stopped wait takes
	 * no item, so none is lost to it.
	 */
	async dequeue(stop?: Promise<unknown>): Promise<T | undefined> {
		while (true) {
			await this.#mutex.lock();

			try {
				// If we have items, return the first one
				if (this.#items.length > 0) {
					const item = this.#items.shift();
					this.#mutex.unlock();

					// item is guaranteed to be defined here
					if (!item) {
						continue;
					}
					return item;
				}

				if (this.#closed) {
					this.#mutex.unlock();
					return undefined;
				}

				// No items available - set up waiting if not already waiting
				if (!this.#pending) {
					let resolve: () => void;
					const chan = new Promise<void>((res) => {
						resolve = res;
					});
					this.#pending = [resolve!, chan];
				}

				const [, chan] = this.#pending;

				// Release lock before waiting
				this.#mutex.unlock();

				if (stop) {
					const stopped = await Promise.race([
						chan.then(() => false),
						stop.then(() => true, () => true),
					]);
					if (stopped) return undefined;
				} else {
					await chan;
				}
			} catch (e) {
				this.#mutex.unlock();
				throw e;
			}
		}
	}

	async drain(): Promise<T[]> {
		await this.#mutex.lock();
		try {
			return this.#items.splice(0);
		} finally {
			this.#mutex.unlock();
		}
	}

	close(): void {
		if (this.#closed) {
			return;
		}
		this.#closed = true;

		this.#mutex.lock().then(() => {
			try {
				if (this.#pending) {
					const [resolve] = this.#pending;
					this.#pending = undefined;
					resolve();
				}
			} finally {
				this.#mutex.unlock();
			}
		});
	}

	get closed(): boolean {
		return this.#closed;
	}
}
