export class Cond {
    #promise: Promise<void>;
    #resolve!: (value: void | PromiseLike<void>) => void;

    constructor() {
        this.#promise = new Promise<void>((resolve) => {
            this.#resolve = resolve;
        });
    }

    broadcast(): void {
        this.#resolve();

        this.#promise = new Promise<void>((resolve) => {
            this.#resolve = resolve;
        });
    }

    wait(): Promise<void> {
        return this.#promise;
    }
}
