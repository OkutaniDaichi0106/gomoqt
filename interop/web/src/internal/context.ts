export interface Context {
    readonly signal: AbortSignal;
    done(): Promise<void>;
    err(): Error | undefined;
}

export type CancelFunc = () => void;
export type CancelCauseFunc = (err: Error | undefined) => void;

class DefaultContext implements Context {
    #signal: AbortSignal;
    #err?: Error;
    #donePromise: Promise<void>;

    constructor(signal: AbortSignal) {
        this.#signal = signal;
        
        // Set error from signal reason when signal is aborted
        if (this.#signal.aborted) {
            this.#err = this.#signal.reason || new Error('Context cancelled');
            // Initialize promise as already rejected for aborted signals
            this.#donePromise = Promise.reject(this.#err);
        } else {
            // Initialize promise with abort listener for non-aborted signals
            this.#donePromise = new Promise((resolve, reject) => {
                this.#signal.addEventListener('abort', () => {
                    this.#err = this.#signal.reason || new Error('Context cancelled');
                    reject(this.#err);
                }, { once: true });
            });
        }
    }

    get signal(): AbortSignal {
        return this.#signal;
    }

    done(): Promise<void> {
        // Promise is always initialized in constructor - just return it
        return this.#donePromise;
    }

    err(): Error | undefined {
        return this.#err;
    }
}

function createChildContext(parent: Context, controller: AbortController): Context {
    if (parent.signal.aborted) {
        controller.abort(parent.err());
    } else {
        parent.signal.addEventListener('abort', () => {
            controller.abort(parent.err());
        }, { once: true }); // Automatic cleanup for derived contexts
    }
    return new DefaultContext(controller.signal);
}

let backgroundContext: Context | undefined = undefined;

function createBackgroundContext(): Context {
    const controller = new AbortController();
    
    // Unified handler for better performance
    const handlePageTermination = (eventType: string) => {
        controller.abort(new Error(`Page ${eventType}`));
    };
    
    // Use modern event handling with passive listeners where appropriate
    const options = { once: true, passive: true };
    window.addEventListener('beforeunload', () => handlePageTermination('unloading'), options);
    window.addEventListener('unload', () => handlePageTermination('unloaded'), options);
    window.addEventListener('pagehide', () => handlePageTermination('hidden'), options);

    // Service Worker detection with better type safety
    if (typeof self !== 'undefined' && 'serviceWorker' in navigator) {
        self.addEventListener?.('activate', () => {
            controller.abort(new Error('Service Worker activated'));
        }, { once: true });
    }

    return new DefaultContext(controller.signal);
}

// Public API functions
export function background(): Context {
    if (!backgroundContext) {
        backgroundContext = createBackgroundContext();
    }
    return backgroundContext;
}

export function withSignal(parent: Context, signal: AbortSignal): Context {
    const controller = new AbortController();
    
    // More efficient event handling - avoid function creation in loops
    if (parent.signal.aborted) {
        controller.abort(parent.err());
    } else {
        parent.signal.addEventListener('abort', () => {
            controller.abort(parent.err());
        }, { once: true });
    }
    
    if (signal.aborted) {
        controller.abort(signal.reason || new Error('Context cancelled'));
    } else {
        signal.addEventListener('abort', () => {
            controller.abort(signal.reason || new Error('Context cancelled'));
        }, { once: true });
    }

    return new DefaultContext(controller.signal);
}

export function withCancel(parent: Context): [Context, CancelFunc] {
    const controller = new AbortController();
    const context = createChildContext(parent, controller);
    return [context, () => controller.abort(new Error('Context cancelled'))];
}

export function withCancelCause(parent: Context): [Context, CancelCauseFunc] {
    const controller = new AbortController();
    const context = createChildContext(parent, controller);
    return [context, (err: Error | undefined) => controller.abort(err)];
}

export function withTimeout(parent: Context, timeout: number): Context {
    const controller = new AbortController();
    const context = createChildContext(parent, controller);
    
    const timeoutId = setTimeout(() => {
        controller.abort(new Error(`Context timeout after ${timeout}ms`));
    }, timeout);
    
    // Clean up timeout if parent is cancelled first - TypeScript optimization
    const cleanup = () => clearTimeout(timeoutId);
    parent.signal.addEventListener('abort', cleanup, { once: true });
    controller.signal.addEventListener('abort', cleanup, { once: true });
    
    return context;
}

export function withPromise<T>(parent: Context, promise: Promise<T>): Context {
    const controller = new AbortController();
    const context = createChildContext(parent, controller);
    
    // Use more specific error handling for TypeScript
    promise.then(
        (value) => controller.abort(new Error(`Promise resolved with: ${String(value)}`)),
        (reason) => {
            const error = reason instanceof Error 
                ? reason 
                : new Error(`Promise rejected: ${String(reason)}`);
            controller.abort(error);
        }
    );
    
    return context;
}