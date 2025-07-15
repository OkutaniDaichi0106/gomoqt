export class Cond {
    #waiters: Array<(value: void | PromiseLike<void>) => void> = [];

    constructor() {}

    broadcast(): void {
        // Resolve all current waiters
        const waiters = this.#waiters.splice(0);
        for (const resolve of waiters) {
            resolve();
        }
    }

    wait(): Promise<void> {
        return new Promise<void>((resolve) => {
            this.#waiters.push(resolve);
        });
    }
}
