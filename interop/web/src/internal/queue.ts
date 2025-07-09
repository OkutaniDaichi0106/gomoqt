export class Queue<T> {
	#items: T[] = [];
	#pending?: [() => void, Promise<void>];
	#closeError?: Error;

	enqueue(item: T): void {
		this.#items.push(item);
		if (this.#pending) {
			const [resolve] = this.#pending;
			this.#pending = undefined;
			resolve();
		}
	}

	dequeue(): [T?, Error?] {
		while (true) {
			if (this.#closeError) {
				return [undefined, this.#closeError];
			}
			if (this.#items.length > 0) {
				return [this.#items.shift()!, undefined];
			} else {
				let resolve: () => void;
				const chan = new Promise<void>((res) => {
					resolve = res;
				});
				this.#pending = [resolve!, chan];
			}

			if (this.#pending) {
				const [, chan] = this.#pending;
				this.#pending = undefined;
				chan.then(() => {
					return this.#items.shift()!;
				});
			}
		}
	}

	isEmpty(): boolean {
		return this.#items.length === 0 && this.#pending === undefined;
	}

	close(error?: Error): void {
		if (this.#closeError) {
			return; // Already closed
		}
		this.#closeError = error || new Error("Queue closed");
		if (this.#pending) {
			const [resolve] = this.#pending;
			this.#pending = undefined;
			resolve();
		}
	}
}