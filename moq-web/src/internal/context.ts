export interface Context {
    /**
     * Returns a promise that resolves when the context is finished (cancelled, timed out, etc).
     * Never rejects.
     */
    done(): Promise<void>;
    /**
     * The cancellation error if the context is finished. `undefined` means: not cancelled / finished without error.
     */
    err(): Error | undefined;
}

export type CancelFunc = () => void;
export type CancelCauseFunc = (err: Error | undefined) => void;

export const ContextCancelledError = new Error("Context cancelled");
export const ContextTimeoutError = new Error("Context timeout");

class DefaultContext implements Context {
    #donePromise: Promise<void>;
    #resolve!: () => void;
    #err?: Error; // undefined => either not done yet OR completed cleanly

    constructor(parent?: Context) {
        this.#donePromise = new Promise((resolve) => {
            this.#resolve = resolve;
        });

        if (parent) {
            // Propagate parent cancellation (simplified for idempotency)
            parent.done().then(() => {
                this.cancel(parent.err()); // cancel is idempotent, so no need for #finished check
            });
        }
    }

    done(): Promise<void> {
        return this.#donePromise;
    }

    err(): Error | undefined {
        return this.#err;
    }

    cancel(err: Error = ContextCancelledError): void {
        if (this.#err !== undefined) return; // idempotent
        this.#err = err; // explicitly set, even if undefined
        this.#resolve();
    }
}

const backgroundContext: Context = (() => {
    const context = new DefaultContext();
    // Guard for SSR / non-browser environments
    if (typeof window !== 'undefined' && typeof window.addEventListener === 'function') {
        const handlePageTermination = (eventType: string) => {
            context.cancel(new Error(`Page ${eventType}`));
        };
        const once = { once: true } as const;
        window.addEventListener('beforeunload', () => handlePageTermination('unloading'), once);
        window.addEventListener('unload', () => handlePageTermination('unloaded'), once);
        window.addEventListener('pagehide', () => handlePageTermination('hidden'), once);
    }
    return context;
})();

// Public API functions
export function background(): Context {
    return backgroundContext;
}

export function withSignal(parent: Context, signal: AbortSignal): Context {
    const context = new DefaultContext(parent);

    const done = () => context.done().then(() => signal.removeEventListener('abort', onAbort));

    const onAbort = () => {
        context.cancel(
            (signal as any).reason instanceof Error
                ? (signal as any).reason as Error
                : ContextCancelledError
        );
    };

    if (signal.aborted) {
        onAbort();
    } else {
        signal.addEventListener('abort', onAbort, { once: true });
        // Ensure listener removal when context ends earlier
        done();
    }
    return context;
}

export function withCancel(parent: Context): [Context, CancelFunc] {
    const context = new DefaultContext(parent);
    return [context, () => context.cancel(ContextCancelledError)];
}

export function withCancelCause(parent: Context): [Context, CancelCauseFunc] {
    const context = new DefaultContext(parent);
    return [context, (err: Error | undefined) => context.cancel(err)];
}

export function withTimeout(parent: Context, timeoutMs: number): Context {
    const context = new DefaultContext(parent);
    const id = setTimeout(() => {
        context.cancel(new Error(`Context timeout after ${timeoutMs}ms`));
    }, timeoutMs);
    context.done().finally(() => clearTimeout(id));
    return context;
}

export function withPromise<T>(parent: Context, promise: Promise<T>): Context {
    const context = new DefaultContext(parent);
    promise.then(
        () => context.cancel(), // normal completion, no error cause
        (reason) => {
            const error = reason instanceof Error ? reason : new Error(String(reason));
            context.cancel(error);
        }
    );
    return context;
}

/**
 * Convert a Context into an AbortSignal for integration with fetch / Web APIs.
 */
export function toAbortSignal(ctx: Context): AbortSignal {
    const ac = new AbortController();
    ctx.done().then(() => {
        const err = ctx.err();
        if (err) {
            // AbortController#abort can take a reason in modern browsers
            try {
                (ac as any).abort(err);
            } catch {
                ac.abort();
            }
        } else {
            ac.abort();
        }
    });
    return ac.signal;
}