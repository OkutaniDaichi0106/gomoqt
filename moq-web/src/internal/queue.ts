import { Mutex } from "./mutex";

export class Queue<T> {
	#items: T[] = [];
	#pending?: [() => void, Promise<void>];
	#mutex: Mutex = new Mutex();
	#closed: boolean = false;

	async enqueue(item: T): Promise<void> {
		const unlock = await this.#mutex.lock();
		this.#items.push(item);
		
		if (this.#pending) {
			const [resolve] = this.#pending;
			this.#pending = undefined;
			resolve();
		}

		unlock();
	}

	async dequeue(): Promise<T> {
		while (true) {
			const unlock = await this.#mutex.lock();

				// If we have items, return the first one
				if (this.#items.length > 0) {
					const item = this.#items.shift();
					unlock();

					// item is guaranteed to be defined here
					if (!item) {
						continue;
					}
					return item;
				}

				if (this.#closed) {
					unlock();
					throw new Error("Queue is closed and empty");
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
				unlock();
				
				await chan;
		}
	}

	close(): void {
		if (this.#closed) {
			return;
		}
		this.#closed = true;

		this.#mutex.lock().then((unlock) => {
			if (this.#pending) {
				const [resolve] = this.#pending;
				this.#pending = undefined;
				resolve();
			}

			unlock();
		});
	}

	get closed(): boolean {
		return this.#closed;
	}
}