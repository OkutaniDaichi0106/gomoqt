export class Mutex {
	#lockChain: Promise<void>;

	constructor() {
		this.#lockChain = Promise.resolve();
	}

	async lock(): Promise<UnlockFunction> {
		const previousLock = this.#lockChain;
		let releaseFunction: UnlockFunction;

		// Create next link in the chain
		this.#lockChain = new Promise<void>((resolve) => {
			releaseFunction = resolve;
		});

		await previousLock;
		return releaseFunction!;
	}
}

type UnlockFunction = () => void;