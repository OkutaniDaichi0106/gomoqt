import { Mutex } from "./mutex";

export class Queue<T> {
	#items: T[] = [];
	#pending?: [() => void, Promise<void>];
	#mutex: Mutex = new Mutex();

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

	async dequeue(): Promise<[T?, Error?]> {
		const unlock = await this.#mutex.lock();
		let item: T | undefined;
		while (true) {
			// If we have items, return the first one
			if (this.#items.length > 0) {
				item = this.#items.shift();
				break;
			}
			
			// Wait for next item or close
			let resolve: () => void;
			const chan = new Promise<void>((res) => {
				resolve = res;
			});
			this.#pending = [resolve!, chan];
			
			await chan;
			// Loop will check for items or close error again
		}
		unlock();
		return [item, undefined];
	}
}