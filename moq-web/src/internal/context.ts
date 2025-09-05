export interface Context {
    done(): Promise<void>;
    err(): Error | undefined;
}

export type CancelFunc = () => void;
export type CancelCauseFunc = (err: Error | undefined) => void;

class DefaultContext implements Context {
    #donePromise: Promise<void>;
    #resolve!: () => void;
    #reject!: (reason?: any) => void;
    #err?: Error;

    constructor(parent?: Context) {
        this.#donePromise = new Promise((resolve, reject) => {
            this.#resolve = resolve;
            this.#reject = reject;
        });

        // Set up parent cancellation propagation using Promise
        if (parent) {
            parent.done().then(() => {
                this.#err = parent.err();
                this.#resolve();
            }).catch(() => {
                this.#resolve();
            });
        }
    }

    done(): Promise<void> {
        return this.#donePromise;
    }

    err(): Error | undefined {
        return this.#err;
    }

    cancel(err?: Error): void {
        this.#err = err;
        this.#resolve();
    }
}

const backgroundContext: Context = function(): Context {
    // Create context with its own controller for background lifecycle events
    const context = new DefaultContext();
    
    // Unified handler for better performance
    const handlePageTermination = (eventType: string) => {
        context.cancel(new Error(`Page ${eventType}`));
    };
    
    // Use modern event handling with passive listeners where appropriate
    const options = { once: true, passive: true };
    window.addEventListener('beforeunload', () => handlePageTermination('unloading'), options);
    window.addEventListener('unload', () => handlePageTermination('unloaded'), options);
    window.addEventListener('pagehide', () => handlePageTermination('hidden'), options);

    // Service Worker detection with better type safety
    if (typeof self !== 'undefined' && 'serviceWorker' in navigator) {
        self.addEventListener?.('activate', () => {
            context.cancel(new Error('Service Worker activated'));
        }, { once: true });
    }

    return context;
}();

// Public API functions
export function background(): Context {
    return backgroundContext;
}

export function withSignal(parent: Context, signal: AbortSignal): Context {
    // Create a new context with its own controller for independent cancellation control
    const context = new DefaultContext(parent);

    // More efficient event handling - avoid function creation in loops
    if (signal.aborted) {
        context.cancel(signal.reason || new Error('Context cancelled'));
    } else {
        signal.addEventListener('abort', () => {
            context.cancel(signal.reason || new Error('Context cancelled'));
        }, { once: true });
    }

    return context;
}

export function withCancel(parent: Context): [Context, CancelFunc] {
    const context = new DefaultContext(parent);
    return [context, () => context.cancel(new Error('Context cancelled'))];
}

export function withCancelCause(parent: Context): [Context, CancelCauseFunc] {
    const context = new DefaultContext(parent);
    return [context, (err: Error | undefined) => context.cancel(err)];
}

export function withTimeout(parent: Context, timeout: number): Context {
    const context = new DefaultContext(parent);
    
    const timeoutId = setTimeout(() => {
        context.cancel(new Error(`Context timeout after ${timeout}ms`));
    }, timeout);
    
    // Clean up timeout if parent is cancelled first - using Promise
    const cleanup = () => clearTimeout(timeoutId);
    parent.done().then(cleanup).catch(() => {});
    context.done().then(cleanup).catch(() => {});
    
    return context;
}

export function withPromise<T>(parent: Context, promise: Promise<T>): Context {
    const context = new DefaultContext(parent);

    // Use more specific error handling for TypeScript
    promise.then(
        (value) => context.cancel(new Error(`Promise resolved with: ${String(value)}`)),
        (reason) => {
            const error = reason instanceof Error 
                ? reason 
                : new Error(`Promise rejected: ${String(reason)}`);
            context.cancel(error);
        }
    ).catch(
        // Catch any unexpected errors to prevent unhandled promise rejections
        (err) => context.cancel(new Error(`Unexpected error in promise: ${String(err)}`))
    );
    
    return context;
}